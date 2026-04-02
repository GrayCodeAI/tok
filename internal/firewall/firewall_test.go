package firewall

import "testing"

func TestNewEgressFirewall(t *testing.T) {
	f := NewEgressFirewall()
	if f == nil {
		t.Fatal("Expected non-nil firewall")
	}
}

func TestEgressFirewallAllow(t *testing.T) {
	f := NewEgressFirewall()
	f.AddAllowedDomain("api.example.com")
	f.AddAllowedIP("192.168.1.1")

	if !f.IsAllowed("api.example.com", "1.2.3.4") {
		t.Error("Expected allowed domain to pass")
	}
	if !f.IsAllowed("other.com", "192.168.1.1") {
		t.Error("Expected allowed IP to pass")
	}
	if f.IsAllowed("other.com", "10.0.0.1") {
		t.Error("Expected disallowed domain/IP to fail")
	}
}

func TestEgressFirewallRules(t *testing.T) {
	f := NewEgressFirewall()
	rules := f.GenerateIptablesRules()
	if len(rules) == 0 {
		t.Error("Expected iptables rules")
	}

	status := f.Status()
	if status == nil {
		t.Error("Expected status")
	}
}
