package dns

import (
	"github.com/miekg/dns"
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

// ResolveFast attempts to resolve a hostname using pre-built DNS records
func (r *Resolver) ResolveFast(hostname string, qtype uint16) []dns.RR {
	return r.hostCache.LookupRecords(hostname, qtype)
}

// ResolveLocal attempts to resolve a hostname using cached lookup (legacy compatibility)
func (r *Resolver) ResolveLocal(hostname string) (*config.HostEntry, bool) {
	return r.hostCache.Lookup(hostname)
}

// GetAllHosts returns all configured hosts
func (r *Resolver) GetAllHosts() []config.HostEntry {
	return r.config.Hosts
}

// AddHost adds a new host entry (for future dynamic configuration)
func (r *Resolver) AddHost(host config.HostEntry) {
	r.config.Hosts = append(r.config.Hosts, host)
	r.hostCache.Rebuild(r.config) // Rebuild cache when adding hosts
}
