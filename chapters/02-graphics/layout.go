package main

import "unicode"

// 原文: Creating Windows / Laying Out Text（WIDTH, HEIGHT, HSTEP, VSTEP, SCROLL_STEP）
const (
	WIDTH, HEIGHT = 800, 600
	HSTEP, VSTEP  = 13, 18
	SCROLL_STEP   = 100
)

// 原文: Laying Out Text（lex）
// 前章の show を、テキストを出力せず「返す」関数に変えたもの。
func Lex(body string) string {
	var text []rune
	inTag := false
	for _, c := range body {
		switch {
		case c == '<':
			inTag = true
		case c == '>':
			inTag = false
		case !inTag:
			text = append(text, c)
		}
	}
	return string(text)
}

// 原文: Scrolling Text（display list の要素。Python では (x, y, c) のタプル）
type DisplayItem struct {
	X, Y int
	C    rune
}

// 原文: Scrolling Text（layout）
// ページ座標で各文字の位置を計算する。スクロールのことは一切知らない。
func Layout(text string) []DisplayItem {
	displayList := []DisplayItem{}
	cursorX, cursorY := HSTEP, VSTEP
	for _, c := range text {
		displayList = append(displayList, DisplayItem{cursorX, cursorY, c})
		cursorX += HSTEP
		// 原文: Laying Out Text（画面の端に達したら折り返す）
		if cursorX >= WIDTH-HSTEP {
			cursorY += VSTEP
			cursorX = HSTEP
		}
	}
	return displayList
}

// 原文: Faster Rendering（画面外の文字は描画しない）
// 表示ウィンドウより下、または上にある文字なら false を返す。
// y+VSTEP は文字の下端。半分だけ見えている文字も描画する必要があるため。
func Visible(item DisplayItem, scroll int) bool {
	if item.Y > scroll+HEIGHT {
		return false
	}
	if item.Y+VSTEP < scroll {
		return false
	}
	return true
}

// Tk の create_text は制御文字（改行など）を何も描かない。Go側でも描画を飛ばす。
func Drawable(c rune) bool {
	return !unicode.IsControl(c)
}
