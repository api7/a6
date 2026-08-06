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

var shellFencePattern *regexp.Regexp = regexp.MustCompile("(?s)```(?:bash|sh|shell)\\s*\\n(.*?)```")
var longFlagPattern *regexp.Regexp = regexp.MustCompile(`--[a-z][a-z0-9-]*`)

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
	var rootFlags map[string]bool = availableFlags(rootHelp)

	var regressionPath []string
	var regressionHelp string
	regressionPath, regressionHelp, err = resolveCommand(binary, []string{"route", "creat"}, rootCommands)
	if err == nil {
		t.Fatalf("expected misspelled nested command to fail, got path %q and help %q", regressionPath, regressionHelp)
	}

	var file string
	for _, file = range matches {
		var data []byte
		data, err = os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		var blocks [][]string = shellFencePattern.FindAllStringSubmatch(string(data), -1)
		var block []string
		for _, block = range blocks {
			var lines []string = joinedShellLines(block[1])
			var line string
			for _, line = range lines {
				var fields []string = strings.Fields(line)
				if len(fields) < 2 || fields[0] != "a6" {
					continue
				}
				var path []string
				var help string
				path, help, err = resolveCommand(binary, commandFields(fields[1:]), rootCommands)
				if err != nil {
					t.Fatalf("%s: %v", file, err)
				}
				var validFlags map[string]bool = mergeFlagSets(rootFlags, availableFlags(help))
				var flags []string = longFlagPattern.FindAllString(line, -1)
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

func commandFields(fields []string) []string {
	var valueFlags map[string]bool = map[string]bool{
		"--api-key": true,
		"--context": true,
		"--output":  true,
		"--server":  true,
		"-o":        true,
	}
	for len(fields) > 0 && strings.HasPrefix(fields[0], "-") {
		var flag string = fields[0]
		fields = fields[1:]
		if valueFlags[flag] && len(fields) > 0 {
			fields = fields[1:]
		}
	}
	return fields
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

func availableFlags(help string) map[string]bool {
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

func resolveCommand(binary string, fields []string, commands map[string]bool) ([]string, string, error) {
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
	var field string
	for _, field = range fields[1:] {
		if strings.HasPrefix(field, "-") || strings.ContainsAny(field, "|<>") {
			break
		}
		var subcommands map[string]bool = availableCommands(help)
		if len(subcommands) == 0 {
			break
		}
		if !subcommands[field] {
			return path, help, fmt.Errorf("unsupported nested command %q after %q", field, strings.Join(path, " "))
		}
		path = append(path, field)
		help, err = commandHelp(binary, path)
		if err != nil {
			return nil, "", err
		}
	}
	return path, help, nil
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
