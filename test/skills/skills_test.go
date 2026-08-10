package skills

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"

	"github.com/api7/a6/internal/config"
	cmd "github.com/api7/a6/pkg/cmd"
	rootcmd "github.com/api7/a6/pkg/cmd/root"
	"github.com/api7/a6/pkg/iostreams"
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
	shellFencePattern := regexp.MustCompile("(?s)```(?:bash|sh|shell)\\s*\\n(.*?)```")
	yamlFencePattern := regexp.MustCompile("(?s)```(?:yaml|yml)\\s*\\n(.*?)```")
	invocationPattern := regexp.MustCompile(`(?:^|[^A-Za-z0-9_-])(a6)(?:\s|$)`)
	workflowExpressionPattern := regexp.MustCompile(`\$\{\{.*?\}\}`)
	root, err := locateRepoRoot()
	if err != nil {
		t.Fatalf("failed to locate repository root: %v", err)
	}
	binary := buildA6Binary(t, root)
	commandTree := newA6CommandTree(t)
	rootFlags, valueFlags := rootFlagSets(commandTree)
	matches, err := filepath.Glob(filepath.Join(root, "skills", "*", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Fatal("expected at least one skill file")
	}
	rootHelp, err := commandHelp(binary, nil)
	if err != nil {
		t.Fatal(err)
	}
	rootCommands := availableCommands(rootHelp)
	regressions := []string{
		"a6 route creat",
		"a6 route --server https://example.test creat",
		"a6 --bogus route list",
		"a6 --bogus=value route list",
		"a6 route list --Output json",
		"a6 route list --output_json json",
		"a6 route list -Z",
		"a6 route list --output wide",
		"a6 route list --output=wide",
		"a6 route list -owide",
		"a6 route get example --output table",
		"a6 route list unexpected",
		"a6 credential get",
	}
	for _, invocation := range regressions {
		if err := validateA6Invocation(binary, invocation, commandTree, rootCommands, rootFlags, valueFlags); err == nil {
			t.Fatalf("expected invalid invocation %q to fail", invocation)
		}
	}
	for _, invocation := range []string{
		"a6 --output json route list",
		"a6 --output=json route list",
		"a6 route -o yaml get example",
		"a6 route -oyaml get example",
		`a6 debug trace api --header "X-Test: --not-a-flag"`,
	} {
		if err := validateA6Invocation(binary, invocation, commandTree, rootCommands, rootFlags, valueFlags); err != nil {
			t.Fatalf("expected valid invocation %q: %v", invocation, err)
		}
	}
	embedded := cliInvocations("CURRENT=$(a6 route get blue-green)", invocationPattern)
	if len(embedded) != 1 || !strings.HasPrefix(embedded[0], "a6 route get") {
		t.Fatalf("expected embedded a6 invocation, got %q", embedded)
	}
	quoted := cliInvocations(`a6 debug trace id --header "X-Test: a|b;c&d)" --bogus`, invocationPattern)
	if len(quoted) != 1 || !strings.Contains(quoted[0], "--bogus") {
		t.Fatalf("expected quoted separators to preserve the complete invocation, got %q", quoted)
	}
	yamlBlocks, err := skillShellBlocks("```yaml\n- name: Validate\n  run: >\n    a6 route list\n    --unsupported\n```", shellFencePattern, yamlFencePattern)
	if err != nil {
		t.Fatalf("failed to extract workflow run block: %v", err)
	}
	yamlCommands := joinedShellLines(yamlBlocks[0])
	if len(yamlCommands) != 1 || yamlCommands[0] != "a6 route list --unsupported" {
		t.Fatalf("expected workflow run block, got %q", yamlBlocks)
	}

	for _, file := range matches {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		declaredCommands, err := frontmatterA6Commands(string(data))
		if err != nil {
			t.Fatalf("%s: failed to parse frontmatter: %v", file, err)
		}
		for _, declaredCommand := range declaredCommands {
			fields := strings.Fields(declaredCommand)
			if len(fields) < 2 || fields[0] != "a6" {
				t.Fatalf("%s: a6_commands entry %q must start with a6 and include a command", file, declaredCommand)
			}
			commandArgs, err := commandFields(fields[1:], commandTree, rootFlags, valueFlags)
			if err != nil {
				t.Fatalf("%s: a6_commands entry %q is invalid: %v", file, declaredCommand, err)
			}
			path, _, remaining, err := resolveCommand(binary, commandArgs, rootCommands, rootFlags, valueFlags)
			if err != nil {
				t.Fatalf("%s: a6_commands entry %q is invalid: %v", file, declaredCommand, err)
			}
			if len(remaining) != 0 || strings.Join(path, " ") != strings.Join(fields[1:], " ") {
				t.Fatalf("%s: a6_commands entry %q must contain only a command path", file, declaredCommand)
			}
		}
		blocks, err := skillShellBlocks(string(data), shellFencePattern, yamlFencePattern)
		if err != nil {
			t.Fatalf("%s: failed to parse fenced YAML: %v", file, err)
		}
		for _, block := range blocks {
			for _, line := range joinedShellLines(block) {
				for _, invocation := range cliInvocations(line, invocationPattern) {
					invocation = workflowExpressionPattern.ReplaceAllString(invocation, "workflow-expression")
					if err := validateA6Invocation(binary, invocation, commandTree, rootCommands, rootFlags, valueFlags); err != nil {
						t.Fatalf("%s: command %q is invalid: %v", file, invocation, err)
					}
				}
			}
		}
	}
}

