package api

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var canonicalPaths = map[string]bool{
	"/apisix/admin/routes":          true,
	"/apisix/admin/services":        true,
	"/apisix/admin/upstreams":       true,
	"/apisix/admin/consumers":       true,
	"/apisix/admin/ssls":            true,
	"/apisix/admin/global_rules":    true,
	"/apisix/admin/plugin_configs":  true,
	"/apisix/admin/consumer_groups": true,
	"/apisix/admin/stream_routes":   true,
	"/apisix/admin/protos":          true,
	"/apisix/admin/secrets":         true,
	"/apisix/admin/plugins":         true,
	"/apisix/admin/plugin_metadata": true,
}

var pathRegexp = regexp.MustCompile(`/apisix/admin/[a-z_]+`)

// TestCanonicalAPIPaths scans Go source files under pkg/cmd/ for hardcoded API
// path strings and validates each against the canonical Admin API spec. This
// would have caught the /apisix/admin/ssl typo (correct: /apisix/admin/ssls).
func TestCanonicalAPIPaths(t *testing.T) {
	projectRoot := findProjectRoot(t)

	cmdDir := filepath.Join(projectRoot, "pkg", "cmd")
	if _, err := os.Stat(cmdDir); os.IsNotExist(err) {
		t.Fatalf("pkg/cmd/ directory not found at %s", cmdDir)
	}

	var violations []string

	err := filepath.Walk(cmdDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(info.Name(), ".go") || strings.HasSuffix(info.Name(), "_test.go") {
			return nil
		}

		fset := token.NewFileSet()
		node, parseErr := parser.ParseFile(fset, path, nil, 0)
		if parseErr != nil {
			return nil
		}

		ast.Inspect(node, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}

			val := strings.Trim(lit.Value, `"`)

			matches := pathRegexp.FindAllString(val, -1)
			for _, match := range matches {
				basePath := extractBasePath(match)
				if basePath != "" && !canonicalPaths[basePath] {
					relPath, _ := filepath.Rel(projectRoot, path)
					pos := fset.Position(lit.Pos())
					violations = append(violations, relPath+":"+
						strings.TrimPrefix(pos.String(), path+":")+
						" invalid API path "+match+
						" (base: "+basePath+")")
				}
			}
			return true
		})

		return nil
	})

	if err != nil {
		t.Fatalf("failed to walk pkg/cmd/: %v", err)
	}

	for _, v := range violations {
		t.Errorf("non-canonical API path: %s", v)
	}

	if len(violations) == 0 {
		t.Logf("All API paths in pkg/cmd/ match canonical spec (%d valid paths)", len(canonicalPaths))
	}
}

// TestKnownBadPathsDetected verifies the scanner catches historical typos
// like /apisix/admin/ssl (correct: /apisix/admin/ssls).
func TestKnownBadPathsDetected(t *testing.T) {
	knownBadPaths := []string{
		"/apisix/admin/ssl",
		"/apisix/admin/route",
		"/apisix/admin/service",
		"/apisix/admin/upstream",
		"/apisix/admin/consumer",
		"/apisix/admin/global_rule",
		"/apisix/admin/plugin_config",
		"/apisix/admin/consumer_group",
		"/apisix/admin/stream_route",
		"/apisix/admin/proto",
		"/apisix/admin/secret",
	}

	for _, bad := range knownBadPaths {
		basePath := extractBasePath(bad)
		assert.False(t, canonicalPaths[basePath],
			"path %q should NOT be in canonical set (it's a known bad path)", bad)
	}
}

func extractBasePath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		return ""
	}
	return "/" + parts[1] + "/" + parts[2] + "/" + parts[3]
}

func findProjectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find project root (no go.mod found)")
		}
		dir = parent
	}
}
