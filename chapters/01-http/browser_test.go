package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseURL(t *testing.T) {
	tests := []struct {
		in      string
		want    URL
		wantErr bool
	}{
		{in: "http://example.org/index.html", want: URL{Scheme: "http", Host: "example.org", Port: 80, Path: "/index.html"}},
		{in: "http://example.org", want: URL{Scheme: "http", Host: "example.org", Port: 80, Path: "/"}},
		{in: "https://example.org/a/b", want: URL{Scheme: "https", Host: "example.org", Port: 443, Path: "/a/b"}},
		{in: "http://example.org:8080/index.html", want: URL{Scheme: "http", Host: "example.org", Port: 8080, Path: "/index.html"}},
		{in: "file:///path/goes/here", want: URL{Scheme: "file", Path: "/path/goes/here"}},
		{in: "file://relative/path", wantErr: true},
		{in: "view-source:http://example.org/", want: URL{Scheme: "http", Host: "example.org", Port: 80, Path: "/", ViewSource: true}},
		{in: "view-source:data:text/html,hi", want: URL{Scheme: "data", Data: "hi", ViewSource: true}},
		{in: "view-source:ftp://example.org/", wantErr: true}, // 中身のURLが不正なら、全体も不正
		{in: "data:text/html,Hello world!", want: URL{Scheme: "data", Data: "Hello world!"}},
		{in: "data:text/html,a,b", want: URL{Scheme: "data", Data: "a,b"}}, // 最初の "," だけで区切る
		{in: "data:text/html", wantErr: true},                              // "," がない
		{in: "data:text/html;base64,SGVsbG8=", wantErr: true},              // base64 は未対応
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

func TestShowEntities(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"&lt;div&gt;", "<div>"},
		{"<p>&lt;b&gt;bold&lt;/b&gt;</p>", "<b>bold</b>"}, // 本物のタグは消え、エンティティは文字になる
		{"1 &lt; 2 &gt; 0", "1 < 2 > 0"},
		{"&lt;&lt;", "<<"},
		{"&amp;lt;", "&lt;"}, // 1回だけ変換する。&amp; → & の結果を、さらに &lt; として読み直さない
		{"a &amp; b", "a & b"},
	}
	for _, tt := range tests {
		var sb strings.Builder
		Show(&sb, tt.in)
		if got := sb.String(); got != tt.want {
			t.Errorf("Show(%q) = %q, want %q", tt.in, got, tt.want)
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
		if r.Proto != "HTTP/1.1" {
			t.Errorf("proto = %q, want HTTP/1.1", r.Proto)
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

func TestRequestFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index.html")
	if err := os.WriteFile(path, []byte(page), 0o644); err != nil {
		t.Fatal(err)
	}

	u, err := ParseURL("file://" + path)
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

func TestLoadData(t *testing.T) {
	u, err := ParseURL("data:text/html,<b>hi</b> there")
	if err != nil {
		t.Fatal(err)
	}
	var sb strings.Builder
	if err := Load(&sb, u); err != nil {
		t.Fatal(err)
	}
	if got, want := sb.String(), "hi there"; got != want {
		t.Errorf("Load output = %q, want %q", got, want)
	}
}

func TestLoadViewSource(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"view-source:data:text/html,<b>hi</b>", "<b>hi</b>"},       // タグも含めて、そのまま見える
		{"view-source:data:text/html,&lt;p&gt;", "&lt;p&gt;"},       // 元のエンティティも、元のまま見える
		{"data:text/html,<b>hi</b>", "hi"},                          // view-source なしなら、今までどおり
		{"view-source:data:text/html,<p>a & b</p>", "<p>a & b</p>"}, // & もそのまま
	}
	for _, tt := range tests {
		u, err := ParseURL(tt.in)
		if err != nil {
			t.Fatal(err)
		}
		var sb strings.Builder
		if err := Load(&sb, u); err != nil {
			t.Fatal(err)
		}
		if got := sb.String(); got != tt.want {
			t.Errorf("Load(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestDefaultURL(t *testing.T) {
	u, err := ParseURL(defaultURL())
	if err != nil {
		t.Fatal(err)
	}
	body, err := u.Request()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, "Hello, go_browser!") {
		t.Errorf("default page body = %q", body)
	}
}

func TestRequestFileNotFound(t *testing.T) {
	u, err := ParseURL("file://" + filepath.Join(t.TempDir(), "missing.html"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := u.Request(); err == nil {
		t.Error("want error for a missing file, got nil")
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
