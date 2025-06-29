package dns

import (
	"strings"

	"github.com/risadams/Pocket-Concierge/internal/config"
)

// Resolver handles local hostname resolution
type Resolver struct {
	config *config.Config
}

// NewResolver creates a new resolver
func NewResolver(cfg *config.Config) *Resolver {
	return &Resolver{
		config: cfg,
	}
}

// ResolveLocal attempts to resolve a hostname using local configuration
func (r *Resolver) ResolveLocal(hostname string) (*config.HostEntry, bool) {
	// Normalize hostname
	hostname = strings.ToLower(strings.TrimSpace(hostname))

	// Direct lookup
	if host, found := r.config.GetHostByName(hostname); found {
		return host, true
	}

	// Try with .home suffix if not present
	if !strings.HasSuffix(hostname, ".home") {
		if host, found := r.config.GetHostByName(hostname + ".home"); found {
			return host, true
		}
	}

	// Try without .home suffix if present
	if strings.HasSuffix(hostname, ".home") {
		baseHostname := strings.TrimSuffix(hostname, ".home")
		if host, found := r.config.GetHostByName(baseHostname); found {
			return host, true
		}
	}

	return nil, false
}

// GetAllHosts returns all configured hosts
func (r *Resolver) GetAllHosts() []config.HostEntry {
	return r.config.Hosts
}

// AddHost adds a new host entry (for future dynamic configuration)
func (r *Resolver) AddHost(host config.HostEntry) {
	r.config.Hosts = append(r.config.Hosts, host)
}
