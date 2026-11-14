# 第1章 演習（Exercises）

> 実装は求められたときのみ行います。ここでは各演習の要点とGoでの着手ポイントをTODOとしてまとめます。
> 実装先はすべて [browser.go](browser.go) です。

- [ ] **1-1 HTTP/1.1**: `Host` に加えて `Connection: close` を送り、`HTTP/1.1` と宣言する。`User-Agent` も追加する。今後ヘッダを足しやすくする。
  - 着手点: `Request` 内のヘッダ組み立てを `map[string]string`（または順序付きスライス）にする。`GET %s HTTP/1.1`。
- [ ] **1-2 ファイルURL**: `file` スキームでローカルファイルを開く。URL未指定で起動したら特定のファイルを開く。
  - 着手点: `ParseURL` で `file://` を許可し、`Request` で `os.ReadFile` に分岐。`main` の引数なし時のデフォルトパス。
- [ ] **1-3 `data`**: `data:text/html,Hello world!` をサポートする。
  - 着手点: `ParseURL` で `data:` を `://` なしで扱う（先頭の `scheme:` を別処理）。
- [ ] **1-4 エンティティ**: `&lt;` と `&gt;` を `<` と `>` として表示する。
  - 着手点: `Show` で `strings.NewReplacer` を使うか、状態機械に `&...;` の読み取りを足す。
- [ ] **1-5 `view-source`**: `view-source:http://example.org/` でHTMLソースをそのまま表示する。1-4の実装が前提。
  - 着手点: `URL` に `ViewSource bool` を持たせ、`Load` で `Show` をスキップして本文を出力。
- [ ] **1-6 Keep-alive**: `Connection: keep-alive` を送り、`Content-Length` バイトだけ読み、ソケットを再利用する。
  - 着手点: `map[hostPort]net.Conn`（または `bufio.Reader` ごと）を保持。`io.ReadFull` / `io.LimitReader` でボディを読む。
- [ ] **1-7 リダイレクト**: 3xx で `Location` に再リクエスト。相対 `Location`（`/` 始まり）に対応し、回数上限を設ける。
  - 着手点: ステータスコードを返すよう `Request` を拡張（または内部関数に分離）。テストは `httptest` で連鎖を作る。
- [ ] **1-8 キャッシュ**: `GET` + `200` をキャッシュ。`Cache-Control` の `no-store` と `max-age` に対応し、それ以外の値ならキャッシュしない。
  - 着手点: `map[string]cacheEntry{body, expires}`。時刻は注入可能にしてテストしやすくする。
- [ ] **1-9 圧縮**: `Accept-Encoding: gzip` を送り、`Content-Encoding: gzip` なら展開。`Transfer-Encoding: chunked` にも対応する。
  - 着手点: `compress/gzip`。chunkedは自前でチャンク長（16進）行を読む。`transfer-encoding` / `content-encoding` の拒否チェックを置き換える。
