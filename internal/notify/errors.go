package notify

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"
)

// Fixed error classes. Only these strings (plus an HTTP status) are ever
// persisted or logged, so adapter-originated text cannot leak a secret.
const (
	ClassTimeout       = "timeout"
	ClassDNS           = "dns"
	ClassTLS           = "tls"
	ClassHTTP4xx       = "http_4xx"
	ClassHTTP5xx       = "http_5xx"
	ClassRateLimited   = "rate_limited"
	ClassInvalidConfig = "invalid_config"
	ClassInternal      = "internal"
	// ClassGone means the destination no longer exists (push subscription expired); the channel is removed.
	ClassGone = "gone"
)

const maxRetryAfter = time.Hour

var knownClasses = map[string]bool{
	ClassTimeout: true, ClassDNS: true, ClassTLS: true, ClassHTTP4xx: true,
	ClassHTTP5xx: true, ClassRateLimited: true, ClassInvalidConfig: true, ClassInternal: true,
	ClassGone: true,
}

// SendError is the only error shape a channel may return from Send.
type SendError struct {
	Class      string
	Status     int
	Permanent  bool
	RetryAfter time.Duration
}

func (e *SendError) Error() string {
	if e.Status != 0 {
		return e.Class + " (" + strconv.Itoa(e.Status) + ")"
	}

	return e.Class
}

// Classify maps a transport error to a fixed class without keeping its text.
func Classify(err error) *SendError {
	var (
		dnsErr  *net.DNSError
		netErr  net.Error
		unknown x509.UnknownAuthorityError
		hostErr x509.HostnameError
		certErr x509.CertificateInvalidError
		recHdr  tls.RecordHeaderError
	)

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return &SendError{Class: ClassTimeout}
	case errors.As(err, &dnsErr):
		return &SendError{Class: ClassDNS}
	case errors.As(err, &unknown), errors.As(err, &hostErr), errors.As(err, &certErr), errors.As(err, &recHdr):
		return &SendError{Class: ClassTLS, Permanent: true}
	case errors.As(err, &netErr) && netErr.Timeout():
		return &SendError{Class: ClassTimeout}
	case errors.Is(err, errBlockedAddress), errors.Is(err, errPlainHTTPPublic):
		return &SendError{Class: ClassInvalidConfig, Permanent: true}
	}

	return &SendError{Class: ClassInternal}
}

// ClassifyStatus returns nil for 2xx, otherwise the matching SendError.
func ClassifyStatus(status int, h http.Header) *SendError {
	switch {
	case status >= 200 && status < 300:
		return nil
	case status == http.StatusTooManyRequests:
		return &SendError{Class: ClassRateLimited, Status: status, RetryAfter: parseRetryAfter(h.Get("Retry-After"))}
	case status == http.StatusRequestTimeout:
		return &SendError{Class: ClassHTTP4xx, Status: status}
	case status >= 500:
		return &SendError{Class: ClassHTTP5xx, Status: status}
	case status >= 400:
		return &SendError{Class: ClassHTTP4xx, Status: status, Permanent: true}
	default: // 1xx/3xx: redirects are never followed
		return &SendError{Class: ClassInvalidConfig, Status: status, Permanent: true}
	}
}

func parseRetryAfter(v string) time.Duration {
	if secs, err := strconv.Atoi(v); err == nil && secs > 0 {
		if secs >= int(maxRetryAfter/time.Second) { // also avoids Duration overflow
			return maxRetryAfter
		}

		return time.Duration(secs) * time.Second
	}

	if t, err := http.ParseTime(v); err == nil {
		if d := time.Until(t); d > 0 {
			return min(d, maxRetryAfter)
		}
	}

	return 0
}

// SanitizeErr reduces any error from an adapter to a SendError with a fixed
// class. The underlying error text is dropped.
func SanitizeErr(err error) *SendError {
	if err == nil {
		return nil
	}

	var se *SendError
	if errors.As(err, &se) && se == nil { // adapter returned a typed-nil *SendError as success
		return nil
	}

	if errors.As(err, &se) && knownClasses[se.Class] {
		return &SendError{
			Class:      se.Class,
			Status:     se.Status,
			Permanent:  se.Permanent,
			RetryAfter: max(0, min(se.RetryAfter, maxRetryAfter)),
		}
	}

	if errors.As(err, &se) {
		return &SendError{Class: ClassInternal, Permanent: true}
	}

	return Classify(err)
}

var (
	errBlockedAddress  = errors.New("notify: destination address not allowed")
	errPlainHTTPPublic = errors.New("notify: plain http to a public address")
)
