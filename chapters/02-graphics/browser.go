// 第2章: 画面への描画（Drawing to the Screen）
//
// 原文は Python 標準の Tk（tkinter）でウィンドウとキャンバスを作る。Go の標準ライブラリには
// GUI がないため、Ebitengine を採用した（選定理由は README を参照）。
// Tk との対応:
//
//	tkinter.mainloop()          -> ebiten.RunGame（Update/Draw が1フレームごとに呼ばれるイベントループ）
//	canvas.create_text(x, y)    -> text.Draw（中心揃えで Tk の anchor=center に合わせる）
//	canvas.delete("all")        -> 毎フレーム画面全体を塗りつぶして描き直す
//	window.bind("<Down>", ...)  -> Update で下矢印キーの押下を検出
package main

import (
	"fmt"
	"image/color"
	"os"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// 原文: Creating Windows（class Browser）
type Browser struct {
	displayList []DisplayItem
	scroll      int
	face        text.Face
}

func NewBrowser(face text.Face) *Browser {
	return &Browser{face: face}
}

// 原文: Scrolling Text（load: request -> lex -> layout）
func (b *Browser) Load(u *URL) error {
	body, err := u.Request()
	if err != nil {
		return err
	}
	b.displayList = Layout(Lex(body))
	return nil
}

// 原文: Scrolling Text（scrolldown）
func (b *Browser) ScrollDown() {
	b.scroll += SCROLL_STEP
}

// ebiten.Game の実装。Update が原文の handleEvent、Draw が drawScreen に相当する。
func (b *Browser) Update() error {
	// Tk はキーを押し続けると繰り返しイベントを送るので、押下直後と長押し中の一定間隔で反応させる。
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) ||
		(inpututil.KeyPressDuration(ebiten.KeyArrowDown) > 20 &&
			inpututil.KeyPressDuration(ebiten.KeyArrowDown)%3 == 0) {
		b.ScrollDown()
	}
	return nil
}

// 原文: Scrolling Text / Faster Rendering（draw）
func (b *Browser) Draw(screen *ebiten.Image) {
	screen.Fill(color.White) // canvas.delete("all")
	for _, item := range b.displayList {
		if !Visible(item, b.scroll) || !Drawable(item.C) {
			continue
		}
		// 画面座標 = ページ座標 - scroll
		b.drawText(screen, float64(item.X), float64(item.Y-b.scroll), string(item.C))
	}
}

// 原文: Drawing to the Window（canvas.create_text）
// Tk の create_text は指定した座標がテキストの中心になるので、縦横とも中央揃えにする。
func (b *Browser) drawText(screen *ebiten.Image, x, y float64, s string) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.PrimaryAlign = text.AlignCenter
	op.SecondaryAlign = text.AlignCenter
	op.ColorScale.ScaleWithColor(color.Black)
	text.Draw(screen, s, b.face, op)
}

func (b *Browser) Layout(outsideWidth, outsideHeight int) (int, int) {
	return WIDTH, HEIGHT
}

// 日本語・中国語（『西遊記』）を描くには CJK 対応フォントが必要。環境変数 BROWSER_FONT で指定できる。
var fontCandidates = map[string][]string{
	"darwin": {
		"/System/Library/Fonts/Hiragino Sans GB.ttc",
		"/System/Library/Fonts/ヒラギノ角ゴシック W3.ttc",
		"/System/Library/Fonts/Supplemental/Arial Unicode.ttf",
	},
	"linux": {
		"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
		"/usr/share/fonts/noto-cjk/NotoSansCJK-Regular.ttc",
		"/usr/share/fonts/truetype/wqy/wqy-microhei.ttc",
	},
	"windows": {
		`C:\Windows\Fonts\msyh.ttc`,
		`C:\Windows\Fonts\YuGothR.ttc`,
	},
}

func loadFace(size float64) (text.Face, error) {
	paths := fontCandidates[runtime.GOOS]
	if p := os.Getenv("BROWSER_FONT"); p != "" {
		paths = []string{p}
	}
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		sources, err := text.NewGoTextFaceSourcesFromCollection(f)
		f.Close()
		if err != nil || len(sources) == 0 {
			continue
		}
		return &text.GoTextFace{Source: sources[0], Size: size}, nil
	}
	return nil, fmt.Errorf("no usable font found; set BROWSER_FONT to a .ttf/.ttc path")
}

// 原文: Creating Windows（if __name__ == "__main__": Browser().load(URL(sys.argv[1])); mainloop()）
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: browser <url>")
		os.Exit(2)
	}
	u, err := ParseURL(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	face, err := loadFace(16)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	b := NewBrowser(face)
	if err := b.Load(u); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	ebiten.SetWindowSize(WIDTH, HEIGHT)
	ebiten.SetWindowTitle("go_browser")
	if err := ebiten.RunGame(b); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
