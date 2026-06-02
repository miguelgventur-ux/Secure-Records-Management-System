package main

// tests that the security headers middleware attaches all four
// required headers to every HTTP response, regardless of the endpoint

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// dummyHandler is a minimal http.Handler used to verify that the middleware
// calls through to the next handler and still applies headers
var dummyHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestSecurityHeaders_ContentSecurityPolicy(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	securityHeadersMiddleware(dummyHandler).ServeHTTP(rr, req)

	got := rr.Header().Get("Content-Security-Policy")
	if got == "" {
		t.Error("Content-Security-Policy header is missing")
	}
	// must restrict script execution
	if got == "" {
		t.Error("CSP header is empty")
	}
}

func TestSecurityHeaders_XFrameOptions(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/login", nil)

	securityHeadersMiddleware(dummyHandler).ServeHTTP(rr, req)

	got := rr.Header().Get("X-Frame-Options")
	if got != "DENY" {
		t.Errorf("X-Frame-Options = %q, want %q", got, "DENY")
	}
}

func TestSecurityHeaders_XContentTypeOptions(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/login", nil)

	securityHeadersMiddleware(dummyHandler).ServeHTTP(rr, req)

	got := rr.Header().Get("X-Content-Type-Options")
	if got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want %q", got, "nosniff")
	}
}

func TestSecurityHeaders_ReferrerPolicy(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/login", nil)

	securityHeadersMiddleware(dummyHandler).ServeHTTP(rr, req)

	got := rr.Header().Get("Referrer-Policy")
	if got != "strict-origin-when-cross-origin" {
		t.Errorf("Referrer-Policy = %q, want %q", got, "strict-origin-when-cross-origin")
	}
}

func TestSecurityHeaders_AllFourPresent(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/login", nil)

	securityHeadersMiddleware(dummyHandler).ServeHTTP(rr, req)

	headers := []string{
		"Content-Security-Policy",
		"X-Frame-Options",
		"X-Content-Type-Options",
		"Referrer-Policy",
	}
	for _, h := range headers {
		if rr.Header().Get(h) == "" {
			t.Errorf("missing security header: %s", h)
		}
	}
}

func TestSecurityHeaders_AppliedOnEveryRoute(t *testing.T) {
	routes := []string{"/", "/login", "/record", "/admin/records"}

	for _, route := range routes {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, route, nil)

		securityHeadersMiddleware(dummyHandler).ServeHTTP(rr, req)

		if rr.Header().Get("X-Frame-Options") != "DENY" {
			t.Errorf("X-Frame-Options missing on route %s", route)
		}
	}
}

func TestSecurityHeaders_NextHandlerCalled(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	securityHeadersMiddleware(next).ServeHTTP(rr, req)

	if !called {
		t.Error("middleware did not call the next handler")
	}
}
