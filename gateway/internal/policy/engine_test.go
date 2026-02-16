package policy

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MohamadRazaviDev/GrafanaGateWay/gateway/internal/auth"
	"github.com/MohamadRazaviDev/GrafanaGateWay/gateway/internal/config"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func TestNoPolicies(t *testing.T) {
	engine := NewEngine(nil, testLogger())
	req := httptest.NewRequest("GET", "/api/dashboards", nil)
	id := &auth.Identity{User: "alice", Team: "sre"}

	d := engine.Evaluate(req, id)
	if !d.Allowed {
		t.Error("expected allowed when no policies configured")
	}
}

func TestAdminBlocked(t *testing.T) {
	engine := NewEngine(nil, testLogger())
	req := httptest.NewRequest("GET", "/api/admin/users", nil)
	id := &auth.Identity{User: "alice", Team: "sre"}

	d := engine.Evaluate(req, id)
	if d.Allowed {
		t.Error("expected admin path to be blocked by default")
	}
}

func TestAdminExplicitAllow(t *testing.T) {
	policies := []config.Policy{
		{
			Name:       "admin-access",
			Users:      []string{"superadmin"},
			AllowPaths: []string{"/api/admin/*"},
		},
	}
	engine := NewEngine(policies, testLogger())
	req := httptest.NewRequest("GET", "/api/admin/users", nil)

	// superadmin should be allowed
	d := engine.Evaluate(req, &auth.Identity{User: "superadmin"})
	if !d.Allowed {
		t.Error("expected superadmin to access admin paths")
	}

	// regular user should be blocked
	d = engine.Evaluate(req, &auth.Identity{User: "alice"})
	if d.Allowed {
		t.Error("expected regular user to be blocked from admin paths")
	}
}

func TestAllowByTeam(t *testing.T) {
	policies := []config.Policy{
		{
			Name:       "sre-dashboards",
			Teams:      []string{"sre"},
			AllowPaths: []string{"/d/*", "/api/dashboards/*"},
		},
	}
	engine := NewEngine(policies, testLogger())

	req := httptest.NewRequest("GET", "/d/abc123", nil)
	d := engine.Evaluate(req, &auth.Identity{User: "alice", Team: "sre"})
	if !d.Allowed {
		t.Error("expected SRE team to access dashboards")
	}

	d = engine.Evaluate(req, &auth.Identity{User: "bob", Team: "dev"})
	if d.Allowed {
		t.Error("expected dev team to be denied")
	}
}

func TestDenyTakesPrecedence(t *testing.T) {
	policies := []config.Policy{
		{
			Name:       "limited-access",
			Users:      []string{"*"},
			AllowPaths: []string{"/api/*"},
			DenyPaths:  []string{"/api/org/*"},
		},
	}
	engine := NewEngine(policies, testLogger())

	req := httptest.NewRequest("GET", "/api/org/settings", nil)
	d := engine.Evaluate(req, &auth.Identity{User: "alice"})
	if d.Allowed {
		t.Error("expected deny path to take precedence over allow")
	}
}

func TestMethodRestriction(t *testing.T) {
	policies := []config.Policy{
		{
			Name:         "read-only",
			Users:        []string{"*"},
			AllowPaths:   []string{"/api/dashboards/*"},
			AllowMethods: []string{"GET"},
		},
	}
	engine := NewEngine(policies, testLogger())

	req := httptest.NewRequest("GET", "/api/dashboards/abc", nil)
	d := engine.Evaluate(req, &auth.Identity{User: "alice"})
	if !d.Allowed {
		t.Error("expected GET to be allowed")
	}

	req = httptest.NewRequest("DELETE", "/api/dashboards/abc", nil)
	d = engine.Evaluate(req, &auth.Identity{User: "alice"})
	if d.Allowed {
		t.Error("expected DELETE to be denied")
	}
}

func TestMiddleware(t *testing.T) {
	policies := []config.Policy{
		{
			Name:       "allow-all",
			Users:      []string{"*"},
			AllowPaths: []string{"/*"},
		},
	}
	engine := NewEngine(policies, testLogger())

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := engine.Middleware()(next)
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Error("expected next handler to be called")
	}
}

func TestMatchPath(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		want    bool
	}{
		{"/api/admin/*", "/api/admin/users", true},
		{"/api/admin/*", "/api/admin", true},
		{"/api/admin", "/api/admin", true},
		{"/api/admin", "/api/admin/users", false},
		{"/d/*", "/d/abc123", true},
		{"/healthz", "/healthz", true},
		{"/healthz", "/readyz", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"→"+tt.path, func(t *testing.T) {
			if got := matchPath(tt.pattern, tt.path); got != tt.want {
				t.Errorf("matchPath(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
			}
		})
	}
}
