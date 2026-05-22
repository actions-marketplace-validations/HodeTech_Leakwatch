// Package httpx provides a shared, security-hardened HTTP client and helpers
// for use by secret verifiers.
//
// It is intentionally placed under internal/verifier/internal so that it can
// only be imported by packages within internal/verifier.
//
// Security rationale:
//
//   - Verifiers send provider credentials in custom headers (for example
//     x-api-key, PRIVATE-TOKEN, DD-API-KEY) or embedded in the request URL
//     (for example Telegram and Infura). On a cross-domain 3xx redirect, the
//     Go standard library strips the Authorization header but NOT custom
//     headers, and it re-sends the full URL — which would leak the credential
//     to an attacker-controlled redirect target. To prevent this, the shared
//     client does NOT follow redirects: it returns the 3xx response so the
//     verifier can decide how to map it (see IsRedirect).
//
//   - Response bodies are read through a bounded reader (LimitReader) so a
//     malicious or misbehaving endpoint cannot exhaust memory.
//
//   - The client sets an explicit Timeout as a hard ceiling, in addition to
//     the per-request context deadline applied by the verification engine.
//
// This helper deliberately does NOT implement retry, backoff, or per-provider
// rate limiting. Those concerns are handled (or deferred) elsewhere; keeping
// this package focused on transport safety.
package httpx

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// MaxBodyBytes is the maximum number of response-body bytes a verifier reads.
// It caps memory usage when decoding provider responses. 1 MiB is far larger
// than any legitimate verification response.
const MaxBodyBytes int64 = 1 << 20

// DefaultTimeout is the hard ceiling applied to a single verification request.
// The verification engine also applies a per-request context deadline; this
// timeout guards against a missing or overly generous deadline.
const DefaultTimeout = 30 * time.Second

var (
	clientOnce   sync.Once
	sharedClient *http.Client
)

// noRedirect instructs the HTTP client to return the most recent response
// (the 3xx) instead of following the redirect. This prevents credentials in
// custom headers or in the request URL from being re-sent to a redirect
// target, which the standard library would otherwise do for non-Authorization
// headers.
func noRedirect(_ *http.Request, _ []*http.Request) error {
	return http.ErrUseLastResponse
}

// Client returns the shared, security-hardened HTTP client.
//
// The returned client is safe for concurrent use and is shared across all
// verifiers. Callers MUST NOT mutate it. Tests that need to point a verifier
// at a stub server should inject their own *http.Client through the verifier's
// test seam instead of mutating this client.
func Client() *http.Client {
	clientOnce.Do(func() {
		// Clone the default transport so we benefit from connection pooling
		// and environment proxy settings without sharing mutable state with
		// http.DefaultTransport.
		transport := http.DefaultTransport
		if dt, ok := http.DefaultTransport.(*http.Transport); ok {
			transport = dt.Clone()
		}
		sharedClient = &http.Client{
			Transport:     transport,
			CheckRedirect: noRedirect,
			Timeout:       DefaultTimeout,
		}
	})
	return sharedClient
}

// LimitReader wraps r so that at most MaxBodyBytes are read from it. Verifiers
// should decode response bodies through this reader (for example
// json.NewDecoder(httpx.LimitReader(resp.Body))) to bound memory usage.
func LimitReader(r io.Reader) io.Reader {
	return io.LimitReader(r, MaxBodyBytes)
}

// IsRedirect reports whether the given HTTP status code is a 3xx redirect.
//
// Because Client does not follow redirects, verifiers observe 3xx responses
// directly. A redirect from an API endpoint generally means the credential
// context is wrong (for example a wrong host or a login redirect), so it should
// NOT be treated as a successful verification.
func IsRedirect(statusCode int) bool {
	return statusCode >= 300 && statusCode < 400
}

// EnsureHTTPS returns rawURL unchanged when it already uses the https scheme.
// Otherwise it returns ok=false. It is intended for URLs derived from
// untrusted context (for example a host taken from detector ExtraData) to make
// sure credentials are only ever sent over TLS.
func EnsureHTTPS(rawURL string) (string, bool) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", false
	}
	if !strings.EqualFold(u.Scheme, "https") {
		return "", false
	}
	return rawURL, true
}
