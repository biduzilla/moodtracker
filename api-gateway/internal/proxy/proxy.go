package proxy

import (
	"fmt"
	"gateway/internal/utils/errors"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type Service struct {
	URL        string
	Timeout    time.Duration
	ErrHandler errors.ErrorHandlerInterface
}

type ReverseProxy struct {
	target  *url.URL
	proxy   *httputil.ReverseProxy
	timeout time.Duration
}

func New(svc Service) (*ReverseProxy, error) {
	target, err := url.Parse(svc.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid service URL %q: %w", svc.URL, err)
	}

	rp := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(target)
			pr.Out.Host = target.Host
		},
	}

	rp.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   svc.Timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ResponseHeaderTimeout: svc.Timeout,
		TLSHandshakeTimeout:   svc.Timeout,
		IdleConnTimeout:       90 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
	}

	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		svc.ErrHandler.ProxyError(w, r, err)
	}

	return &ReverseProxy{
		target:  target,
		proxy:   rp,
		timeout: svc.Timeout,
	}, nil
}

func (rp *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rp.proxy.ServeHTTP(w, r)
}
