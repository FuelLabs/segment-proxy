package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type SegmentServer int

const (
	CDN SegmentServer = iota
	TrackingAPI
)

func TestSegmentReverseProxy(t *testing.T) {
	// Test URL prefix stripping
	urlPrefix := "/prefix"
	// Set the environment variable for the test
	strippedPath := "/v1/projects"
	called := false
	cdn := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.URL.Path != strippedPath {
			t.Errorf("Expected path %q, got %q", strippedPath, r.URL.Path)
		}
		fmt.Fprintln(w, "Hello, client")
	}))
	trackingAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("Tracking API should not be called for this test")
	}))
	proxy := httptest.NewServer(NewSegmentReverseProxy(mustParseUrl(cdn.URL), mustParseUrl(trackingAPI.URL), urlPrefix))
	resp, err := http.Get(proxy.URL + urlPrefix + strippedPath)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if !called {
		t.Errorf("CDN server was not called")
	}
	cdn.Close()
	trackingAPI.Close()
	cases := []struct {
		url            string
		expectedServer SegmentServer
	}{
		{"/v1/projects", CDN},
		{"/analytics.js/v1", CDN},
		{"/v1/import", TrackingAPI},
		{"/v1/pixel", TrackingAPI},
	}
	for _, c := range cases {
		cdn := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if c.expectedServer == CDN {
				fmt.Fprintln(w, "Hello, client")
			} else {
				t.Errorf("CDN unexpected request: %v\n", r.URL)
			}
		}))

		trackingAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if c.expectedServer == TrackingAPI {
				fmt.Fprintln(w, "Hello, client")
			} else {
				t.Errorf("Tracking API unexpected request: %v\n", r.URL)
			}
		}))

		proxy := httptest.NewServer(NewSegmentReverseProxy(mustParseUrl(cdn.URL), mustParseUrl(trackingAPI.URL), ""))

		_, err := http.Get(proxy.URL + c.url)
		if err != nil {
			t.Fatal(err)
		}

		cdn.Close()
		trackingAPI.Close()
	}
}

func TestHealthEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/_health", healthHandler)
	server := httptest.NewServer(mux)
	defer server.Close()
	resp, err := http.Get(server.URL + "/_health")
	if err != nil {
		t.Fatalf("Failed to GET /_health: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	buf := make([]byte, 16)
	n, _ := resp.Body.Read(buf)
	body := string(buf[:n])
	if body != "ready" {
		t.Errorf("Expected body 'ready', got %q", body)
	}
}

func mustParseUrl(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		log.Fatal(err)
	}
	return u
}
