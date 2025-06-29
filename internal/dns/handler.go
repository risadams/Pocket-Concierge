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
	config       *config.Config
	resolver     *Resolver
	client       *dns.Client
	secureClient *SecureClient
}

// NewHandler creates a new DNS handler
func NewHandler(cfg *config.Config) *Handler {
	return &Handler{
		config:   cfg,
		resolver: NewResolver(cfg),
		client: &dns.Client{
			Timeout: 5 * time.Second,
		},
		secureClient: NewSecureClient(),
	}
}

func (h *Handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	requestStart := time.Now()
	defer func() {
		totalTime := time.Since(requestStart)
		if totalTime > 10*time.Millisecond {
			log.Printf("🐌 SLOW REQUEST: Total time %v", totalTime)
		}
	}()

	response := &dns.Msg{}
	response.SetReply(r)
	response.Authoritative = true

	for _, question := range r.Question {
		questionStart := time.Now()
		log.Printf("🔍 Query: %s %s", dns.TypeToString[question.Qtype], question.Name)

		// Try local resolution first
		localStart := time.Now()
		if localAnswers := h.resolveLocally(question); len(localAnswers) > 0 {
			localTime := time.Since(localStart)
			if localTime > 1*time.Millisecond {
				log.Printf("🐌 SLOW LOCAL: %s took %v", question.Name, localTime)
			}
			response.Answer = append(response.Answer, localAnswers...)
			log.Printf("✅ Local resolve: %s -> %d answers (took %v)", question.Name, len(localAnswers), localTime)
			continue
		}

		// If not found locally and recursion enabled, forward upstream
		if h.config.DNS.EnableRecursion {
			upstreamStart := time.Now()
			if upstreamAnswers := h.forwardUpstream(question, r); len(upstreamAnswers) > 0 {
				upstreamTime := time.Since(upstreamStart)
				response.Answer = append(response.Answer, upstreamAnswers...)
				log.Printf("🔄 Upstream resolve: %s -> %d answers (took %v)", question.Name, len(upstreamAnswers), upstreamTime)
				continue
			}
		}

		questionTime := time.Since(questionStart)
		log.Printf("❌ No resolution: %s (took %v)", question.Name, questionTime)
	}

	writeStart := time.Now()
	if err := w.WriteMsg(response); err != nil {
		log.Printf("❌ Failed to write response: %v", err)
	}
	writeTime := time.Since(writeStart)

	totalTime := time.Since(requestStart)
	if writeTime > 1*time.Millisecond {
		log.Printf("🐌 SLOW WRITE: Response write took %v", writeTime)
	}
	log.Printf("📊 Request completed in %v", totalTime)
}

// resolveLocally attempts to resolve the query using local configuration
func (h *Handler) resolveLocally(question dns.Question) []dns.RR {
	start := time.Now()
	defer func() {
		if time.Since(start) > 5*time.Millisecond {
			log.Printf("⚠️ Slow local resolve: %s took %v", question.Name, time.Since(start))
		}
	}()

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

// Update forwardUpstream method
func (h *Handler) forwardUpstream(question dns.Question, original *dns.Msg) []dns.RR {
	query := &dns.Msg{}
	query.SetQuestion(question.Name, question.Qtype)
	query.RecursionDesired = true

	// Try each upstream server
	for _, upstream := range h.config.Upstream {
		log.Printf("🔒 Trying %s via %s", upstream.Address, upstream.Protocol)

		response, err := h.secureClient.Query(query, upstream)
		if err != nil {
			log.Printf("⚠️ Upstream %s (%s) failed: %v", upstream.Address, upstream.Protocol, err)
			continue
		}

		if response != nil && len(response.Answer) > 0 {
			log.Printf("✅ %s (%s) resolved with %d answers", upstream.Address, upstream.Protocol, len(response.Answer))
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
