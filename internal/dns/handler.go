package dns

import (
	"log"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/risadams/Pocket-Concierge/internal/config"
)

// Handler manages DNS request processing
type Handler struct {
	config   *config.Config
	resolver *Resolver
	client   *dns.Client
}

// NewHandler creates a new DNS handler
func NewHandler(cfg *config.Config) *Handler {
	return &Handler{
		config:   cfg,
		resolver: NewResolver(cfg),
		client: &dns.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// ServeDNS handles incoming DNS queries (implements dns.Handler interface)
func (h *Handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	// Create response message
	response := &dns.Msg{}
	response.SetReply(r)
	response.Authoritative = false
	response.RecursionAvailable = h.config.DNS.EnableRecursion

	// Process each question
	for _, question := range r.Question {
		log.Printf("🔍 Query: %s %s", dns.TypeToString[question.Qtype], question.Name)

		// Try local resolution first
		if answers := h.resolveLocally(question); len(answers) > 0 {
			response.Answer = append(response.Answer, answers...)
			response.Authoritative = true
			log.Printf("✅ Local resolve: %s -> %d answers", question.Name, len(answers))
		} else if h.config.DNS.EnableRecursion {
			// Forward to upstream DNS
			if answers := h.forwardUpstream(question, r); len(answers) > 0 {
				response.Answer = append(response.Answer, answers...)
				log.Printf("🔄 Upstream resolve: %s -> %d answers", question.Name, len(answers))
			} else {
				log.Printf("❌ No resolution: %s", question.Name)
			}
		} else {
			log.Printf("🚫 Recursion disabled: %s", question.Name)
		}
	}

	// Send response
	if err := w.WriteMsg(response); err != nil {
		log.Printf("❌ Failed to write response: %v", err)
	}
}

// resolveLocally attempts to resolve the query using local configuration
func (h *Handler) resolveLocally(question dns.Question) []dns.RR {
	var answers []dns.RR

	// Remove trailing dot and convert to lowercase
	hostname := strings.ToLower(strings.TrimSuffix(question.Name, "."))

	// Look up in local hosts
	if host, found := h.resolver.ResolveLocal(hostname); found {
		switch question.Qtype {
		case dns.TypeA:
			// IPv4 addresses
			for _, ip := range host.IPv4 {
				if record := h.createARecord(question.Name, ip); record != nil {
					answers = append(answers, record)
				}
			}
		case dns.TypeAAAA:
			// IPv6 addresses
			for _, ip := range host.IPv6 {
				if record := h.createAAAARecord(question.Name, ip); record != nil {
					answers = append(answers, record)
				}
			}
		case dns.TypeANY:
			// Both IPv4 and IPv6
			for _, ip := range host.IPv4 {
				if record := h.createARecord(question.Name, ip); record != nil {
					answers = append(answers, record)
				}
			}
			for _, ip := range host.IPv6 {
				if record := h.createAAAARecord(question.Name, ip); record != nil {
					answers = append(answers, record)
				}
			}
		}
	}

	return answers
}

// forwardUpstream forwards the query to upstream DNS servers
func (h *Handler) forwardUpstream(question dns.Question, original *dns.Msg) []dns.RR {
	// Create a new query with just this question
	query := &dns.Msg{}
	query.SetQuestion(question.Name, question.Qtype)
	query.RecursionDesired = true

	// Try each upstream server
	for _, upstream := range h.config.Upstream {
		response, _, err := h.client.Exchange(query, upstream)
		if err != nil {
			log.Printf("⚠️ Upstream %s failed: %v", upstream, err)
			continue
		}

		if response != nil && len(response.Answer) > 0 {
			return response.Answer
		}
	}

	return nil
}

// createARecord creates an A record for IPv4
func (h *Handler) createARecord(name, ip string) dns.RR {
	if net.ParseIP(ip) == nil {
		log.Printf("⚠️ Invalid IPv4 address: %s", ip)
		return nil
	}

	record := &dns.A{
		Hdr: dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    uint32(h.config.DNS.TTL),
		},
		A: net.ParseIP(ip),
	}
	return record
}

// createAAAARecord creates an AAAA record for IPv6
func (h *Handler) createAAAARecord(name, ip string) dns.RR {
	if net.ParseIP(ip) == nil {
		log.Printf("⚠️ Invalid IPv6 address: %s", ip)
		return nil
	}

	record := &dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    uint32(h.config.DNS.TTL),
		},
		AAAA: net.ParseIP(ip),
	}
	return record
}
