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
// Python の class は struct＋メソッドで表す。
type URL struct {
	Scheme string
	Host   string
	Port   int
	Path   string

	// Data は data スキームの内容そのもの（data:text/html,<ここ>）。他のスキームでは空。
	Data string

	// TLSConfig が nil なら、システムの証明書で検証するデフォルト設定を使う。
	// テストで自己署名証明書のサーバーに接続するためだけに差し替える。
	TLSConfig *tls.Config
}

// 原文: Telnet in Python / Encrypted Connections（URL.__init__）
// Python の assert は error 戻り値に置き換える。
func ParseURL(raw string) (*URL, error) {
	// 原文: 演習1-3（data スキーム）
	// data:text/html,Hello world! は "://" を持たないので、先に別処理する。
	// "," より前がメディアタイプ、後ろ全体が内容になる。
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

	// url.split("://", 1)
	scheme, rest, ok := strings.Cut(raw, "://")
	if !ok {
		return nil, fmt.Errorf("invalid URL %q: missing scheme", raw)
	}
	// 原文: 演習1-2（file スキーム）
	// file:///path/goes/here は、ホストを持たず、"://" の後ろ全体がファイルのパスになる。
	if scheme == "file" {
		if !strings.HasPrefix(rest, "/") {
			return nil, fmt.Errorf("invalid file URL %q: path must be absolute (file:///...)", raw)
		}
		return &URL{Scheme: scheme, Path: rest}, nil
	}
	// assert self.scheme in ["http", "https"]
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme %q", scheme)
	}
	u := &URL{Scheme: scheme}

	// if self.scheme == "http": self.port = 80 / elif "https": self.port = 443
	if scheme == "http" {
		u.Port = 80
	} else {
		u.Port = 443
	}

	// if "/" not in url: url = url + "/"
	if !strings.Contains(rest, "/") {
		rest += "/"
	}
	// self.host, url = url.split("/", 1); self.path = "/" + url
	host, path, _ := strings.Cut(rest, "/")
	u.Host = host
	u.Path = "/" + path

	// if ":" in self.host: self.host, port = self.host.split(":", 1)
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
// レスポンスボディを返す。ステータスコードは検査せず、ボディをそのまま返す。
func (u *URL) Request() (string, error) {
	// 原文: 演習1-3（data スキーム）
	// 内容はURLの中にあるので、ネットワークにもファイルにも触れずにそのまま返す。
	if u.Scheme == "data" {
		return u.Data, nil
	}

	// 原文: 演習1-2（file スキーム）
	// ネットワークには出ず、ローカルのファイルをそのまま読んで返す。
	if u.Scheme == "file" {
		data, err := os.ReadFile(u.Path)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	// s = socket.socket(AF_INET, SOCK_STREAM, IPPROTO_TCP) と s.connect(...)
	// Goの net.Dial("tcp", ...) がソケット生成と接続をまとめて行う。
	addr := net.JoinHostPort(u.Host, strconv.Itoa(u.Port))
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return "", err
	}

	// 原文: Encrypted Connections（ctx.wrap_socket(s, server_hostname=...)）
	// wrap_socket と同様、元のコネクションは使わず、ラップした方を以降で使う。
	if u.Scheme == "https" {
		cfg := &tls.Config{}
		if u.TLSConfig != nil {
			cfg = u.TLSConfig.Clone()
		}
		cfg.ServerName = u.Host // server_hostname
		conn = tls.Client(conn, cfg)
	}
	defer conn.Close() // s.close()

	// 原文: Request and Response / 演習1-1
	// request = "GET {} HTTP/1.1\r\n" + ヘッダ + 空行; s.send(request.encode("utf8"))
	// Goの string は UTF-8 のバイト列なので encode は不要。\r\n は厳密に守る。
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

	// response = s.makefile("r", encoding="utf8", newline="\r\n")
	response := bufio.NewReader(conn)

	// statusline = response.readline(); version, status, explanation = split(" ", 2)
	statusline, err := response.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading status line: %w", err)
	}
	parts := strings.SplitN(strings.TrimRight(statusline, "\r\n"), " ", 3)
	if len(parts) < 2 {
		return "", fmt.Errorf("malformed status line %q", statusline)
	}
	// サーバーのHTTPバージョン(parts[0])は確認しない。設定ミスのサーバーが多いため。
	// ステータスコード(parts[1])も検査せず、エラーページもそのまま表示する。

	// response_headers = {} ... header.casefold(): value.strip()
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
		// ヘッダ名は大文字小文字を区別しないので小文字に正規化する。
		headers[strings.ToLower(name)] = strings.TrimSpace(value)
	}

	// assert "transfer-encoding" not in response_headers など
	if _, ok := headers["transfer-encoding"]; ok {
		return "", fmt.Errorf("transfer-encoding is not supported")
	}
	if _, ok := headers["content-encoding"]; ok {
		return "", fmt.Errorf("content-encoding is not supported")
	}

	// content = response.read()
	body, err := io.ReadAll(response)
	if err != nil {
		return "", fmt.Errorf("reading body: %w", err)
	}
	return string(body), nil
}

// 原文: Displaying the HTML（show）
// タグ（< と > の間）を除いたテキストだけを w に書き出す。
func Show(w io.Writer, body string) {
	inTag := false
	for _, c := range body { // range は rune 単位なので UTF-8 の多バイト文字も壊さない
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

// 原文: 演習1-2（URLを指定せずに起動したときに開く、コンピューター上の特定のファイル）
// 手早いテストに使えるファイルの URL（file:///...）を返す。
func defaultURL() string {
	// 実行場所（カレントディレクトリ）に依存しないよう、このソースファイルの位置を基準にする。
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
