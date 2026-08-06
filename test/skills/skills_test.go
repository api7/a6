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

var shellFencePattern = regexp.MustCompile("(?s)```(?:bash|sh|shell)\\s*\\n(.*?)```")
var longFlagPattern = regexp.MustCompile(`--[a-z][a-z0-9-]*`)
var a6Binary string

func locateRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

func TestMain(m *testing.M) {
	root, err := locateRepoRoot()
	if err != nil {
		os.Exit(1)
	}
	tmpDir, err := os.MkdirTemp("", "a6-skills-test-*")
	if err != nil {
		os.Exit(1)
	}
	a6Binary = filepath.Join(tmpDir, "a6")
	cmd := exec.Command("go", "build", "-o", a6Binary, "./cmd/a6")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		_ = os.RemoveAll(tmpDir)
		os.Exit(1)
	}
	exitCode := m.Run()
	if err := os.RemoveAll(tmpDir); err != nil && exitCode == 0 {
		fmt.Fprintf(os.Stderr, "failed to remove temp dir %s: %v\n", tmpDir, err)
		exitCode = 1
	}
	os.Exit(exitCode)
}

func TestSkillShellExamplesUseSupportedA6CommandsAndFlags(t *testing.T) {
	root, err := locateRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	matches, err := filepath.Glob(filepath.Join(root, "skills", "*", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	rootHelp := commandHelp(t, nil)
	rootCommands := availableCommands(rootHelp)
	for _, file := range matches {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, block := range shellFencePattern.FindAllStringSubmatch(string(data), -1) {
			for _, line := range joinedShellLines(block[1]) {
				fields := strings.Fields(line)
				if len(fields) < 2 || fields[0] != "a6" {
					continue
				}
				path, help := resolveCommand(t, file, commandFields(fields[1:]), rootCommands)
				validHelp := rootHelp + "\n" + help
				for _, flag := range longFlagPattern.FindAllString(line, -1) {
					if flag != "--help" && !strings.Contains(validHelp, flag) {
						t.Fatalf("%s: command %q uses unsupported flag %q", file, "a6 "+strings.Join(path, " "), flag)
					}
				}
			}
		}
	}
}

func commandFields(fields []string) []string {
	valueFlags := map[string]bool{"--api-key": true, "--context": true, "--output": true, "--server": true, "-o": true}
	for len(fields) > 0 && strings.HasPrefix(fields[0], "-") {
		flag := fields[0]
		fields = fields[1:]
		if valueFlags[flag] && len(fields) > 0 {
			fields = fields[1:]
		}
	}
	return fields
}

func commandHelp(t *testing.T, path []string) string {
	t.Helper()
	args := append(append([]string{}, path...), "--help")
	output, err := exec.Command(a6Binary, args...).CombinedOutput()
	if err != nil {
		t.Fatalf("a6 %s --help failed: %v\n%s", strings.Join(path, " "), err, output)
	}
	return string(output)
}

func availableCommands(help string) map[string]bool {
	commands := map[string]bool{}
	inCommands := false
	for _, line := range strings.Split(help, "\n") {
		heading := strings.TrimSpace(line)
		if heading == "Available Commands:" || heading == "Additional Commands:" {
			inCommands = true
			continue
		}
		if !inCommands {
			continue
		}
		if strings.TrimSpace(line) == "" {
			break
		}
		fields := strings.Fields(line)
		if len(fields) > 0 {
			commands[fields[0]] = true
		}
	}
	return commands
}

func resolveCommand(t *testing.T, file string, fields []string, commands map[string]bool) ([]string, string) {
	t.Helper()
	if len(fields) == 0 || !commands[fields[0]] {
		t.Fatalf("%s: unsupported a6 command %q", file, strings.Join(fields, " "))
	}
	path := []string{fields[0]}
	help := commandHelp(t, path)
	for _, field := range fields[1:] {
		if strings.HasPrefix(field, "-") || strings.ContainsAny(field, "|<>") {
			break
		}
		subcommands := availableCommands(help)
		if !subcommands[field] {
			break
		}
		path = append(path, field)
		help = commandHelp(t, path)
	}
	return path, help
}

func joinedShellLines(block string) []string {
	var commands []string
	var current string
	for _, raw := range strings.Split(block, "\n") {
		line := strings.TrimSpace(raw)
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
