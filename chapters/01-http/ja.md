# Webページのダウンロード（Downloading Web Pages）

> 出典: https://browser.engineering/http.html
> 取得日: 2026-11-14
> 原文: *Web Browser Engineering*（Pavel Panchekha、Chris Harrelson 著）第1章の日本語訳（学習用）

Webブラウザ（web browser）は、URLで識別される情報を表示します。その最初のステップは、そのURLを使ってインターネット上のどこかにあるサーバー（server）に接続し、情報をダウンロードすることです。

## サーバーへの接続（Connecting to a Server）

インターネットの閲覧はURLから始まります。[^1]URLとは、ブラウザが訪れるべき特定のWebページを識別する短い文字列です。

```
http://example.org/index.html
```

> 図の注釈: スキーム（Scheme）、ホスト名（Hostname）、パス（Path）

図1: URLの構文。

URLは3つの部分からなります（図1参照）。スキーム（scheme）は情報を*どのように*取得するかを、ホスト名（host name）は*どこから*取得するかを、パス（path）は*どの*情報を取得するかを表します。URLにはほかにも、ポート、クエリ、フラグメントといった省略可能な部分がありますが、それらは後で見ていきます。

URLからWebページのダウンロード処理を始められます。ブラウザはまず、ローカルのオペレーティングシステム（OS）に、*ホスト名*で示された*サーバー*と接続してもらうよう頼みます。するとOSは、ドメインネームシステム（Domain Name System、DNS）サーバーに問い合わせ、`example.org` のようなホスト名を `93.184.216.34` のような*宛先IPアドレス*に変換します。[^2][^3]次にOSは、*ルーティングテーブル*（routing table）と呼ばれるものを使って、その宛先IPアドレスとの通信にどのハードウェア（たとえば無線か有線か）が最適かを決め、デバイスドライバを使って有線または無線で信号を送り出します。[^4]それらの信号は一連の*ルーター*（router）[^5]に拾われて中継され、ルーターはそれぞれメッセージを送るのに最適な方向を選び、最終的に宛先へ届きます。[^6]メッセージがサーバーに届くと、接続が確立されます。要するに、ブラウザがOSに「`example.org` と接続してほしい」と頼めば、そのとおりにしてくれる、ということです。

多くのシステムでは、`telnet` プログラムを使って次のようにこの種の接続を張れます。[^7]

```
telnet example.org 80
```

（注: 灰色の枠が付いているコードは例示にすぎず、私たちのブラウザのコードの一部では*ありません*。）

