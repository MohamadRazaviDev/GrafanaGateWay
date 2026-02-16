package policy

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"path"
	"strings"

	"github.com/MohamadRazaviDev/GrafanaGateWay/gateway/internal/auth"
	"github.com/MohamadRazaviDev/GrafanaGateWay/gateway/internal/config"
)

type contextKey string

const DecisionKey contextKey = "gateway.policy_decision"

// GetDecision retrieves the policy decision from the request context.
func GetDecision(ctx context.Context) *Decision {
	if d, ok := ctx.Value(DecisionKey).(*Decision); ok {
		return d
	}
	return nil
}

// Engine evaluates authorization policies against incoming requests.
type Engine struct {
	policies     []config.Policy
	defaultDeny  bool
	blockedPaths []string
	logger       *slog.Logger
}

// Decision represents a policy evaluation result.
type Decision struct {
	Allowed    bool
	PolicyName string
	Reason     string
}

// NewEngine creates a policy engine from config.
func NewEngine(policies []config.Policy, logger *slog.Logger) *Engine {
	return &Engine{
		policies:     policies,
		defaultDeny:  true,
		blockedPaths: []string{"/api/admin/*", "/api/admin"},
		logger:       logger,
	}
}

// Evaluate checks if the given request is allowed by the configured policies.
func (e *Engine) Evaluate(r *http.Request, identity *auth.Identity) Decision {
	reqPath := r.URL.Path
	method := r.Method

	// Always block dangerous admin endpoints unless explicitly allowed
	for _, blocked := range e.blockedPaths {
		if matchPath(blocked, reqPath) {
			// Check if any policy explicitly allows admin access
			if !e.hasExplicitAllow(reqPath, method, identity) {
				return Decision{
					Allowed:    false,
					PolicyName: "_builtin_admin_block",
					Reason:     "admin endpoints are blocked by default",
				}
			}
		}
	}

	// If no policies configured, allow everything (proxy-only mode)
	if len(e.policies) == 0 {
		return Decision{Allowed: true, PolicyName: "_default", Reason: "no policies configured"}
	}

	// Evaluate policies in order
	for _, pol := range e.policies {
		if !e.matchesIdentity(pol, identity) {
			continue
		}

		// Check deny paths first (deny takes precedence)
		for _, dp := range pol.DenyPaths {
			if matchPath(dp, reqPath) {
				return Decision{
					Allowed:    false,
					PolicyName: pol.Name,
					Reason:     "path denied by policy",
				}
			}
		}

		// Check allow paths
		for _, ap := range pol.AllowPaths {
			if matchPath(ap, reqPath) {
				// Check method restriction if specified
				if len(pol.AllowMethods) > 0 && !containsMethod(pol.AllowMethods, method) {
					continue
				}
				return Decision{
					Allowed:    true,
					PolicyName: pol.Name,
					Reason:     "path allowed by policy",
				}
			}
		}
	}

	// Default deny
	if e.defaultDeny {
		return Decision{
			Allowed:    false,
			PolicyName: "_default_deny",
			Reason:     "no matching allow policy",
		}
	}

	return Decision{Allowed: true, PolicyName: "_default_allow", Reason: "default allow"}
}

// Middleware creates an HTTP middleware that enforces policies.
func (e *Engine) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity := auth.GetIdentity(r.Context())
			if identity == nil {
				// No identity = anonymous; still evaluate policies
				identity = &auth.Identity{User: "_anonymous"}
			}

			decision := e.Evaluate(r, identity)

			// Store decision in context for audit logging
			ctx := context.WithValue(r.Context(), DecisionKey, &decision)
			r = r.WithContext(ctx)

			if !decision.Allowed {
				e.logger.Warn("request denied by policy",
					"user", identity.User,
					"team", identity.Team,
					"path", r.URL.Path,
					"method", r.Method,
					"policy", decision.PolicyName,
					"reason", decision.Reason,
				)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error":   "forbidden",
					"message": decision.Reason,
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (e *Engine) hasExplicitAllow(reqPath, method string, identity *auth.Identity) bool {
	for _, pol := range e.policies {
		if !e.matchesIdentity(pol, identity) {
			continue
		}
		for _, ap := range pol.AllowPaths {
			if matchPath(ap, reqPath) {
				if len(pol.AllowMethods) == 0 || containsMethod(pol.AllowMethods, method) {
					return true
				}
			}
		}
	}
	return false
}

func (e *Engine) matchesIdentity(pol config.Policy, identity *auth.Identity) bool {
	if identity == nil {
		return false
	}

	// If policy has no user/team restrictions, it applies to everyone
	if len(pol.Users) == 0 && len(pol.Teams) == 0 {
		return true
	}

	for _, u := range pol.Users {
		if u == "*" || strings.EqualFold(u, identity.User) {
			return true
		}
	}
	for _, t := range pol.Teams {
		if t == "*" || strings.EqualFold(t, identity.Team) {
			return true
		}
	}
	return false
}

// matchPath checks if a request path matches a pattern.
// Supports glob-style matching with * wildcards.
func matchPath(pattern, reqPath string) bool {
	matched, _ := path.Match(pattern, reqPath)
	if matched {
		return true
	}

	// Support /api/admin/* matching /api/admin/users etc.
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		if strings.HasPrefix(reqPath, prefix+"/") || reqPath == prefix {
			return true
		}
	}

	return pattern == reqPath
}

func containsMethod(methods []string, method string) bool {
	for _, m := range methods {
		if strings.EqualFold(m, method) {
			return true
		}
	}
	return false
}