func validateA6Invocation(binary, invocation string, root *cobra.Command, rootCommands, rootFlags, valueFlags map[string]bool) error {
	fields, err := shellFields(invocation)
	if err != nil {
		return err
	}
	if len(fields) < 2 || fields[0] != "a6" {
		return nil
	}
	commandArgs, err := commandFields(fields[1:], root, rootFlags, valueFlags)
	if err != nil {
		return err
	}
	path, _, remaining, err := resolveCommand(binary, commandArgs, rootCommands, rootFlags, valueFlags)
	if err != nil {
		return err
	}
	return validatePositionalArgs(root, path, remaining)
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

func commandFields(fields []string, root *cobra.Command, rootFlags, valueFlags map[string]bool) ([]string, error) {
	for len(fields) > 0 && strings.HasPrefix(fields[0], "-") {
		field := fields[0]
		flagName, flag, value, hasInlineValue := rootFlag(root, field)
		if flag == nil || !rootFlags[flagName] {
			return nil, fmt.Errorf("unsupported root flag %q", flagName)
		}
		fields = fields[1:]
		if valueFlags[flagName] && !hasInlineValue {
			if len(fields) == 0 {
				return nil, fmt.Errorf("flag %q requires a value", flagName)
			}
			value = fields[0]
			fields = fields[1:]
		}
		if err := validateKnownFlagValue(flag, value); err != nil {
			return nil, err
		}
	}
	return fields, nil
}

func rootFlag(root *cobra.Command, field string) (string, *pflag.Flag, string, bool) {
	if strings.HasPrefix(field, "--") {
		nameValue := strings.TrimPrefix(field, "--")
		name, value, hasInlineValue := strings.Cut(nameValue, "=")
		return "--" + name, lookupFlag(root, name), value, hasInlineValue
	}
	shorthandValue := strings.TrimPrefix(field, "-")
	if shorthandValue == "" {
		return field, nil, "", false
	}
	shorthand := shorthandValue[:1]
	value := strings.TrimPrefix(shorthandValue[1:], "=")
	return "-" + shorthand, lookupShorthandFlag(root, shorthand), value, len(shorthandValue) > 1
}

func rootFlagSets(root *cobra.Command) (map[string]bool, map[string]bool) {
	rootFlags := map[string]bool{}
	valueFlags := map[string]bool{}
	root.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		longName := "--" + flag.Name
		rootFlags[longName] = true
		if flag.NoOptDefVal == "" {
			valueFlags[longName] = true
		}
		if flag.Shorthand != "" {
			shortName := "-" + flag.Shorthand
			rootFlags[shortName] = true
			if flag.NoOptDefVal == "" {
				valueFlags[shortName] = true
			}
		}
	})
	return rootFlags, valueFlags
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

func resolveCommand(binary string, fields []string, commands, rootFlags, valueFlags map[string]bool) ([]string, string, []string, error) {
	if len(fields) == 0 || !commands[fields[0]] {
		return nil, "", nil, fmt.Errorf("unsupported a6 command %q", strings.Join(fields, " "))
	}
	var path []string = []string{fields[0]}
	var help string
	var err error
	help, err = commandHelp(binary, path)
	if err != nil {
		return nil, "", nil, err
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
			flag := strings.SplitN(field, "=", 2)[0]
			hasInlineValue := strings.Contains(field, "=")
			if strings.HasPrefix(flag, "-") && !strings.HasPrefix(flag, "--") && len(flag) > 2 {
				flag = flag[:2]
				hasInlineValue = true
			}
			if !rootFlags[flag] {
				return path, help, nil, fmt.Errorf("unsupported interspersed flag %q before a6 subcommand", flag)
			}
			index++
			if valueFlags[flag] && !hasInlineValue {
				if index >= len(fields) {
					return path, help, nil, fmt.Errorf("flag %q requires a value", flag)
				}
				index++
			}
			continue
		}
		if !subcommands[field] {
			return path, help, nil, fmt.Errorf("unsupported nested command %q after %q", field, strings.Join(path, " "))
		}
		path = append(path, field)
		help, err = commandHelp(binary, path)
		if err != nil {
			return nil, "", nil, err
		}
		index++
	}
	return path, help, fields[index:], nil
}

