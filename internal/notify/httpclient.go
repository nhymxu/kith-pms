package notify

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/netip"
	"syscall"
	"time"
)

const (
	attemptTimeout = 10 * time.Second
	maxDrainBytes  = 64 << 10
)

// HTTPClient is the only outbound HTTP path for channels. It never follows
// redirects, bypasses env proxies, and vets the resolved IP at dial time (which
// also covers DNS rebinding). Loopback and private ranges are allowed so
// self-hosted LAN targets work; link-local (cloud metadata), multicast and
// unspecified addresses are not. Plain http is only dialable to loopback/private.
type HTTPClient struct {
	secure *http.Client
	plain  *http.Client
}

func NewHTTPClient() *HTTPClient {
	return &HTTPClient{secure: newClient(true), plain: newClient(false)}
}

func newClient(allowPublic bool) *http.Client {
	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
		Control: func(_, address string, _ syscall.RawConn) error { return checkAddress(address, allowPublic) },
	}

	return &http.Client{
		Timeout: attemptTimeout,
		Transport: &http.Transport{
			Proxy:               nil,
			DialContext:         dialer.DialContext,
			TLSHandshakeTimeout: 5 * time.Second,
			DisableKeepAlives:   true,
		},
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}

// Beyond link-local: AWS IPv6 IMDS sits in ULA space, and NAT64 can embed 169.254.0.0/16.
var deniedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("fd00:ec2::254/128"),
	netip.MustParsePrefix("64:ff9b::/96"),
}

func checkAddress(address string, allowPublic bool) error {
	ap, err := netip.ParseAddrPort(address)
	if err != nil {
		return errBlockedAddress
	}

	ip := ap.Addr().Unmap()
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsInterfaceLocalMulticast() ||
		ip.IsMulticast() || ip.IsUnspecified() {
		return errBlockedAddress
	}

	for _, p := range deniedPrefixes {
		if p.Contains(ip) {
			return errBlockedAddress
		}
	}

	if !allowPublic && !ip.IsLoopback() && !ip.IsPrivate() {
		return errPlainHTTPPublic
	}

	return nil
}

// Do sends req and returns nil for a 2xx response, otherwise a *SendError.
// The response body is drained and discarded; it never reaches an error.
func (c *HTTPClient) Do(ctx context.Context, req *http.Request) error {
	client := c.plain // fail closed: only an exact https scheme may reach public hosts
	if req.URL.Scheme == "https" {
		client = c.secure
	}

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return Classify(err)
	}
	defer func() { _ = resp.Body.Close() }()

	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxDrainBytes))

	// explicit nil check: returning a typed-nil *SendError as error would be non-nil
	if se := ClassifyStatus(resp.StatusCode, resp.Header); se != nil { //nolint:revive
		return se
	}

	return nil
}
