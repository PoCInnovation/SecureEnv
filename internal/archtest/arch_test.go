package archtest

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const module = "github.com/PoCInnovation/SecureEnv/"

// allowed lists, for each package, the internal packages it may import in
// non-test code. Anything missing from this table is a violation, which keeps
// the core independent from transport and storage details.
var allowed = map[string][]string{
	"internal/domain":              {},
	"internal/dotenv":              {"internal/domain"},
	"internal/project":             {"internal/domain"},
	"internal/project/projecttest": {"internal/domain"},
	"internal/project/storetest":   {"internal/domain", "internal/project"},
	"internal/vaultstore":          {"internal/domain"},
	"internal/apiv1":               {"internal/domain"},
	"internal/httpapi":             {"internal/domain", "internal/apiv1"},
	"internal/apiclient":           {"internal/domain", "internal/apiv1"},
	"internal/envsync":             {"internal/domain", "internal/dotenv"},
	"internal/gitremote":           {},
	"internal/cli":                 {"internal/domain", "internal/apiv1", "internal/dotenv", "internal/envsync", "internal/gitremote"},
	"internal/apiserver":           {"internal/httpapi", "internal/project", "internal/vaultstore"},
	"internal/archtest":            {},
	"cmd/secureenv":                {"internal/apiclient", "internal/cli", "internal/gitremote"},
	"cmd/secureenv-api":            {"internal/apiserver"},
}

func TestDependencyRules(t *testing.T) {
	imports := collectImports(t, "../..")

	for pkg, deps := range imports {
		rules, known := allowed[pkg]
		if !known {
			t.Errorf("package %s has no dependency rule, add it to allowed", pkg)
			continue
		}
		for _, dep := range deps {
			if !slices.Contains(rules, dep) {
				t.Errorf("%s must not import %s", pkg, dep)
			}
		}
	}
}

// collectImports returns, for each package under root, the internal packages
// imported by its non-test files.
func collectImports(t *testing.T, root string) map[string][]string {
	t.Helper()
	imports := map[string][]string{}
	fset := token.NewFileSet()

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (d.Name() == ".git" || d.Name() == "deploy") {
			return filepath.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, filepath.Dir(path))
		if err != nil {
			return err
		}
		pkg := filepath.ToSlash(rel)
		if _, ok := imports[pkg]; !ok {
			imports[pkg] = nil
		}
		for _, spec := range file.Imports {
			dep, ok := strings.CutPrefix(strings.Trim(spec.Path.Value, `"`), module)
			if ok && !slices.Contains(imports[pkg], dep) {
				imports[pkg] = append(imports[pkg], dep)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return imports
}
