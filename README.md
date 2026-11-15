# go_browser

[Web Browser Engineering](https://browser.engineering/) を、日本語訳を読みながらGoで実装していく学習用リポジトリ。

## 構成

```
chapters/<番号>-<スラッグ>/
  source.md      # 原文Markdown（出典付き）
  ja.md          # 日本語訳
  img/           # 図（ローカル保存）
  *.go           # Go実装（章ごとに package main）
  exercises.md   # 演習のTODO
GLOSSARY.md      # 訳語の統一
```

## 方針

- 標準ライブラリ優先。`net/http` は使わず、`net` と `crypto/tls` で原文と同じ粒度の低レベルAPIを使う。
- GUIは標準ライブラリにないため、第2章から [Ebitengine](https://ebitengine.org/) を採用した。選定理由: (1) macOSでcgo不要でビルドでき環境構築が軽い、(2) `Update`/`Draw` のループが原文のイベントループ（`mainloop`）と対応づけやすい、(3) `text/v2` でCJK（『西遊記』）を含む文字単位の描画ができる。Fyne等のウィジェット型ツールキットは、原文が「ウィジェットではなくキャンバスで細かく制御する」方針なので不採用。中国語の表示にはCJKフォントが必要で、環境変数 `BROWSER_FONT` で指定できる。
- `ja.md` のコードブロックと、Python固有の説明（`self`、`encode`、`makefile` など）もGoに置き換える。原書の見出しは、`GoでのTelnet（Telnet in Python）` のように原題を残し、訳注で置き換えを明記する。Go版の全体は各章の `*.go` にある。
- 原文の `assert` は `error` 戻り値、Pythonの class は struct＋メソッドに置き換える。
- 各ブロックに `// 原文: <節タイトル>` コメントを付ける。
- 検証: `gofmt -l .` / `go vet ./...` / `go test ./...`。テストは外部ネットワークに出ない（`httptest`）。

## 章

章番号とスラッグは [目次](https://browser.engineering/) に合わせる（ディレクトリは `chapters/<番号>-<スラッグ>/`）。
Introduction の3ページには番号がないため、`00a`〜`00c` として目次の順に並べる。

| ディレクトリ | 目次での位置 | タイトル | 状態 |
|---|---|---|---|
| [00a-preface](chapters/00a-preface/) | Introduction | Preface | 翻訳済み（コードなし） |
| [00b-intro](chapters/00b-intro/) | Introduction | Browsers and the Web | 翻訳済み（コードなし） |
| [00c-history](chapters/00c-history/) | Introduction | History of the Web | 翻訳済み（コードなし、図8枚） |
| [01-http](chapters/01-http/) | Part 1: Loading Pages | Downloading Web Pages | 翻訳・実装済み（演習は未着手） |
| [02-graphics](chapters/02-graphics/) | Part 1: Loading Pages | Drawing to the Screen | 翻訳・実装済み（演習は未着手） |
| `03-text` | Part 1: Loading Pages | Formatting Text | 未着手 |
| `04-html` | Part 2: Viewing Documents | Constructing an HTML Tree | 未着手 |
| `05-layout` | Part 2: Viewing Documents | Laying Out Pages | 未着手 |
| `06-styles` | Part 2: Viewing Documents | Applying Author Styles | 未着手 |
| `07-chrome` | Part 2: Viewing Documents | Handling Buttons and Links | 未着手 |
| `08-forms` | Part 3: Running Applications | Sending Information to Servers | 未着手 |
| `09-scripts` | Part 3: Running Applications | Running Interactive Scripts | 未着手 |
| `10-security` | Part 3: Running Applications | Keeping Data Private | 未着手 |
| `11-visual-effects` | Part 4: Modern Browsers | Adding Visual Effects | 未着手 |
| `12-scheduling` | Part 4: Modern Browsers | Scheduling Tasks and Threads | 未着手 |
| `13-animations` | Part 4: Modern Browsers | Animating and Compositing | 未着手 |
| `14-accessibility` | Part 4: Modern Browsers | Making Content Accessible | 未着手 |
| `15-embeds` | Part 4: Modern Browsers | Supporting Embedded Content | 未着手 |
| `16-invalidation` | Part 4: Modern Browsers | Reusing Previous Computation | 未着手 |
| `99a-skipped` | Conclusion | What Wasn't Covered | 未着手 |
| `99b-change` | Conclusion | A Changing Landscape | 未着手 |

実行例:

```bash
go run ./chapters/01-http http://example.org/
go run ./chapters/02-graphics https://browser.engineering/examples/xiyouji.html
```

第2章は下矢印キーでスクロールする。
