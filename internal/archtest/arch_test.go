package archtest_test

import (
	"os/exec"
	"strings"
	"testing"
)

// The UI toolkit (systray/AppKit, via cgo) must stay confined to cmd/bauhaus.
// If it leaks into internal/, the business logic can no longer be tested without
// a display server, and the daemon can no longer run headless under launchd.
func TestNoGUIToolkitInInternalPackages(t *testing.T) {
	// Query the module's own internal packages (excluding this archtest package,
	// which has no non-test files and so is not a build target).
	out, err := exec.Command("go", "list",
		"github.com/areppel/bauhaus/internal/app",
		"github.com/areppel/bauhaus/internal/config",
		"github.com/areppel/bauhaus/internal/discovery",
		"github.com/areppel/bauhaus/internal/gateway",
		"github.com/areppel/bauhaus/internal/hub",
		"github.com/areppel/bauhaus/internal/registry",
		"github.com/areppel/bauhaus/internal/runtime",
		"github.com/areppel/bauhaus/internal/ui",
	).CombinedOutput()
	if err != nil {
		t.Fatalf("go list: %v\n%s", err, out)
	}
	// Now list each package's full dependency set.
	out, err = exec.Command("go", "list", "-deps",
		"github.com/areppel/bauhaus/internal/app",
		"github.com/areppel/bauhaus/internal/gateway",
		"github.com/areppel/bauhaus/internal/runtime",
		"github.com/areppel/bauhaus/internal/discovery",
	).CombinedOutput()
	if err != nil {
		t.Fatalf("go list -deps: %v\n%s", err, out)
	}
	for _, dep := range strings.Fields(string(out)) {
		if strings.Contains(dep, "fyne.io/systray") {
			t.Errorf("an internal package imports %s — the GUI toolkit must stay in cmd/bauhaus, "+
				"or internal packages can no longer run headless or be tested without a display", dep)
		}
	}
}
