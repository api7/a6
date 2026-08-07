package skills

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
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

func TestSkillShellExamplesUseSupportedA6CommandsAndFlags(t *testing.T) {
	var shellFencePattern *regexp.Regexp = regexp.MustCompile("(?s)```(?:bash|sh|shell)\\s*\\n(.*?)```")
	var yamlFencePattern *regexp.Regexp = regexp.MustCompile("(?s)```(?:yaml|yml)\\s*\\n(.*?)```")
	var longFlagPattern *regexp.Regexp = regexp.MustCompile(`--[a-z][a-z0-9-]*`)
	var invocationPattern *regexp.Regexp = regexp.MustCompile(`(?:^|[^A-Za-z0-9_-])(a6(?:\s+[^|;&)]*)?)`)
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
	var rootFlags map[string]bool = availableFlags(rootHelp, longFlagPattern)

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
	var embedded []string = cliInvocations("CURRENT=$(a6 route get blue-green)", invocationPattern)
	if len(embedded) != 1 || !strings.HasPrefix(embedded[0], "a6 route get") {
		t.Fatalf("expected embedded a6 invocation, got %q", embedded)
	}
	var yamlBlocks []string = skillShellBlocks("```yaml\n- name: Validate\n  run: |\n    a6 route list\n    a6 config validate -f config.yaml\n```", shellFencePattern, yamlFencePattern)
	if len(yamlBlocks) != 1 || !strings.Contains(yamlBlocks[0], "a6 config validate") {
		t.Fatalf("expected workflow run block, got %q", yamlBlocks)
	}

	var file string
	for _, file = range matches {
		var data []byte
		data, err = os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		var blocks []string = skillShellBlocks(string(data), shellFencePattern, yamlFencePattern)
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

func skillShellBlocks(data string, shellFencePattern *regexp.Regexp, yamlFencePattern *regexp.Regexp) []string {
	var blocks []string
	var match []string
	for _, match = range shellFencePattern.FindAllStringSubmatch(data, -1) {
		blocks = append(blocks, match[1])
	}
	for _, match = range yamlFencePattern.FindAllStringSubmatch(data, -1) {
		blocks = append(blocks, yamlRunBlocks(match[1])...)
	}
	return blocks
}

func yamlRunBlocks(block string) []string {
	var runBlocks []string
	var lines []string = strings.Split(block, "\n")
	var index int
	for index = 0; index < len(lines); index++ {
		var line string = lines[index]
		var trimmed string = strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "run:") {
			continue
		}
		var value string = strings.TrimSpace(strings.TrimPrefix(trimmed, "run:"))
		if value == "" {
			continue
		}
		if value[0] != '|' && value[0] != '>' {
			runBlocks = append(runBlocks, value)
			continue
		}

		var runIndent int = len(line) - len(strings.TrimLeft(line, " \t"))
		var body []string
		var next int
		for next = index + 1; next < len(lines); next++ {
			var bodyLine string = lines[next]
			if strings.TrimSpace(bodyLine) == "" {
				body = append(body, bodyLine)
				continue
			}
			var bodyIndent int = len(bodyLine) - len(strings.TrimLeft(bodyLine, " \t"))
			if bodyIndent <= runIndent {
				break
			}
			body = append(body, bodyLine)
		}
		runBlocks = append(runBlocks, strings.Join(body, "\n"))
		index = next - 1
	}
	return runBlocks
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
	var matches [][]string = invocationPattern.FindAllStringSubmatch(line, -1)
	var match []string
	for _, match = range matches {
		if len(match) > 1 {
			invocations = append(invocations, strings.TrimSpace(match[1]))
		}
	}
	return invocations
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
