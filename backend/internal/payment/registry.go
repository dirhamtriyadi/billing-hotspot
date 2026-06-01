package payment

import (
	"sync"

	"github.com/dirhamt/billing-hotspot/backend/internal/apperror"
	"github.com/dirhamt/billing-hotspot/backend/internal/config"
)

// Registry resolves payment providers by name. It is safe for concurrent use and
// can be hot-reloaded when an operator edits credentials from the admin UI.
type Registry struct {
	mu       sync.RWMutex
	gateways map[string]Gateway
}

// NewRegistry constructs every supported provider from configuration. Providers
// without credentials are still registered but report Configured() == false.
func NewRegistry(cfg config.PaymentConfig) *Registry {
	return &Registry{gateways: buildGateways(cfg)}
}

// buildGateways instantiates the provider map from a config snapshot.
func buildGateways(cfg config.PaymentConfig) map[string]Gateway {
	list := []Gateway{
		NewXendit(cfg.Xendit),
		NewMidtrans(cfg.Midtrans),
		NewTripay(cfg.Tripay),
	}
	m := make(map[string]Gateway, len(list))
	for _, g := range list {
		m[g.Name()] = g
	}
	return m
}

// Reload rebuilds every provider from a fresh config snapshot. Called when the
// operator updates gateway credentials so changes take effect without a restart.
func (r *Registry) Reload(cfg config.PaymentConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.gateways = buildGateways(cfg)
}

// Get returns a configured provider or an AppError explaining why it cannot be
// used (unknown provider, or missing credentials).
func (r *Registry) Get(name string) (Gateway, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	g, ok := r.gateways[name]
	if !ok {
		return nil, apperror.BadRequest("Unsupported payment provider: " + name)
	}
	if !g.Configured() {
		return nil, apperror.Unprocessable("Payment provider '" + name + "' is not configured on the server")
	}
	return g, nil
}

// Available returns the names of providers that are configured and ready.
func (r *Registry) Available() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.gateways))
	for name, g := range r.gateways {
		if g.Configured() {
			out = append(out, name)
		}
	}
	return out
}
