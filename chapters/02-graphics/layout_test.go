package main

import (
	"strings"
	"testing"
)

func TestLex(t *testing.T) {
	got := Lex("<html><body>Hi <b>there</b>!</body></html>")
	if want := "Hi there!"; got != want {
		t.Errorf("Lex = %q, want %q", got, want)
	}
}

func TestLayoutPositions(t *testing.T) {
	got := Layout("abc")
	want := []DisplayItem{
		{HSTEP, VSTEP, 'a'},
		{HSTEP * 2, VSTEP, 'b'},
		{HSTEP * 3, VSTEP, 'c'},
	}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("item %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestLayoutWraps(t *testing.T) {
	// 1行に入る文字数: cursor_x が WIDTH-HSTEP(787) 以上になったら折り返す。
	// 13, 26, ... と増え、13*60=780 は未満、13*61=793 で折り返す。
	text := strings.Repeat("x", 200)
	items := Layout(text)
	if items[0].Y != VSTEP {
		t.Fatalf("first Y = %d", items[0].Y)
	}
	perLine := 0
	for _, it := range items {
		if it.Y != VSTEP {
			break
		}
		perLine++
	}
	if perLine != 60 {
		t.Errorf("chars per line = %d, want 60", perLine)
	}
	if items[perLine].X != HSTEP || items[perLine].Y != 2*VSTEP {
		t.Errorf("wrapped item = %+v, want X=%d Y=%d", items[perLine], HSTEP, 2*VSTEP)
	}
	for _, it := range items {
		if it.X >= WIDTH { // 右端を超えない
			t.Fatalf("item beyond right edge: %+v", it)
		}
	}
}

func TestLayoutIsRuneBased(t *testing.T) {
	// 西游记のような多バイト文字も1文字ずつ配置される。
	items := Layout("西游记")
	if len(items) != 3 || items[2].C != '记' {
		t.Errorf("items = %+v", items)
	}
}

func TestVisible(t *testing.T) {
	scroll := 1000
	tests := []struct {
		name string
		y    int
		want bool
	}{
		{"above window", scroll - VSTEP - 1, false},
		{"bottom edge touches top", scroll - VSTEP, true}, // 半分見えている文字は描画する
		{"inside", scroll + 100, true},
		{"last visible", scroll + HEIGHT, true},
		{"below window", scroll + HEIGHT + 1, false},
	}
	for _, tt := range tests {
		if got := Visible(DisplayItem{Y: tt.y}, scroll); got != tt.want {
			t.Errorf("%s: Visible(y=%d) = %v, want %v", tt.name, tt.y, got, tt.want)
		}
	}
}

func TestScrollDown(t *testing.T) {
	b := NewBrowser(nil)
	b.ScrollDown()
	b.ScrollDown()
	if b.scroll != 2*SCROLL_STEP {
		t.Errorf("scroll = %d, want %d", b.scroll, 2*SCROLL_STEP)
	}
}

func TestDrawable(t *testing.T) {
	if Drawable('\n') || Drawable('\t') {
		t.Error("control characters must not be drawn")
	}
	if !Drawable('西') || !Drawable('a') {
		t.Error("printable characters must be drawn")
	}
}
