package discovery

import (
	"strings"
	"testing"

	"github.com/areppel/bauhaus/internal/config"
)

// Regression test for a bug that renamed the user's Mac.
//
// brutella/dnssd is a standalone responder: it publishes A/AAAA records claiming
// whatever Config.Host is set to. macOS's mDNSResponder already owns
// <LocalHostName>.local. When Bauhaus claimed that same name, macOS detected a
// collision and renamed the machine (MacStudio -> MacStudio-2) — a persistent
// change to the user's system settings.
//
// The name we publish under must therefore never equal the machine's own.
func TestServiceHostNeverClaimsTheMachineHostname(t *testing.T) {
	local := config.LocalHostName()
	if local == "" {
		t.Skip("no LocalHostName on this machine")
	}

	got := serviceHost(local)

	if strings.EqualFold(got, local) {
		t.Fatalf("serviceHost(%q) = %q — publishing address records for the machine's own "+
			"hostname makes macOS rename the machine to avoid the collision", local, got)
	}
	if !strings.HasPrefix(got, "bauhaus-") {
		t.Errorf("serviceHost(%q) = %q, want a bauhaus- prefixed name that nothing else can own", local, got)
	}
}

func TestServiceHostIsALegalDNSLabel(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"MacStudio", "bauhaus-macstudio"},
		{"Alex's iMac", "bauhaus-alex-s-imac"},
		{"Mac-Pro-2", "bauhaus-mac-pro-2"},
		{"", "bauhaus-host"},
	}
	for _, tt := range tests {
		got := serviceHost(tt.in)
		if got != tt.want {
			t.Errorf("serviceHost(%q) = %q, want %q", tt.in, got, tt.want)
		}
		for _, r := range got {
			legal := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-'
			if !legal {
				t.Errorf("serviceHost(%q) = %q contains %q, which is not legal in a DNS label", tt.in, got, r)
			}
		}
	}
}
