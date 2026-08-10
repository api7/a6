package skills

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func locateRepoRoot() (string, error) {
	var dir string
	var err error
	dir, err = os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		var statErr error
		_, statErr = os.Stat(filepath.Join(dir, "go.mod"))
		if statErr == nil {
			return dir, nil
		}
		var parent string = filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

func buildA6Binary(t *testing.T, root string) string {
	t.Helper()
	var binary string = filepath.Join(t.TempDir(), "a6")
	var cmd *exec.Cmd = exec.Command("go", "build", "-o", binary, "./cmd/a6")
	cmd.Dir = root
	var output []byte
	var err error
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build a6: %v\n%s", err, output)
	}
	return binary
}

func TestSkillCommandsUseSupportedA6CommandsAndFlags(t *testing.T) {
	var shellFencePattern *regexp.Regexp = regexp.MustCompile("(?s)```(?:bash|sh|shell)\\s*\\n(.*?)```")
	var yamlFencePattern *regexp.Regexp = regexp.MustCompile("(?s)```(?:yaml|yml)\\s*\\n(.*?)```")
	var longFlagPattern *regexp.Regexp = regexp.MustCompile(`--[a-z][a-z0-9-]*`)
	var invocationPattern *regexp.Regexp = regexp.MustCompile(`(?:^|[^A-Za-z0-9_-])(a6)(?:\s|$)`)
	var valueFlags map[string]bool = a6GlobalValueFlags()
	var root string
	var err error
	root, err = locateRepoRoot()
	if err != nil {
		t.Fatalf("failed to locate repository root: %v", err)
	}
	var binary string = buildA6Binary(t, root)
	var matches []string
	matches, err = filepath.Glob(filepath.Join(root, "skills", "*", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Fatal("expected at least one skill file")
	}
	var rootHelp string
	rootHelp, err = commandHelp(binary, nil)
	if err != nil {
		t.Fatal(err)
	}
	var rootCommands map[string]bool = availableCommands(rootHelp)
	var rootFlags map[string]bool = mergeFlagSets(availableFlags(rootHelp, longFlagPattern), availableShortFlags(rootHelp))

	var regressionPath []string
	var regressionHelp string
	regressionPath, regressionHelp, err = resolveCommand(binary, []string{"route", "creat"}, rootCommands, rootFlags, valueFlags)
	if err == nil {
		t.Fatalf("expected misspelled nested command to fail, got path %q and help %q", regressionPath, regressionHelp)
	}
	regressionPath, regressionHelp, err = resolveCommand(binary, []string{"route", "--server", "https://example.test", "creat"}, rootCommands, rootFlags, valueFlags)
	if err == nil {
		t.Fatalf("expected misspelled command after a global flag to fail, got path %q and help %q", regressionPath, regressionHelp)
	}
	regressionPath, _, err = resolveCommand(binary, []string{"route", "-o", "yaml", "get", "example"}, rootCommands, rootFlags, valueFlags)
	if err != nil || strings.Join(regressionPath, " ") != "route get" {
		t.Fatalf("expected supported short root flag before subcommand, got path %q and error %v", regressionPath, err)
	}
	var embedded []string = cliInvocations("CURRENT=$(a6 route get blue-green)", invocationPattern)
	if len(embedded) != 1 || !strings.HasPrefix(embedded[0], "a6 route get") {
		t.Fatalf("expected embedded a6 invocation, got %q", embedded)
	}
	var quoted []string = cliInvocations(`a6 debug trace id --header "X-Test: a|b;c&d)" --bogus`, invocationPattern)
	if len(quoted) != 1 || !strings.Contains(quoted[0], "--bogus") {
		t.Fatalf("expected quoted separators to preserve the complete invocation, got %q", quoted)
	}
	var yamlBlocks []string
	yamlBlocks, err = skillShellBlocks("```yaml\n- name: Validate\n  run: >\n    a6 route list\n    --unsupported\n```", shellFencePattern, yamlFencePattern)
	if err != nil {
		t.Fatalf("failed to extract workflow run block: %v", err)
	}
	var yamlCommands []string = joinedShellLines(yamlBlocks[0])
	if len(yamlCommands) != 1 || yamlCommands[0] != "a6 route list --unsupported" {
		t.Fatalf("expected workflow run block, got %q", yamlBlocks)
	}

	var file string
	for _, file = range matches {
		var data []byte
		data, err = os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		var declaredCommands []string
		declaredCommands, err = frontmatterA6Commands(string(data))
		if err != nil {
			t.Fatalf("%s: failed to parse frontmatter: %v", file, err)
		}
		var declaredCommand string
		for _, declaredCommand = range declaredCommands {
			var fields []string = strings.Fields(declaredCommand)
			if len(fields) < 2 || fields[0] != "a6" {
				t.Fatalf("%s: a6_commands entry %q must start with a6 and include a command", file, declaredCommand)
			}
			var path []string
			path, _, err = resolveCommand(binary, commandFields(fields[1:]), rootCommands, rootFlags, valueFlags)
			if err != nil {
				t.Fatalf("%s: a6_commands entry %q is invalid: %v", file, declaredCommand, err)
			}
			if strings.Join(path, " ") != strings.Join(fields[1:], " ") {
				t.Fatalf("%s: a6_commands entry %q must contain only a command path", file, declaredCommand)
			}
		}
		var blocks []string
		blocks, err = skillShellBlocks(string(data), shellFencePattern, yamlFencePattern)
		if err != nil {
			t.Fatalf("%s: failed to parse fenced YAML: %v", file, err)
		}
		var block string
		for _, block = range blocks {
			var lines []string = joinedShellLines(block)
			var line string
			for _, line = range lines {
				var invocation string
				for _, invocation = range cliInvocations(line, invocationPattern) {
					var fields []string = strings.Fields(invocation)
					if len(fields) < 2 {
						continue
					}
					var path []string
					var help string
					path, help, err = resolveCommand(binary, commandFields(fields[1:]), rootCommands, rootFlags, valueFlags)
					if err != nil {
						t.Fatalf("%s: %v", file, err)
					}
					var validFlags map[string]bool = mergeFlagSets(rootFlags, availableFlags(help, longFlagPattern))
					var flags []string = longFlagPattern.FindAllString(invocation, -1)
					var flag string
					for _, flag = range flags {
						if flag != "--help" && !validFlags[flag] {
							t.Fatalf("%s: command %q uses unsupported flag %q", file, "a6 "+strings.Join(path, " "), flag)
						}
					}
				}
			}
		}
	}
}

func frontmatterA6Commands(data string) ([]string, error) {
	var lines []string = strings.Split(data, "\n")
	if len(lines) < 3 || lines[0] != "---" {
		return nil, fmt.Errorf("missing opening frontmatter delimiter")
	}
	var end int = -1
	var index int
	for index = 1; index < len(lines); index++ {
		if lines[index] == "---" {
			end = index
			break
		}
	}
	if end == -1 {
		return nil, fmt.Errorf("missing closing frontmatter delimiter")
	}
	var frontmatter struct {
		Metadata struct {
			A6Commands []string `yaml:"a6_commands"`
		} `yaml:"metadata"`
	}
	var err error = yaml.Unmarshal([]byte(strings.Join(lines[1:end], "\n")), &frontmatter)
	if err != nil {
		return nil, err
	}
	return frontmatter.Metadata.A6Commands, nil
}

func skillShellBlocks(data string, shellFencePattern *regexp.Regexp, yamlFencePattern *regexp.Regexp) ([]string, error) {
	var blocks []string
	var match []string
	for _, match = range shellFencePattern.FindAllStringSubmatch(data, -1) {
		blocks = append(blocks, match[1])
	}
	for _, match = range yamlFencePattern.FindAllStringSubmatch(data, -1) {
		var runBlocks []string
		var err error
		runBlocks, err = yamlRunBlocks(match[1])
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, runBlocks...)
	}
	return blocks, nil
}

