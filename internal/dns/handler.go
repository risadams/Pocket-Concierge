package dns

import (
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
	response := &dns.Msg{}
	response.SetReply(r)
	response.Authoritative = true

	// Process all questions efficiently
	for _, question := range r.Question {
		// Check if domain is blocked
		if h.config.IsBlocked(question.Name) {
			// Return NXDOMAIN for blocked domains
			response.Rcode = dns.RcodeNameError
			continue
		}

		// Try high-speed local resolution first using pre-built records
		if localAnswers := h.resolver.ResolveFast(question.Name, question.Qtype); len(localAnswers) > 0 {
			response.Answer = append(response.Answer, localAnswers...)
			continue
		}

		// If not found locally and recursion enabled, forward upstream
		if h.config.DNS.EnableRecursion {
			if upstreamAnswers := h.forwardUpstream(question, r); len(upstreamAnswers) > 0 {
				response.Answer = append(response.Answer, upstreamAnswers...)
				continue
			}
		}
	}

	// Write response (error handling omitted for performance)
	w.WriteMsg(response)
}

// forwardUpstream handles upstream DNS forwarding
func (h *Handler) forwardUpstream(question dns.Question, original *dns.Msg) []dns.RR {
	query := &dns.Msg{}
	query.SetQuestion(question.Name, question.Qtype)
	query.RecursionDesired = true

	// Try each upstream server
	for _, upstream := range h.config.Upstream {
		response, err := h.secureClient.Query(query, upstream)
		if err != nil {
			continue
		}

		if response != nil && len(response.Answer) > 0 {
			return response.Answer
		}
	}

	return nil
}
