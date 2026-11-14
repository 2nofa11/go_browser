# 第2章 演習（Exercises）

> 実装は求められたときのみ行います。実装先は [layout.go](layout.go) と [browser.go](browser.go) です。

- [ ] **2-1 改行**: `Layout` で改行文字を見たら行を終えて新しい行を始める。段落区切りに見えるよう `VSTEP` より大きく y を進める。『西遊記』の詩が読めるようになる。
  - 着手点: `Layout` のループで `c == '\n'` を分岐。`Drawable` は制御文字を描かないので描画側の変更は不要。
- [ ] **2-2 マウスホイール**: 上矢印で上にスクロール。ページ先頭より上には行かない。ホイールでもスクロールする。
  - 着手点: `Update` で `ebiten.KeyArrowUp` と `ebiten.Wheel()`。Ebitengine はOS間のホイールの符号差を吸収してくれるので、Tkの `<MouseWheel>` / `<Button-4>` / `<Button-5>` の分岐は不要。`b.scroll` を0でクランプする。
- [ ] **2-3 サイズ変更**: ウィンドウをリサイズ可能にし、幅に応じて `Layout` をやり直す。
  - 着手点: `ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)`、`Browser.Layout(outsideWidth, outsideHeight)` で幅・高さを受け取る。`Layout` 関数に幅を引数で渡すよう変更し、サイズが変わったら再計算する。
- [ ] **2-4 スクロールバー**: 最後のディスプレイリストより下にスクロールしない。右端に青い長方形のスクロールバーを描き、文書全体に対する表示範囲を反映する。全体が収まれば隠す。
  - 着手点: `vector.DrawFilledRect`（`ebiten/v2/vector`）。文書の高さ = 最後のエントリの `Y + VSTEP`。
- [ ] **2-5 絵文字**: 絵文字を画像で描く。OpenMojiの「grinning face」PNGを16×16にして使う。
  - 着手点: `ebiten.NewImageFromImage` と `DrawImage`。Tk の `PhotoImage` / `create_image` に相当。`DisplayItem` に画像を持てるようにする。
- [ ] **2-6 `about:blank`**: 不正なURLでクラッシュせず、空白ページを表示する。`about:blank` を特別に扱う。
  - 着手点: `ParseURL` のエラー時と `about:blank` のとき、`Load` を飛ばして空の `displayList` にする。
- [ ] **2-7 文字方向**: コマンドラインフラグで右から左への基本的なサポートを入れる。英語の文は左から右のまま、画面の右側から伸びていく。
  - 着手点: `flag` パッケージ。`Layout` の `cursorX` を右端から減らす。
