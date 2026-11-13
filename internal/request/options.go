package request

// The transport choices a project makes: how long to wait, whether to follow
// redirects, whether to accept a certificate nothing signed, and whether to
// keep the cookies the server sets.

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/yaffalhakim1/endpoint/internal/workspace"
)

// Build returns a client for these options, and the jar that carries cookies
// between calls, which the caller keeps and passes back in.
func Build(options workspace.HTTPOptions, previous *http.Client) (*http.Client, error) {
	transport := &http.Transport{
		// The same settings Go's default transport uses, except the one this
		// app may change.
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          32,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	if options.Proxy != "" {
		proxy, err := url.Parse(options.Proxy)
		if err != nil || proxy.Scheme == "" || proxy.Host == "" {
			return nil, fmt.Errorf("the proxy address needs a scheme and a host, as http://127.0.0.1:8080")
		}
		transport.Proxy = http.ProxyURL(proxy)
	}

	if options.InsecureTLS {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client := &http.Client{Transport: transport}

	if options.TimeoutSeconds > 0 {
		client.Timeout = time.Duration(options.TimeoutSeconds) * time.Second
	}

	if options.NoRedirects {
		client.CheckRedirect = func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	if !options.NoCookies {
		jar := jarFrom(previous)
		if jar == nil {
			created, err := cookiejar.New(nil)
			if err != nil {
				return nil, err
			}
			jar = created
		}
		client.Jar = jar
	}

	return client, nil
}

// jarFrom keeps the cookies of the client being replaced, so changing a
// timeout does not sign you out.
func jarFrom(previous *http.Client) http.CookieJar {
	if previous == nil || previous.Jar == nil {
		return nil
	}
	return previous.Jar
}

// WithoutCookies sends a call with no jar at all, for a request that must not
// keep state.
func WithoutCookies(client *http.Client) *http.Client {
	if client == nil {
		return &http.Client{}
	}
	out := *client
	out.Jar = nil
	return &out
}
