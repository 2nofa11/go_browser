// 第1章: Webページのダウンロード（Downloading Web Pages）
//
// 原文の Python 実装を、net / crypto/tls といった低レベルAPIでGoに書き直したもの。
// net/http は章の学習目的（ソケットから自前でHTTPを話す）を損なうため使わない。
package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// 原文: Telnet in Python（class URL）
type URL struct {
	Scheme string
	Host   string
	Port   int
	Path   string

	// data スキームの内容そのもの。他のスキームでは空。
	Data string

	// nil ならシステムの証明書で検証する。テストで自己署名証明書を信頼するためだけに差し替える。
	TLSConfig *tls.Config
}

// 原文: Telnet in Python / Encrypted Connections（URL.__init__）
func ParseURL(raw string) (*URL, error) {
	// data:<メディアタイプ>,<内容> は "://" を持たないので先に処理する。
	if rest, ok := strings.CutPrefix(raw, "data:"); ok {
		mediaType, data, ok := strings.Cut(rest, ",")
		if !ok {
			return nil, fmt.Errorf("invalid data URL %q: missing ','", raw)
		}
		if strings.HasSuffix(mediaType, ";base64") {
			return nil, fmt.Errorf("base64 data URL is not supported: %q", raw)
		}
		return &URL{Scheme: "data", Data: data}, nil
	}

	scheme, rest, ok := strings.Cut(raw, "://")
	if !ok {
		return nil, fmt.Errorf("invalid URL %q: missing scheme", raw)
	}

	// file:///path/goes/here は、ホストを持たず、"://" の後ろ全体がパスになる。
	if scheme == "file" {
		if !strings.HasPrefix(rest, "/") {
			return nil, fmt.Errorf("invalid file URL %q: path must be absolute (file:///...)", raw)
		}
		return &URL{Scheme: scheme, Path: rest}, nil
	}

	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme %q", scheme)
	}
	u := &URL{Scheme: scheme}

	if scheme == "http" {
		u.Port = 80
	} else {
		u.Port = 443
	}

	if !strings.Contains(rest, "/") {
		rest += "/"
	}
	host, path, _ := strings.Cut(rest, "/")
	u.Host = host
	u.Path = "/" + path

	if h, p, ok := strings.Cut(u.Host, ":"); ok {
		port, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid port %q: %w", p, err)
		}
		u.Host, u.Port = h, port
	}
	return u, nil
}

// 原文: Request and Response（URL.request）
func (u *URL) Request() (string, error) {
	switch u.Scheme {
	case "data":
		return u.Data, nil
	case "file":
		data, err := os.ReadFile(u.Path)
		return string(data), err
	default:
		return u.requestHTTP()
	}
}

// http / https のレスポンスボディを返す。ステータスコードは検査せず、ボディをそのまま返す。
func (u *URL) requestHTTP() (string, error) {
	// net.Dial がソケットの作成と接続をまとめて行う。
	addr := net.JoinHostPort(u.Host, strconv.Itoa(u.Port))
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return "", err
	}

	// 元の接続ではなく、ラップした方を以降で使う。
	if u.Scheme == "https" {
		cfg := &tls.Config{}
		if u.TLSConfig != nil {
			cfg = u.TLSConfig.Clone()
		}
		cfg.ServerName = u.Host
		conn = tls.Client(conn, cfg)
	}
	defer conn.Close()

	// 改行は \r\n で、最後に空行を送る。これを忘れるとサーバーが待ち続ける。
	request := fmt.Sprintf("GET %s HTTP/1.1\r\n", u.Path)
	requestHeaders := map[string]string{
		"Host":       u.Host,
		"Connection": "close",
		"User-Agent": "go-browser",
	}
	for name, value := range requestHeaders {
		request += fmt.Sprintf("%s: %s\r\n", name, value)
	}
	request += "\r\n"
	if _, err := io.WriteString(conn, request); err != nil {
		return "", err
	}

	response := bufio.NewReader(conn)

	statusline, err := response.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading status line: %w", err)
	}
	parts := strings.SplitN(strings.TrimRight(statusline, "\r\n"), " ", 3)
	if len(parts) < 2 {
		return "", fmt.Errorf("malformed status line %q", statusline)
	}
	// バージョン(parts[0])は、設定ミスのサーバーが多いので確認しない。
	// ステータスコード(parts[1])も検査せず、エラーページもそのまま表示する。

	headers := map[string]string{}
	for {
		line, err := response.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("reading headers: %w", err)
		}
		if line == "\r\n" {
			break
		}
		name, value, ok := strings.Cut(line, ":")
		if !ok {
			return "", fmt.Errorf("malformed header line %q", line)
		}
		// ヘッダ名は大文字小文字を区別しないので、小文字に正規化する。
		headers[strings.ToLower(name)] = strings.TrimSpace(value)
	}

	if _, ok := headers["transfer-encoding"]; ok {
		return "", fmt.Errorf("transfer-encoding is not supported")
	}
	if _, ok := headers["content-encoding"]; ok {
		return "", fmt.Errorf("content-encoding is not supported")
	}

	body, err := io.ReadAll(response)
	if err != nil {
		return "", fmt.Errorf("reading body: %w", err)
	}
	return string(body), nil
}

// 原文: Displaying the HTML（show）
func Show(w io.Writer, body string) {
	inTag := false
	for _, c := range body { // range は rune 単位なので、UTF-8 の多バイト文字も壊さない
		switch {
		case c == '<':
			inTag = true
		case c == '>':
			inTag = false
		case !inTag:
			fmt.Fprint(w, string(c))
		}
	}
}

// 原文: Displaying the HTML（load）
func Load(w io.Writer, u *URL) error {
	body, err := u.Request()
	if err != nil {
		return err
	}
	Show(w, body)
	return nil
}

// URLを指定せずに起動したときに開く、手早いテスト用のファイルの URL を返す。
func defaultURL() string {
	// 実行場所に依存しないよう、このソースファイルの位置を基準にする。
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	return "file://" + filepath.Join(filepath.Dir(file), "testdata", "index.html")
}

// 原文: Displaying the HTML（if __name__ == "__main__": load(URL(sys.argv[1]))）
func main() {
	rawURL := defaultURL()
	if len(os.Args) >= 2 {
		rawURL = os.Args[1]
	}
	u, err := ParseURL(rawURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if err := Load(os.Stdout, u); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
