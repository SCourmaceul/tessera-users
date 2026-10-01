package tesserausers_test

import (
	"os/exec"
	"strings"
	"testing"
)

// Règle de dépendance : le domaine et les cas d'usage ne dépendent ni d'AWS, ni de HTTP, ni
// des adaptateurs (ceux du service comme ceux de tetra-kit), même indirectement.
func TestInnerLayersStayFreeOfInfrastructure(t *testing.T) {
	inner := []string{"./internal/domain", "./internal/app"}
	forbidden := []string{
		"github.com/aws/",
		"github.com/SCourmaceul/tetra-kit/adapter",
		"github.com/SCourmaceul/tessera-users/internal/adapter",
		"github.com/SCourmaceul/tessera-users/internal/platform",
		"net/http",
	}
	out, err := exec.Command("go", append([]string{"list", "-deps"}, inner...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("go list: %v\n%s", err, out)
	}
	for _, dep := range strings.Fields(string(out)) {
		for _, prefix := range forbidden {
			if strings.HasPrefix(dep, prefix) {
				t.Errorf("le domaine ou les cas d'usage dépendent de %s", dep)
			}
		}
	}
}