func yamlRunBlocks(block string) ([]string, error) {
	var root yaml.Node
	var err error = yaml.Unmarshal([]byte(block), &root)
	if err != nil {
		return nil, err
	}
	var runBlocks []string
	collectYAMLRunBlocks(&root, &runBlocks)
	return runBlocks, nil
}

func collectYAMLRunBlocks(node *yaml.Node, runBlocks *[]string) {
	if node.Kind == yaml.MappingNode {
		var index int
		for index = 0; index+1 < len(node.Content); index += 2 {
			var key *yaml.Node = node.Content[index]
			var value *yaml.Node = node.Content[index+1]
			if key.Value == "run" && value.Kind == yaml.ScalarNode {
				*runBlocks = append(*runBlocks, value.Value)
			}
			collectYAMLRunBlocks(value, runBlocks)
		}
		return
	}
	var child *yaml.Node
	for _, child = range node.Content {
		collectYAMLRunBlocks(child, runBlocks)
	}
}

func commandFields(fields []string) []string {
	var valueFlags map[string]bool = a6GlobalValueFlags()
	for len(fields) > 0 && strings.HasPrefix(fields[0], "-") {
		var flag string = strings.SplitN(fields[0], "=", 2)[0]
		var hasInlineValue bool = strings.Contains(fields[0], "=")
		fields = fields[1:]
		if valueFlags[flag] && !hasInlineValue && len(fields) > 0 {
			fields = fields[1:]
		}
	}
	return fields
}