func newA6CommandTree(t *testing.T) *cobra.Command {
	t.Helper()
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewFileConfigWithPath(filepath.Join(t.TempDir(), "config.yaml"))
	factory := &cmd.Factory{
		IOStreams: ios,
		HttpClient: func() (*http.Client, error) {
			return http.DefaultClient, nil
		},
		Config: func() (config.Config, error) {
			return cfg, nil
		},
	}
	return rootcmd.NewCmdRoot(factory)
}

func validatePositionalArgs(root *cobra.Command, path, fields []string) error {
	command, remainingPath, err := root.Find(path)
	if err != nil {
		return err
	}
	if len(remainingPath) != 0 {
		return fmt.Errorf("failed to resolve command path %q", strings.Join(path, " "))
	}
	args, err := positionalArgs(command, fields)
	if err != nil {
		return err
	}
	return command.ValidateArgs(args)
}

func positionalArgs(command *cobra.Command, fields []string) ([]string, error) {
	var args []string
	for index := 0; index < len(fields); index++ {
		field := fields[index]
		if strings.ContainsAny(field, "|<>") {
			break
		}
		if field == "--" {
			for _, arg := range fields[index+1:] {
				if strings.ContainsAny(arg, "|<>") {
					break
				}
				args = append(args, arg)
			}
			break
		}
		if strings.HasPrefix(field, "--") {
			nameValue := strings.TrimPrefix(field, "--")
			name, value, hasInlineValue := strings.Cut(nameValue, "=")
			flag := lookupFlag(command, name)
			if flag == nil {
				return nil, fmt.Errorf("unsupported flag %q", field)
			}
			if !hasInlineValue && flag.NoOptDefVal == "" {
				index++
				if index >= len(fields) {
					return nil, fmt.Errorf("flag %q requires a value", field)
				}
				value = fields[index]
			}
			if err := validateKnownFlagValue(flag, value); err != nil {
				return nil, err
			}
			continue
		}
		if strings.HasPrefix(field, "-") && field != "-" {
			shorthandValue := strings.TrimPrefix(field, "-")
			shorthand := shorthandValue[:1]
			flag := lookupShorthandFlag(command, shorthand)
			if flag == nil {
				return nil, fmt.Errorf("unsupported shorthand flag %q", field)
			}
			hasInlineValue := len(shorthandValue) > 1
			value := strings.TrimPrefix(shorthandValue[1:], "=")
			if !hasInlineValue && flag.NoOptDefVal == "" {
				index++
				if index >= len(fields) {
					return nil, fmt.Errorf("flag %q requires a value", field)
				}
				value = fields[index]
			}
			if err := validateKnownFlagValue(flag, value); err != nil {
				return nil, err
			}
			continue
		}
		args = append(args, field)
	}
	return args, nil
}

func validateKnownFlagValue(flag *pflag.Flag, value string) error {
	if flag == nil || flag.Name != "output" || value == "" || value == "workflow-expression" || strings.HasPrefix(value, "$") || strings.HasPrefix(value, "<") {
		return nil
	}
	_, formats, ok := strings.Cut(flag.Usage, ":")
	if !ok {
		return nil
	}
	for _, format := range strings.Split(formats, ",") {
		if value == strings.TrimSpace(format) {
			return nil
		}
	}
	return fmt.Errorf("unsupported output format %q", value)
}

func lookupFlag(command *cobra.Command, name string) *pflag.Flag {
	if flag := command.Flags().Lookup(name); flag != nil {
		return flag
	}
	if flag := command.InheritedFlags().Lookup(name); flag != nil {
		return flag
	}
	return command.Root().PersistentFlags().Lookup(name)
}

func lookupShorthandFlag(command *cobra.Command, shorthand string) *pflag.Flag {
	if flag := command.Flags().ShorthandLookup(shorthand); flag != nil {
		return flag
	}
	if flag := command.InheritedFlags().ShorthandLookup(shorthand); flag != nil {
		return flag
	}
	return command.Root().PersistentFlags().ShorthandLookup(shorthand)
}

func shellFields(line string) ([]string, error) {
	var fields []string
	var current strings.Builder
	var quote rune
	var escaped bool
	var started bool
	for _, char := range line {
		if escaped {
			current.WriteRune(char)
			escaped = false
			started = true
			continue
		}
		if quote != 0 {
			if char == quote {
				quote = 0
				continue
			}
			if quote == '"' && char == '\\' {
				escaped = true
				continue
			}
			current.WriteRune(char)
			started = true
			continue
		}
		switch {
		case char == '\\':
			escaped = true
			started = true
		case char == '\'' || char == '"':
			quote = char
			started = true
		case unicode.IsSpace(char):
			if started {
				fields = append(fields, current.String())
				current.Reset()
				started = false
			}
		default:
			current.WriteRune(char)
			started = true
		}
	}
	if escaped {
		return nil, fmt.Errorf("unfinished escape")
	}
	if quote != 0 {
		return nil, fmt.Errorf("unterminated quote %q", string(quote))
	}
	if started {
		fields = append(fields, current.String())
	}
	return fields, nil
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
