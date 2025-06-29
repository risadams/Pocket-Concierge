package dns

import (
	"strings"
	"sync"

	"github.com/risadams/Pocket-Concierge/internal/config"
)

// HostCache provides fast hostname lookups
type HostCache struct {
	hosts map[string]*config.HostEntry
	mutex sync.RWMutex
}

// NewHostCache creates an optimized host cache
func NewHostCache(cfg *config.Config) *HostCache {
	cache := &HostCache{
		hosts: make(map[string]*config.HostEntry),
	}
	cache.Rebuild(cfg)
	return cache
}

// Rebuild updates the cache with current config
func (hc *HostCache) Rebuild(cfg *config.Config) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	// Clear existing cache
	hc.hosts = make(map[string]*config.HostEntry)

	// Build optimized lookup table
	for i := range cfg.Hosts {
		host := &cfg.Hosts[i]
		normalizedName := strings.ToLower(strings.TrimSuffix(host.Hostname, "."))

		// Store with multiple variations for fast lookup
		hc.hosts[normalizedName] = host
		hc.hosts[normalizedName+"."] = host // With trailing dot

		// Add .home suffix variations
		if !strings.HasSuffix(normalizedName, ".home") {
			hc.hosts[normalizedName+".home"] = host
			hc.hosts[normalizedName+".home."] = host
		}

		// Add without .home suffix variations
		if strings.HasSuffix(normalizedName, ".home") {
			baseHostname := strings.TrimSuffix(normalizedName, ".home")
			hc.hosts[baseHostname] = host
			hc.hosts[baseHostname+"."] = host
		}
	}
}

// Lookup finds a host entry quickly
func (hc *HostCache) Lookup(hostname string) (*config.HostEntry, bool) {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()

	normalizedName := strings.ToLower(strings.TrimSpace(hostname))
	host, found := hc.hosts[normalizedName]
	return host, found
}