func a6GlobalValueFlags() map[string]bool {
	return map[string]bool{
		"--api-key": true,
		"--context": true,
		"--output":  true,
		"--server":  true,
		"-o":        true,
	}
}

func commandHelp(binary string, path []string) (string, error) {
	var args []string = append(append([]string{}, path...), "--help")
	var output []byte
	var err error
	output, err = exec.Command(binary, args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("a6 %s --help failed: %w\n%s", strings.Join(path, " "), err, output)
	}
	return string(output), nil
}

func availableCommands(help string) map[string]bool {
	var commands map[string]bool = map[string]bool{}
	var inCommands bool
	var lines []string = strings.Split(help, "\n")
	var line string
	for _, line = range lines {
		var heading string = strings.TrimSpace(line)
		if heading == "Available Commands:" || heading == "Additional Commands:" {
			inCommands = true
			continue
		}
		if !inCommands {
			continue
		}
		if heading == "" {
			break
		}
		var fields []string = strings.Fields(line)
		if len(fields) > 0 {
			commands[fields[0]] = true
		}
	}
	return commands
}

func availableFlags(help string, longFlagPattern *regexp.Regexp) map[string]bool {
	var flags map[string]bool = map[string]bool{}
	var inFlags bool
	var lines []string = strings.Split(help, "\n")
	var line string
	for _, line = range lines {
		var heading string = strings.TrimSpace(line)
		if heading == "Flags:" || heading == "Global Flags:" {
			inFlags = true
			continue
		}
		if heading == "" {
			inFlags = false
			continue
		}
		if !inFlags {
			continue
		}
		var flag string = longFlagPattern.FindString(line)
		if flag != "" {
			flags[flag] = true
		}
	}
	return flags
}

func availableShortFlags(help string) map[string]bool {
	var flags map[string]bool = map[string]bool{}
	var shortFlagPattern *regexp.Regexp = regexp.MustCompile(`(?:^|\s)(-[A-Za-z])(?:,|\s|$)`)
	var inFlags bool
	var line string
	for _, line = range strings.Split(help, "\n") {
		var heading string = strings.TrimSpace(line)
		if heading == "Flags:" || heading == "Global Flags:" {
			inFlags = true
			continue
		}
		if heading == "" {
			inFlags = false
			continue
		}
		if !inFlags {
			continue
		}
		var match []string = shortFlagPattern.FindStringSubmatch(line)
		if len(match) > 1 {
			flags[match[1]] = true
		}
	}
	return flags
}