> `telnet` のインストールが必要な場合があります。デフォルトでは無効になっていることが多いためです。Windowsでは、コントロールパネルの[「プログラムと機能」→「Windowsの機能の有効化または無効化」](https://www.lifewire.com/what-is-telnet-2626026)に進みます。再起動が必要です。実行すると、何かを表示する代わりに画面がクリアされますが、それ以外は通常どおり動作します。macOSでは、`telnet` の代わりに `nc -v` コマンドが使えます。
>
> ```
> nc -v example.org 80
> ```
>
> 出力は少し違いますが、同じように動作します。ほとんどのLinuxシステムでは、パッケージマネージャーから `telnet` または `nc` をインストールできます。通常は `telnet` と `netcat` というパッケージです。

次のような出力が得られます。

```
Trying 93.184.216.34...
Connected to example.org.
Escape character is '^]'.
```

これは、OSがホスト名 `example.org` をIPアドレス `93.184.216.34` に変換し、接続に成功したことを意味します。[^8]これで `example.org` と会話できます。

> **さらに詳しく（Going Further）**
>
> URLの構文は [RFC 3986](https://tools.ietf.org/html/rfc3986) で定義されています。第一著者はティム・バーナーズ＝リー（Tim Berners-Lee）で、驚きはありません。第二著者のロイ・フィールディング（Roy Fielding）はHTTPの設計に大きく貢献した人物で、[博士論文](https://ics.uci.edu/~fielding/pubs/dissertation/fielding_dissertation_2up.pdf)でWebのREST（Representational State Transfer）アーキテクチャを説明したことでも知られています。この論文は、RESTによってWebが分散的に成長できた理由を説明しています。今日では多くのサービスがこれらの原則に従った「RESTful API」を提供していますが、それについては[いくつかの混乱](https://twobithistory.org/2020/06/28/rest.html)もあるようです。

## 情報のリクエスト（Requesting Information）

接続が確立すると、ブラウザはサーバーに*パス*を伝えて情報を要求します。パスとは、`/index.html` のようにURLのホスト名より後ろの部分です。リクエストの構造は図2に示します。試すには、これを `telnet` に入力してください。

```
GET /index.html HTTP/1.0
Host: example.org

```

> 図の注釈: メソッド（Method）、パス（Path）、HTTPバージョン（HTTP Version）、ヘッダ（Header）、値（Value）

図2: 注釈付きのHTTP GETリクエスト。

ここで `GET` という単語は、ブラウザが情報を受け取りたいことを意味します。[^9]続いてパスが来て、最後に `HTTP/1.0` という単語があります。これは、ブラウザが [HTTP](https://developer.mozilla.org/en-US/docs/Web/HTTP) のバージョン1.0を話すことをホストに伝えるものです。HTTPにはいくつかのバージョンがあります（[0.9、1.0、1.1、2.0、3.0](https://medium.com/platform-engineer/evolution-of-http-69cfe6531ba0)）。HTTP 1.1 標準にはkeep-aliveなど便利な機能が数多く追加されていますが、簡単さを優先して、私たちのブラウザはそれらを使いません。HTTP 2.0 も実装しません。1.*x* 系列よりはるかに複雑で、大規模で複雑なWebアプリケーション向けのものですが、私たちのブラウザはどのみちそれらを実行できないからです。

最初の行の後ろの各行には*ヘッダ*（header）が入ります。ヘッダは名前（`Host` など）と値（`example.org` など）を持ちます。ヘッダが違えば意味も違います。たとえば `Host` ヘッダは、あなたが相手をどのホストだと考えているかをサーバーに伝えます。[^10]送れるヘッダはほかにもたくさんありますが、今のところは `Host` だけにしておきましょう。

最後に、ヘッダの後ろに空行が1行入ります。これは、ヘッダが終わったことをホストに伝えます。では `telnet` に空行を入力してください（2行のリクエストを入力した後、Enterを2回押します）。すると `example.org` から応答が返ってくるはずです。

> **さらに詳しく（Going Further）**
>
> HTTP/1.0 は [RFC 1945](https://tools.ietf.org/html/rfc1945) で、HTTP/1.1 は [RFC 2616](https://tools.ietf.org/html/rfc2616) で標準化されています。HTTPは理解しやすく実装しやすいように設計されており、あらゆる種類のコンピューターが採用しやすくなっています。HTTPを `telnet` に直接入力できるのは偶然ではありません。HTTPがプレーンテキストと改行を使う「行指向プロトコル」（line-based protocol）であり、電子メール用のSimple Mail Transfer Protocol（[SMTP](https://en.wikipedia.org/wiki/Simple_Mail_Transfer_Protocol)）に似ているのも、偶然ではありません。そもそもこのパターンは、初期のコンピューターが行単位のテキスト入力しか持たなかったことに由来します。実際、最初期の2つのブラウザのうち1つには、[行モードのUI](https://en.wikipedia.org/wiki/Line_Mode_Browser)がありました。

## サーバーの応答（The Server's Response）

サーバーの応答は、図3の行から始まります。

```
HTTP/1.0 200 OK
```

> 図の注釈: HTTPバージョン（HTTP Version）、レスポンスコード（Response Code）、レスポンスの説明（Response Description）

図3: HTTPレスポンスの最初の行（注釈付き）。

これは、ホストも `HTTP/1.0` を話すこと、そしてリクエストが「OK」（数値コードは200）だったと判断したことを伝えています。`404 Not Found` はご存じかもしれません。これも数値コードと応答の組で、`403 Forbidden` や `500 Server Error` も同様です。こうしたコードはたくさんあり、かなりうまく整理されています。[^11]

- 100番台は情報を知らせるメッセージです。

- 200番台は成功を意味します。

- 300番台は追加の操作を求めるものです（たいていはリダイレクトです）。

- 400番台は、不正なリクエストを送ったことを意味します。

- 500番台は、サーバーがリクエストをうまく処理できなかったことを意味します。

エラーコードが2系統（400番台と500番台）あり、サーバーとブラウザのどちらに非があるかがわかるようになっているのは、見事な発想です。[^12]コードの全一覧は[Wikipedia](https://en.wikipedia.org/wiki/List_of_HTTP_status_codes)で見られ、新しいコードも折にふれて追加されます。

`200 OK` の行の後ろで、サーバーは自分のヘッダを送ってきます。私が試したときは、次のようなヘッダが返ってきました（皆さんの結果は異なるでしょう）。

```
Age: 545933
Cache-Control: max-age=604800
Content-Type: text/html; charset=UTF-8
Date: Mon, 25 Feb 2019 16:49:28 GMT
Etag: "1541025663+gzip+ident"
Expires: Mon, 04 Mar 2019 16:49:28 GMT
Last-Modified: Fri, 09 Aug 2013 23:54:35 GMT
Server: ECS (sec/96EC)
Vary: Accept-Encoding
X-Cache: HIT
Content-Length: 1270
Connection: close
```

ここには*たくさんの*情報があります。要求している情報についてのもの（`Content-Type`、`Content-Length`、`Last-Modified`）、サーバーについてのもの（`Server`、`X-Cache`）、ブラウザがこの情報をどれくらいの間キャッシュすべきかについてのもの（`Cache-Control`、`Expires`、`Etag`）、そのほかさまざまなものです。今のところは先へ進みましょう。

ヘッダの後ろには空行があり、続いて大量の [HTML](https://developer.mozilla.org/en-US/docs/Web/HTML) コードが来ます。これはサーバーの応答の*ボディ*（body）と呼ばれます。ブラウザがそれをHTMLだと判断できるのは、`Content-Type` ヘッダが `text/html` だと言っているからです。Webページそのものの内容を含んでいるのが、このHTMLコードです。

HTTPのリクエストとレスポンスのやり取りは、図4にまとめてあります。手作業での接続から、Python（原書の実装言語）に話を移しましょう。

![図4: HTTPのリクエストとレスポンスの組によって、Webブラウザはウェブサーバーからウェブページを取得する。](img/http-request-2.gif)

*図4: HTTPのリクエストとレスポンスの組によって、Webブラウザはウェブサーバーからウェブページを取得する。*

> 図の補足: 画像内の英語テキストは訳していません。ブラウザがサーバーへリクエストを送り、サーバーがレスポンスを返す流れを表しています。

> **さらに詳しく（Going Further）**
>
> Wikipediaには、HTTPの[ヘッダ](https://en.wikipedia.org/wiki/List_of_HTTP_header_fields)と[レスポンスコード](https://en.wikipedia.org/wiki/List_of_HTTP_status_codes)の優れた一覧があります。レスポンスコードの中には、[402](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/402)「Payment Required」のように、ほとんど使われないものもあります。このコードは「デジタル現金や（マイクロ）決済システム」で使うことを想定していました。eコマースはレスポンスコード402がなくても盛んですが、[マイクロペイメント](https://en.wikipedia.org/wiki/Micropayment)は（まだ？）あまり普及していません。それでも、私を含む多くの人が良いアイデアだと思っています。

## PythonでのTelnet（Telnet in Python）

ここまでは `telnet` を使って別のコンピューターと通信してきました。しかし `telnet` はごく単純なプログラムで、同じことをプログラムから行えます。そのためには、URLからホスト名とパスを取り出し、*ソケット*（socket）を作り、リクエストを送り、レスポンスを受け取る必要があります。[^13]

まずURLの解析から始めましょう。URLの解析は `URL` オブジェクトを返すようにし、解析のコードはコンストラクタに入れます。

```python
class URL:
    def __init__(self, url):
        # ...
```

`__init__` メソッドはクラスのコンストラクタを表すPython独特の構文です。`self` パラメータは、どのメソッドでも必ず最初のパラメータにしなければならないもので、C++やJavaの `this` に相当するPythonのものです。

まずスキームから始めましょう。スキームは、URLの残りの部分と `://` で区切られています。私たちのブラウザがサポートするのは `http` だけなので、それも確認しておきます。

```python
class URL:
    def __init__(self, url):
        self.scheme, url = url.split("://", 1)
        assert self.scheme == "http"
```

次に、ホストとパスを分けなければなりません。ホストは最初の `/` の前にあり、パスはそのスラッシュとそれ以降のすべてです。

```python
class URL:
    def __init__(self, url):
        # ...
        if "/" not in url:
            url = url + "/"
        self.host, url = url.split("/", 1)
        self.path = "/" + url
```

（`# ...` を含むコードブロックは、このように既存のメソッドやブロックにコードを追加することを意味します。）`split(s, n)` メソッドは、文字列を最初の `n` 個の `s` の位置で分割します。ホスト名とパスの間のスラッシュの扱いに、少し込み入ったロジックがあることに注意してください。その（省略可能な）スラッシュはパスの一部です。

これで `URL` は `host` と `path` のフィールドを持ったので、そのURLのWebページをダウンロードできます。新しいメソッド `request` で行いましょう。

```python
class URL:
    def request(self):
        # ...
```

Pythonではメソッドに必ず `self` パラメータを書く必要があることに注意してください。今後は、メソッドの定義をここまで大げさに説明することはしません。まだ存在しないメソッドや関数の中にコードがあるコードブロックを見たら、それはそのメソッドや関数を定義しているということです。

Webページをダウンロードする最初のステップは、ホストへの接続です。OSはそのために「ソケット」という機能を提供しています。ほかのコンピューターと話したいとき（何かを伝えるためでも、相手が伝えてくるのを待つためでも）、ソケットを作ると、そのソケットで情報を双方向にやり取りできます。ほかのコンピューターと話す方法は複数あるため、ソケットにはいくつかの種類があります。

- ソケットには*アドレスファミリ*（address family）があり、相手のコンピューターをどう見つけるかを示します。アドレスファミリの名前は `AF` で始まります。ここで使いたいのは `AF_INET` ですが、たとえば `AF_BLUETOOTH` などもあります。

- ソケットには*タイプ*（type）があり、どのような会話が行われるかを表します。タイプの名前は `SOCK` で始まります。ここで使いたいのは `SOCK_STREAM` で、これは各コンピューターが任意の量のデータを送れることを意味します。一方、`SOCK_DGRAM` というものもあり、こちらは一定サイズのパケットをやり取りします。[^14]

- ソケットには*プロトコル*（protocol）があり、2台のコンピューターが接続を確立する手順を表します。プロトコルの名前はアドレスファミリによって変わりますが、ここで使いたいのは `IPPROTO_TCP` です。[^15]

これらのオプションをすべて選ぶことで、次のようにソケットを作れます。[^16]

```python
import socket

class URL:
    def request(self):
        s = socket.socket(
            family=socket.AF_INET,
            type=socket.SOCK_STREAM,
            proto=socket.IPPROTO_TCP,
        )
```

ソケットができたら、相手のコンピューターに接続するよう指示する必要があります。そのためにはホストと*ポート*（port）が必要です。ポートは使うプロトコルによって決まります。今のところは80にします。

```python
class URL:
    def request(self):
        # ...
        s.connect((self.host, 80))
```

これで `example.org` と通信して接続を確立し、両方のコンピューターがデータをやり取りできる状態になります。

> **注意点（Quirk）**
>
> オフラインのときは、当然これは動きません。プロキシの内側にいる場合や、もっと複雑なさまざまなネットワーク環境でも動かないことがあります。回避策は環境によって異なります。プロキシを無効にするだけで済むかもしれませんし、もっとずっと複雑かもしれません。

`connect` の呼び出しでは括弧が2つあることに注意してください。`connect` は引数を1つだけ取り、その引数がホストとポートの組だからです。アドレスファミリによって引数の数が異なるため、このようになっています。

> **さらに詳しく（Going Further）**
>
> Pythonがほぼそのまま実装している「ソケット」APIは、1983年の4.2 BSD Unix向けの元祖「[バークレーソケット](https://en.wikipedia.org/wiki/Berkeley_sockets)」APIの設計に由来します。もちろんWindowsやLinuxはこのAPIを再実装しているだけですが、macOSやiOSは、BSD Unixの流れをくむ大量のコードを[今も使っています](https://developer.apple.com/library/archive/documentation/Darwin/Conceptual/KernelProgramming/BSD/BSD.html)。

## リクエストとレスポンス（Request and Response）

接続ができたので、相手のサーバーにリクエストを送ります。そのために、`send` メソッドでデータを送ります。

```python
class URL:
    def request(self):
        # ...
        request = "GET {} HTTP/1.0\r\n".format(self.path)
        request += "Host: {}\r\n".format(self.host)
        request += "\r\n"
        s.send(request.encode("utf8"))
```

`send` メソッドは、リクエストをサーバーに送るだけです。[^17]このコードには、正確でなければならない点がいくつかあります。まず、改行には `\n` ではなく `\r\n` を使うことが非常に重要です。また、リクエストの最後に `\r\n` の改行を*2つ*置いて、末尾に空行を送ることも不可欠です。これを忘れると、相手のコンピューターはあなたがその改行を送ってくるのを待ち続け、あなたは相手の応答を待ち続けることになります。[^18]

`encode` の呼び出しにも注目してください。データを送るときは、送っているのは生のビットとバイトであり、それがテキストにも画像にも動画にもなりうることを覚えておく必要があります。しかしPythonの文字列は、特にテキストを表すためのものです。`encode` メソッドはテキストをバイト列に変換し、逆方向の `decode` メソッドもあります。[^19]Pythonは、テキストとバイト列に別々の型を与えることで、注意を促してくれます。

```python
>>> type("text")
<class 'str'>
>>> type("text".encode("utf8"))
<class 'bytes'>
```

`str` と `bytes` に関するエラーが出たら、どこかで `encode` か `decode` を呼び忘れたということです。

サーバーの応答を読むには、ソケットの `read` 関数を使って、すでに届いている応答の断片を受け取れます。そのうえで、届いた断片を集めるループを書きます。しかしPythonでは、そのループを隠してくれる `makefile` というヘルパー関数が使えます。[^20]

```python
class URL:
    def request(self):
        # ...
        response = s.makefile("r", encoding="utf8", newline="\r\n")
```

ここで `makefile` は、サーバーから受け取ったすべてのバイトを含むファイルのようなオブジェクトを返します。私はPythonに、`utf8` *エンコーディング*（encoding）、つまりバイトを文字に対応付ける方式を使って、それらのバイトを文字列に変換するよう指示しています。[^21]また、HTTPの変わった改行コードもPythonに伝えています。

では、応答をいくつかの部分に分割しましょう。最初の行はステータス行です。[^22]

```python
class URL:
    def request(self):
        # ...
        statusline = response.readline()
        version, status, explanation = statusline.split(" ", 2)
```

サーバーのHTTPバージョンが自分のものと同じかどうかは確認*しない*ことに注意してください。確認するのは良い考えに思えるかもしれませんが、HTTP 1.0で話しかけてもHTTP 1.1で応答する設定ミスのサーバーが世の中にはたくさんあるのです。[^23]

ステータス行の後ろにはヘッダが続きます。

```python
class URL:
    def request(self):
        # ...
        response_headers = {}
        while True:
            line = response.readline()
            if line == "\r\n": break
            header, value = line.split(":", 1)
            response_headers[header.casefold()] = value.strip()
```

ヘッダについては、各行を最初のコロンで分割して、ヘッダ名からヘッダ値へのマップ（map）を埋めていきます。ヘッダは大文字小文字を区別しないので、小文字に正規化します。[^24]また、HTTPヘッダ値の空白には意味がないので、前後の余分な空白は取り除きます。

ヘッダはあらゆる種類の情報を表せますが、そのうちのいくつかは特に重要です。私たちが取得しようとしているデータが普通ではない方法で送られていることを、教えてくれるからです。それらが存在しないことを確認しておきましょう。[^25]

```python
class URL:
    def request(self):
        # ...
        assert "transfer-encoding" not in response_headers
        assert "content-encoding" not in response_headers
```

すると、送られてきたデータを得る通常の方法は、ヘッダの後ろにあるすべてを読むことです。

```python
class URL:
    def request(self):
        # ...
        content = response.read()
        s.close()
```

表示したいのはボディなので、それを返しましょう。

```python
class URL:
    def request(self):
        # ...
        return content
```

では、レスポンスボディのテキストを実際に表示しましょう。

> **さらに詳しく（Going Further）**
>
> [`Content-Encoding`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Encoding) ヘッダを使うと、サーバーはWebページを圧縮してから送れます。テキストの多い大きなWebページはよく圧縮でき、その結果、ページの読み込みが速くなります。ブラウザは、サポートする圧縮アルゴリズムを列挙した [`Accept-Encoding` ヘッダ](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept-Encoding)をリクエストに含めて送る必要があります。[`Transfer-Encoding`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Transfer-Encoding) も同様のもので、データを「チャンク化」（chunked）することもできます。多くのサーバーが、これを圧縮と併用しているようです。

## HTMLの表示（Displaying the HTML）

レスポンスボディのHTMLコードは、`http://example.org/index.html` にアクセスしたときにブラウザウィンドウで目にする内容を定義しています。HTMLについては今後の章でずっと詳しく説明しますが、今のところはごく単純にしておきます。

HTMLには*タグ*（tag）と*テキスト*（text）があります。タグはそれぞれ `<` で始まり `>` で終わります。大まかに言えば、タグはある内容がどんな種類のものかを伝え、テキストが実際の内容です。[^26]ほとんどのタグは開始タグと終了タグのペアになっています。たとえばページのタイトルは、`<title>` と `</title>` というタグのペアで囲まれます。各タグは山括弧の中にタグ名（ここでは `title`）を持ち、その後ろに必要に応じて空白と*属性*（attribute）が続きます。ペアとなる終了タグは、`/` とタグ名（属性は付きません）を持ちます。

そこで、ごく単純なWebブラウザを作るために、ページのHTMLを受け取って、その中のテキストだけを（タグは除いて）すべて表示することにしましょう。[^27]これを新しい関数 `show` で行います。[^28]

```python
def show(body):
    in_tag = False
    for c in body:
        if c == "<":
            in_tag = True
        elif c == ">":
            in_tag = False
        elif not in_tag:
            print(c, end="")
```

このコードはかなり複雑です。リクエストボディを1文字ずつ処理し、2つの状態を持ちます。現在山括弧の間にいる `in_tag` と、そうでない `not in_tag` です。現在の文字が山括弧のとき、状態が切り替わります。タグの外にある通常の文字は表示されます。[^29]

これで `request` と `show` をつなぎ合わせるだけで、Webページを読み込めます。[^30]

```python
def load(url):
    body = url.request()
    show(body)
```

コマンドラインから `load` を実行するために、次のコードを追加してください。

```python
if __name__ == "__main__":
    import sys
    load(URL(sys.argv[1]))
```

最初の行は、Pythonにおける `main` 関数にあたるもので、このスクリプトをコマンドラインから実行したときだけ動きます。このコードは、コマンドラインの最初の引数（`sys.argv[1]`）を読み取り、それをURLとして使います。URL `http://example.org/` でこのコードを実行してみましょう。

```sh
python3 browser.py http://example.org/
```

公式のサンプルWebページへようこそ、という短いテキストが表示されるはずです。[この章](http://browser.engineering/http.html)に対して使ってみることもできます！

> **さらに詳しく（Going Further）**
>
> HTMLも、URLやHTTPと同じように、基本的なレベルでは非常に簡単に解析・表示できるよう設計されています。そして当初のHTMLには機能がほとんどなかったため、ここで見たものよりそれほど高機能でないものを作るだけで、内容を実用的に表示できました。私たちのごく単純で基本的なHTMLパーサーでも、すでに [browser.engineering](https://browser.engineering/) のWebサイトのテキストを表示できます。

## 暗号化された接続（Encrypted Connections）

ここまでで、私たちのブラウザは `http` スキームをサポートしています。これはごく一般的なスキームです。しかし、`https` スキームへ移行するWebサイトはますます増えており、それを必須とするWebサイトも多くあります。

`http` と `https` の違いは、`https` のほうが安全であることです。ただ、もう少し具体的に言いましょう。`https` スキーム、より正式にはHTTP over TLS（Transport Layer Security）は、ブラウザとホストの間のすべての通信が暗号化される点を除いて、通常の `http` スキームと同じです。この仕組みには、どの暗号化アルゴリズムを使うか、共通の暗号鍵をどう取り決めるか、そしてもちろんブラウザが正しいホストに接続していることをどう確かめるかなど、多くの詳細があります。関係するプロトコル層の違いを図5に示します。

![図5: HTTPとHTTPSの違いは、TLS層が加わることである。](img/http-tls-2.gif)

*図5: HTTPとHTTPSの違いは、TLS層が加わることである。*

> 図の補足: 画像内の英語テキストは訳していません。HTTPSでは、HTTPとTCPの間にTLS層が加わることを表しています。

幸いPythonの `ssl` ライブラリがこれらの詳細をすべて実装してくれているので、暗号化された接続を作るのは、通常の接続を作るのとほとんど同じくらい簡単です。この手軽さは、状況によっては不適切かもしれないデフォルト設定を受け入れることと引き換えですが、教育目的なら問題ありません。

`ssl` で暗号化された接続を作るのはとても簡単です。すでにソケット `s` を作って `example.org` に接続したとします。接続を暗号化するには、`ssl.create_default_context` で*コンテキスト*（context）`ctx` を作り、そのコンテキストでソケット `s` を*ラップ*（wrap）します。

```python
import ssl
ctx = ssl.create_default_context()
s = ctx.wrap_socket(s, server_hostname=host)
```

`wrap_socket` は新しいソケットを返すので、それを変数 `s` に保存し直していることに注意してください。元のソケットではデータを送りたくないからです。暗号化されませんし、混乱のもとにもなります。`server_hostname` 引数は、正しいサーバーに接続したことを確認するために使われます。`Host` ヘッダと一致している必要があります。

> macOSでは、ほとんどのWebサイトでPythonの `ssl` パッケージを使う前に、[「Install Certificates」というプログラムを実行する](https://stackoverflow.com/questions/52805115/certificate-verify-failed-unable-to-get-local-issuer-certificate)必要があります。

このコードを `request` に加えてみましょう。まず、どのスキームが使われているかを検出する必要があります。

```python
import ssl

class URL:
    def __init__(self, url):
        self.scheme, url = url.split("://", 1)
        assert self.scheme in ["http", "https"]
        # ...
```

（ここでは、既存のスキーム解析のコードをこの新しいコードに置き換えることになっている点に注意してください。何を置き換えるべきかは、たいてい文脈とコード自体から明らかです。）

暗号化されたHTTP接続は、通常80番ではなく443番ポートを使います。

```python
class URL:
    def __init__(self, url):
        # ...
        if self.scheme == "http":
            self.port = 80
        elif self.scheme == "https":
            self.port = 443
```

ソケットを作るときに、このポートを使えます。

```python
class URL:
    def request(self):
        # ...
        s.connect((self.host, self.port))
        # ...
```

次に、`ssl` ライブラリでソケットをラップします。

```python
class URL:
    def request(self):
        # ...
        s.connect((self.host, self.port))
        if self.scheme == "https":
            ctx = ssl.create_default_context()
            s = ctx.wrap_socket(s, server_hostname=self.host)
        # ...
```

これで、ブラウザはHTTPSのサイトに接続できるはずです。

ついでに、カスタムポートのサポートも追加しましょう。カスタムポートは、図6のように、URLでホスト名の後ろにコロンを置いて指定します。

```
http://example.org:8080/index.html
```

> 図の注釈: ポート（Port）

図6: URLにおけるポートの位置。

URLにポートがあれば、それを解析して使えます。

```python
class URL:
    def __init__(self, url):
        # ...
        if ":" in self.host:
            self.host, port = self.host.split(":", 1)
            self.port = int(port)
```

カスタムポートはデバッグに便利です。Pythonには、自分のコンピューター上のファイルを配信できる組み込みのWebサーバーがあります。たとえば次のように実行すると、

```sh
python3 -m http.server 8000 -d /some/directory
```

`http://localhost:8000/` にアクセスすると、そのディレクトリ内のすべてのファイルが表示されるはずです。これはブラウザをテストするのに良い方法です。

> **さらに詳しく（Going Further）**
>
> TLSはかなり複雑です。詳細は [RFC 8446](https://tools.ietf.org/html/rfc8446) で読めますが、自前で実装することはお勧めしません。正しいだけでなく安全でもあるカスタムTLS実装を書くのは、非常に難しいのです。

ここまでで、プログラムを任意のWebページに対して実行できるはずです。[簡単な例](https://browser.engineering/examples/example1-simple.html)に対する出力は次のようになります。

```

    This is a simple
    web page with some
    text in it.

```

## まとめ（Summary）

この章では、空のファイルから出発して、次のことができる初歩的なWebブラウザを作りました。

- URLをスキーム、ホスト、ポート、パスに解析する。

- `socket` と `ssl` ライブラリを使ってそのホストに接続する。

- `Host` ヘッダを含むHTTPリクエストをそのホストに送る。

- HTTPレスポンスをステータス行、ヘッダ、ボディに分割する。

- ボディ内のテキストを（タグは除いて）表示する。

確かに、これはまだWebブラウザというよりコマンドラインツールですが、すでにブラウザの中核的な機能のいくつかを備えています。

## アウトライン（Outline）

私たちのブラウザの関数、クラス、メソッドの全体は、次のようになっているはずです。

```python
class URL:
    def __init__(url)

    def request()


def show(body)

def load(url)
```

## 演習（Exercises）

1-1 *HTTP/1.1*。`request` 関数で、`Host` に加えて、値を `close` にした `Connection` ヘッダを送ってください。これでブラウザは `HTTP/1.1` を使っていると宣言できます。あわせて `User-Agent` ヘッダも追加してください。値は何でもかまいません。ホストに対してあなたのブラウザを識別するものです。今後ヘッダを追加しやすいようにしてください。

1-2 *ファイルURL*。ローカルファイルを開けるようにする `file` スキームのサポートを追加してください。たとえば `file:///path/goes/here` は、あなたのコンピューター上の `/path/goes/here` にあるファイルを指すようにします。また、URLを指定せずにブラウザを起動した場合は、コンピューター上の特定のファイルを開くようにしてください。そのファイルを手早いテストに使えます。

1-3 *`data`*。もう1つのスキームが `data` で、HTMLの内容をURL自体にインライン化できます。実際のブラウザで `data:text/html,Hello world!` に移動して、何が起こるか試してみてください。このスキームのサポートをブラウザに追加してください。`data` スキームは、別ファイルにしなくてもテストが作れるので、特に便利です。

1-4 *エンティティ*。小なり（`&lt;`）と大なり（`&gt;`）のエンティティ（entity）のサポートを実装してください。これらはそれぞれ `<` と `>` として表示されるべきです。たとえば、HTMLレスポンスが `&lt;div&gt;` だった場合、ブラウザの `show` メソッドは `<div>` を表示するようにします。エンティティを使うと、Webページは、ブラウザにタグとして解釈されることなく、これらの特殊文字を含められます。

1-5 *`view-source`*。`view-source` スキームのサポートを追加してください。`view-source:http://example.org/` に移動すると、レンダリングされたページの代わりにHTMLソースを表示するようにします。このスキームのサポートを追加してください。ブラウザは、HTMLファイル全体をテキストのように表示するはずです。演習1-4も実装しておくとよいでしょう。

1-6 *Keep-alive*。演習1-1を実装してください。ただし `Connection: close` ヘッダは送らず、代わりに `Connection: keep-alive` を送ります。ソケットからボディを読むときは、`Content-Length` ヘッダで示されたバイト数だけを読み、その後ソケットを閉じないでください。代わりにソケットを保存しておき、同じサーバーへ別のリクエストが行われたら、新しくソケットを作らずに同じソケットを再利用します。（`makefile` に `"rb"` オプションも渡す必要があります。そうしないと、`Content-Length` が報告する値と、読んでいる文字列の長さが一致しないかもしれません。）これにより、同じサーバーへの繰り返しのリクエスト（よくあることです）が高速になります。

1-7 *リダイレクト*。300番台のエラーコードはリダイレクトを要求します。ブラウザがそれに遭遇したら、`Location` ヘッダで指定されたURLへ新しいリクエストを行うようにしてください。`Location` ヘッダが完全なURLのこともありますが、ホストとスキームを省略して `/` で始まるだけのこともあります（元のリクエストと同じホストとスキームを意味します）。新しいURLがさらにリダイレクトの場合もあるので、そのケースも扱ってください。ただし、リダイレクトのループにはまり込みたくないので、ブラウザが連続して従うリダイレクトの回数に上限を設けてください。この章のページにリダイレクトするURL [http://browser.engineering/redirect](http://browser.engineering/redirect) と、より複雑なリダイレクトの連鎖を行う兄弟分の [/redirect2](http://browser.engineering/redirect2)、[/redirect3](http://browser.engineering/redirect3) で試せます。

1-8 *キャッシュ*。一般に、同じ画像、スタイル、スクリプトが複数のページで使われます。それらを何度もダウンロードするのは無駄です。`GET` でリクエストし、`200` のレスポンスを受け取ったものであれば、どのHTTPレスポンスもキャッシュして一般に問題ありません。[^31]ブラウザにキャッシュを実装し、同じファイルを複数回リクエストしてテストしてください。サーバーは `Cache-Control` ヘッダでキャッシュを制御します。このヘッダ、具体的には `no-store` と `max-age` の値のサポートを追加してください。`Cache-Control` ヘッダにこの2つ以外の値が含まれている場合は、そのレスポンスをキャッシュしないのが無難です。

1-9 *圧縮*。HTTP圧縮のサポートを追加してください。これは、ブラウザが圧縮データを受け付けることを[サーバーに知らせる](https://developer.mozilla.org/en-US/docs/Web/HTTP/Content_negotiation)ものです。ブラウザは、値を `gzip` にした `Accept-Encoding` ヘッダを送らなければなりません。サーバーが圧縮をサポートしていれば、レスポンスには値が `gzip` の `Content-Encoding` ヘッダが付きます。その場合、ボディは圧縮されています。このケースのサポートを追加してください。データの展開には、`gzip` モジュールの `decompress` メソッドが使えます。GZipデータは `utf8` エンコードされていないので、`makefile` に `"rb"` を渡して、代わりに生のバイト列で扱ってください。ほとんどのWebサーバーは、[`chunked`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Transfer-Encoding) と呼ばれる `Transfer-Encoding` で圧縮データを送ります。[^32]それへの対応も追加する必要があります。

[^1]: 「URL」は「uniform resource locator」の略です。Webページ（リソース）を識別するための可搬な（uniform）方法であること、そしてそれらのファイルへのアクセス方法も記述している（locator）ことを意味します。
[^2]: `nslookup.io` のようなDNS検索ツールや `dig` コマンドを使えば、この変換を自分で行えます。（[nslookup.io](https://nslookup.io)）
[^3]: 現在、IP（Internet Protocol）にはIPv4とIPv6の2つのバージョンがあります。IPv6アドレスはずっと長く、通常は16進数で書かれますが、それ以外の違いはここでは重要ではありません。
[^4]: ここでは手順を省略しています。有線では、まず通信をイーサネットフレームで包む必要があり、無線ではさらに多くのことが必要です。簡潔にしようとしています。
[^5]: あるいはスイッチやアクセスポイントなど、さまざまな可能性がありますが、最終的にはルーターに行き着きます。
[^6]: ルーターは、返信を転送し返せるように、メッセージの送信元を記録することもあります。
[^7]: 「80」はポートで、後で説明します。
[^8]: エスケープ文字についての行は、`telnet` の難解な機能の使い方を示す案内にすぎません。
[^9]: 情報を送るつもりなら `POST` と言うこともできます。そのほかにも、もっと難解な選択肢がいくつかあります。
[^10]: これは、同じIPアドレスが複数のホスト名に対応し、複数のWebサイト（たとえば `example.com` と `example.org`）をホストしている場合に役立ちます。`Host` ヘッダは、複数のWebサイトのうちどれが欲しいのかをサーバーに伝えます。これらのWebサイトは、基本的に `Host` ヘッダがないと正しく動作しません。1台のコンピューターで複数のドメインをホストするのは非常によくあることです。
[^11]: `OK` のようなステータステキストは、実際には何でもよく、機械のためではなく人間のためにあるだけです。
[^12]: より正確には、サーバーが非があると考えているのはどちらか、です。
[^13]: Pythonには、URLを解析するための `urllib.parse` というライブラリがありますが、自前で実装したほうが学習に良いと思います。それに、そうすればこの本がPython固有のものになりすぎません。
[^14]: `DGRAM` は「datagram」（データグラム）の略で、私は葉書のようなものだと想像しています。
[^15]: 新しいバージョンのHTTPは、Transmission Control Protocol（TCP）の代わりに[QUIC](https://en.wikipedia.org/wiki/QUIC)というものを使いますが、私たちのブラウザはHTTP 1.0にとどまります。
[^16]: このコードはPythonの `socket` ライブラリを使っていますが、皆さんのお気に入りの言語にもおそらく非常によく似たライブラリがあるでしょう。このAPIは基本的に標準化されています。Pythonでは、ここで渡しているフラグはデフォルト値なので、実際には `socket.socket()` と呼ぶだけでも構いません。ほかの言語で試している方のために、ここではフラグを残しています。
[^17]: `send` は実際には数を返します。この場合は `47` です。それは、相手のコンピューターに何バイトのデータを送ったかを表します。たとえば、データの送信途中でネットワーク接続が切れた場合、切れるまでにどれだけ送ったかを知りたいかもしれません。
[^18]: コンピューターは、どこまでも文字どおりにしか受け取りません。
[^19]: `encode` と `decode` を呼ぶときは、どの*文字エンコーディング*（character encoding）を使いたいかをコンピューターに伝える必要があります。これは複雑な話題です。ここでは一般的な文字エンコーディングで多くのページで動作する `utf8` を使っていますが、現実の世界ではもっと注意深くある必要があります。
[^20]: ほかの言語を使っている場合は、`socket.read` しか使えないかもしれません。その場合は、ソケットの状態を確認しながら、ループを自分で書く必要があります。
[^21]: `utf8` を決め打ちにするのは正しくありませんが、英語のWebサイトの大半ではそれなりに動く近道です。実際には、`Content-Type` ヘッダにはたいてい、ボディのエンコーディングを指定する `charset` 宣言が含まれています。それがない場合でも、ブラウザは `utf8` をデフォルトにはしません。文字の出現頻度から推測し、推測を外すと、� strange áççêñ£ß のような醜い文字が表示されます。
[^22]: 200が必須だと `assert` することもできました。私たちのブラウザがサポートするコードはそれだけだからです。しかし、返されたボディをそのままレンダリングするほうが良いのです。サーバーは通常、エラーコードに対しても、有用でユーザーが読めるHTMLのエラーページを出力するからです。これも、Webを段階的に実装しやすいことの一例です。
[^23]: 幸い、プロトコルは混乱を招かない程度には似ています。
[^24]: `lower` ではなく [`casefold`](https://docs.python.org/3/library/stdtypes.html#str.casefold) を使ったのは、より多くの言語でうまく動作するからです。
[^25]: これらのヘッダが存在する場合のブラウザの扱い方は、演習1-9で説明しています。
[^26]: とはいえ、`img` のように、情報についてのものではなく、それ自体が内容であるタグもあります。
[^27]: この例でPythonが最終行の `end` を指す `SyntaxError` を出す場合は、Python 3ではなくPython 2で実行している可能性が高いです。必ずPython 3を使ってください。
[^28]: これはグローバル関数であり、`URL` クラスの中にはないことに注意してください。
[^29]: `end` 引数は、Pythonに、文字の後ろで改行を出力しないよう伝えます。指定しなければ改行が出力されます。
[^30]: `show` と同様、これもグローバル関数です。
[^31]: `301` や `404` など、ほかのステータスコードもキャッシュできる場合があります。
[^32]: データを圧縮する `Transfer-Encoding` もいくつかあります。あまり一般的には使われません。
