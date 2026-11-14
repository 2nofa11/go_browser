package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseURL(t *testing.T) {
	tests := []struct {
		in      string
		want    URL
		wantErr bool
	}{
		{in: "http://example.org/index.html", want: URL{"http", "example.org", 80, "/index.html", nil}},
		{in: "http://example.org", want: URL{"http", "example.org", 80, "/", nil}},
		{in: "https://example.org/a/b", want: URL{"https", "example.org", 443, "/a/b", nil}},
		{in: "http://example.org:8080/index.html", want: URL{"http", "example.org", 8080, "/index.html", nil}},
		{in: "ftp://example.org/", wantErr: true},
		{in: "example.org", wantErr: true},
		{in: "http://example.org:abc/", wantErr: true},
	}
	for _, tt := range tests {
		got, err := ParseURL(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseURL(%q): want error, got %+v", tt.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseURL(%q): unexpected error: %v", tt.in, err)
			continue
		}
		if *got != tt.want {
			t.Errorf("ParseURL(%q) = %+v, want %+v", tt.in, *got, tt.want)
		}
	}
}

func TestShow(t *testing.T) {
	var sb strings.Builder
	Show(&sb, "<html><body>こんにちは <b>world</b></body></html>")
	if got, want := sb.String(), "こんにちは world"; got != want {
		t.Errorf("Show = %q, want %q", got, want)
	}
}

const page = "<html><body>\n  This is a simple\n  web page with some\n  text in it.\n</body></html>\n"

func handler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/index.html" {
			http.NotFound(w, r)
			return
		}
		if r.Proto != "HTTP/1.0" {
			t.Errorf("proto = %q, want HTTP/1.0", r.Proto)
		}
		fmt.Fprint(w, page)
	}
}

func TestRequestHTTP(t *testing.T) {
	srv := httptest.NewServer(handler(t))
	defer srv.Close()

	u, err := ParseURL(srv.URL + "/index.html") // http://127.0.0.1:<port> のカスタムポート
	if err != nil {
		t.Fatal(err)
	}
	body, err := u.Request()
	if err != nil {
		t.Fatal(err)
	}
	if body != page {
		t.Errorf("body = %q, want %q", body, page)
	}
}

func TestRequestHTTPSAndLoad(t *testing.T) {
	srv := httptest.NewTLSServer(handler(t))
	defer srv.Close()

	u, err := ParseURL(srv.URL + "/index.html")
	if err != nil {
		t.Fatal(err)
	}
	pool := x509.NewCertPool()
	pool.AddCert(srv.Certificate())
	u.TLSConfig = &tls.Config{RootCAs: pool}

	var sb strings.Builder
	if err := Load(&sb, u); err != nil {
		t.Fatal(err)
	}
	want := "\n  This is a simple\n  web page with some\n  text in it.\n\n"
	if sb.String() != want {
		t.Errorf("Load output = %q, want %q", sb.String(), want)
	}
}

func TestRequestRejectsUntrustedCert(t *testing.T) {
	srv := httptest.NewTLSServer(handler(t))
	defer srv.Close()

	u, _ := ParseURL(srv.URL + "/index.html") // TLSConfig なし → 自己署名証明書は拒否される
	if _, err := u.Request(); err == nil {
		t.Error("want certificate error, got nil")
	}
}
