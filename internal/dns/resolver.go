package dns

import (
	"log"
	"time"

	"github.com/risadams/Pocket-Concierge/internal/config"
)

// Resolver handles local hostname resolution
type Resolver struct {
	config    *config.Config
	hostCache *HostCache
}

// NewResolver creates a new resolver
func NewResolver(cfg *config.Config) *Resolver {
	return &Resolver{
		config:    cfg,
		hostCache: NewHostCache(cfg),
	}
}

// ResolveLocal attempts to resolve a hostname using cached lookup
func (r *Resolver) ResolveLocal(hostname string) (*config.HostEntry, bool) {
	// Add timing check
	start := time.Now()
	defer func() {
		if time.Since(start) > 100*time.Microsecond {
			log.Printf("🐌 ResolveLocal slow: %s took %v", hostname, time.Since(start))
		}
	}()

	return r.hostCache.Lookup(hostname)
}

// GetAllHosts returns all configured hosts
func (r *Resolver) GetAllHosts() []config.HostEntry {
	return r.config.Hosts
}

// AddHost adds a new host entry (for future dynamic configuration)
func (r *Resolver) AddHost(host config.HostEntry) {
	r.config.Hosts = append(r.config.Hosts, host)
}