func mergeFlagSets(first map[string]bool, second map[string]bool) map[string]bool {
	var merged map[string]bool = map[string]bool{}
	var flag string
	for flag = range first {
		merged[flag] = true
	}
	for flag = range second {
		merged[flag] = true
	}
	return merged
}

func resolveCommand(binary string, fields []string, commands map[string]bool, rootFlags map[string]bool, valueFlags map[string]bool) ([]string, string, error) {
	if len(fields) == 0 || !commands[fields[0]] {
		return nil, "", fmt.Errorf("unsupported a6 command %q", strings.Join(fields, " "))
	}
	var path []string = []string{fields[0]}
	var help string
	var err error
	help, err = commandHelp(binary, path)
	if err != nil {
		return nil, "", err
	}
	var index int = 1
	for index < len(fields) {
		var field string = fields[index]
		if strings.ContainsAny(field, "|<>") {
			break
		}
		var subcommands map[string]bool = availableCommands(help)
		if len(subcommands) == 0 {
			break
		}
		if strings.HasPrefix(field, "-") {
			var flag string = strings.SplitN(field, "=", 2)[0]
			if !rootFlags[flag] {
				return path, help, fmt.Errorf("unsupported interspersed flag %q before a6 subcommand", flag)
			}
			index++
			if valueFlags[flag] && !strings.Contains(field, "=") {
				if index >= len(fields) {
					return path, help, fmt.Errorf("flag %q requires a value", flag)
				}
				index++
			}
			continue
		}
		if !subcommands[field] {
			return path, help, fmt.Errorf("unsupported nested command %q after %q", field, strings.Join(path, " "))
		}
		path = append(path, field)
		help, err = commandHelp(binary, path)
		if err != nil {
			return nil, "", err
		}
		index++
	}
	return path, help, nil
}

func cliInvocations(line string, invocationPattern *regexp.Regexp) []string {
	var invocations []string
	var matches [][]int = invocationPattern.FindAllStringSubmatchIndex(line, -1)
	var match []int
	for _, match = range matches {
		if len(match) >= 4 {
			var start int = match[2]
			var end int = shellInvocationEnd(line, match[3])
			invocations = append(invocations, strings.TrimSpace(line[start:end]))
		}
	}
	return invocations
}

func shellInvocationEnd(line string, start int) int {
	var quote byte
	var escaped bool
	var substitutionDepth int
	var index int
	for index = start; index < len(line); index++ {
		var current byte = line[index]
		if escaped {
			escaped = false
			continue
		}
		if quote != '\'' && current == '\\' {
			escaped = true
			continue
		}
		if quote == '\'' {
			if current == '\'' {
				quote = 0
			}
			continue
		}
		if quote == '"' {
			if current == '"' {
				quote = 0
				continue
			}
			if current == '$' && index+1 < len(line) && line[index+1] == '(' {
				substitutionDepth++
				index++
				continue
			}
			if current == ')' && substitutionDepth > 0 {
				substitutionDepth--
			}
			continue
		}
		if quote == '`' {
			if current == '`' {
				quote = 0
			}
			continue
		}

		switch current {
		case '\'', '"', '`':
			quote = current
		case '$':
			if index+1 < len(line) && line[index+1] == '(' {
				substitutionDepth++
				index++
			}
		case ')':
			if substitutionDepth == 0 {
				return index
			}
			substitutionDepth--
		case '|', ';', '&':
			if substitutionDepth == 0 {
				return index
			}
		}
	}
	return len(line)
}

func joinedShellLines(block string) []string {
	var commands []string
	var current string
	var raw string
	for _, raw = range strings.Split(block, "\n") {
		var line string = strings.TrimSpace(raw)
		if current == "" && (line == "" || strings.HasPrefix(line, "#")) {
			continue
		}
		current += " " + strings.TrimSuffix(line, "\\")
		if strings.HasSuffix(line, "\\") {
			continue
		}
		commands = append(commands, strings.TrimSpace(current))
		current = ""
	}
	if strings.TrimSpace(current) != "" {
		commands = append(commands, strings.TrimSpace(current))
	}
	return commands
}
