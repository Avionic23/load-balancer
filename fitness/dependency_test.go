package fitness

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

func TestDependencyDirection(t *testing.T) {
	hierarchy := []string{
		"listener",
		"proxy",
		"router",
		"backend",
	}

	for i, pkg := range hierarchy {
		imports := getImports("load-balancer/" + pkg)
		for _, imp := range imports {
			for j := 0; j < i; j++ {
				if strings.Contains(imp, "load-balancer/"+hierarchy[j]) {
					t.Errorf(
						"%s imports %s — violates downward dependency rule",
						pkg, hierarchy[j],
					)
				}
			}
		}
	}
}

func getImports(pkg string) []string {
	cmd := exec.Command("go", "list", "-json", pkg)
	out, _ := cmd.Output()
	var info struct{ Imports []string }
	json.Unmarshal(out, &info)
	return info.Imports
}
