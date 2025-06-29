package dns

import (
	"net"
	"strings"
	"sync"

	"github.com/miekg/dns"
	"github.com/risadams/Pocket-Concierge/internal/config"
)

// HostCache provides fast hostname lookups
type HostCache struct {
	hosts   map[string]*config.HostEntry
	records map[string][]dns.RR
	mutex   sync.RWMutex
}

// NewHostCache creates an optimized host cache
func NewHostCache(cfg *config.Config) *HostCache {
	cache := &HostCache{
		hosts:   make(map[string]*config.HostEntry),
		records: make(map[string][]dns.RR),
	}
	cache.Rebuild(cfg)
	return cache
}

// Rebuild updates the cache with current config
func (hc *HostCache) Rebuild(cfg *config.Config) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	// Clear existing caches
	hc.hosts = make(map[string]*config.HostEntry)
	hc.records = make(map[string][]dns.RR)

	homeDomain := cfg.HomeDNSDomain
	if homeDomain == "" {
		homeDomain = "home"
	}

	// Build optimized lookup table
	for i := range cfg.Hosts {
		host := &cfg.Hosts[i]
		normalizedName := strings.ToLower(strings.TrimSuffix(host.Hostname, "."))

		// Determine the full hostname (add domain if not present)
		var fullHostname string
		if strings.Contains(normalizedName, ".") {
			// Already has a domain
			fullHostname = normalizedName
		} else {
			// Add the home domain
			fullHostname = normalizedName + "." + homeDomain
		}

		// Store with multiple variations for fast lookup
		hc.hosts[normalizedName] = host
		hc.hosts[normalizedName+"."] = host
		hc.hosts[fullHostname] = host
		hc.hosts[fullHostname+"."] = host

		// Build DNS records for fast resolution
		hc.buildRecords(normalizedName, fullHostname, host, cfg.DNS.TTL)
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

// buildRecords creates DNS records for a host entry
func (hc *HostCache) buildRecords(shortName, fullHostname string, host *config.HostEntry, ttl int) {
	// Build A records for IPv4 addresses
	for _, ipv4 := range host.IPv4 {
		ip := net.ParseIP(ipv4)
		if ip == nil {
			continue
		}
		rr := &dns.A{
			Hdr: dns.RR_Header{
				Name:   dns.Fqdn(fullHostname),
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    uint32(ttl),
			},
			A: ip.To4(),
		}
		hc.records[strings.ToLower(fullHostname)+":A"] = append(hc.records[strings.ToLower(fullHostname)+":A"], rr)
		hc.records[strings.ToLower(shortName)+":A"] = append(hc.records[strings.ToLower(shortName)+":A"], rr)
	}

	// Build AAAA records for IPv6 addresses
	for _, ipv6 := range host.IPv6 {
		ip := net.ParseIP(ipv6)
		if ip == nil {
			continue
		}
		rr := &dns.AAAA{
			Hdr: dns.RR_Header{
				Name:   dns.Fqdn(fullHostname),
				Rrtype: dns.TypeAAAA,
				Class:  dns.ClassINET,
				Ttl:    uint32(ttl),
			},
			AAAA: ip.To16(),
		}
		hc.records[strings.ToLower(fullHostname)+":AAAA"] = append(hc.records[strings.ToLower(fullHostname)+":AAAA"], rr)
		hc.records[strings.ToLower(shortName)+":AAAA"] = append(hc.records[strings.ToLower(shortName)+":AAAA"], rr)
	}
}

// LookupRecords finds DNS records for a hostname and query type
func (hc *HostCache) LookupRecords(hostname string, qtype uint16) []dns.RR {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()

	normalizedName := strings.ToLower(strings.TrimSuffix(hostname, "."))

	var qtypeStr string
	switch qtype {
	case dns.TypeA:
		qtypeStr = "A"
	case dns.TypeAAAA:
		qtypeStr = "AAAA"
	default:
		return nil
	}

	key := normalizedName + ":" + qtypeStr
	if records, found := hc.records[key]; found {
		return records
	}

	return nil
}
