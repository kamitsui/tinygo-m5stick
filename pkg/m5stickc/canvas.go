package m5stickc

import "image/color"

// Canvas はメモリ上のオフスクリーン描画先（フレームバッファ）。
// ここに描いてから Display.DrawBuffer で一括転送すると、ちらつき無く
// アニメーションできる。drivers.Displayer 互換なので tinyfont でも描ける。
type Canvas struct {
	w, h int16
	buf  []color.RGBA
}

// NewCanvas は w×h のキャンバスを確保する。
func NewCanvas(w, h int16) *Canvas {
	return &Canvas{w: w, h: h, buf: make([]color.RGBA, int(w)*int(h))}
}

// Size はキャンバスのサイズを返す。
func (c *Canvas) Size() (int16, int16) { return c.w, c.h }

// SetPixel は1点を打つ（範囲外は無視）。
func (c *Canvas) SetPixel(x, y int16, col color.RGBA) {
	if x < 0 || y < 0 || x >= c.w || y >= c.h {
		return
	}
	c.buf[int(y)*int(c.w)+int(x)] = col
}

// Display は drivers.Displayer 互換のための no-op。
func (c *Canvas) Display() error { return nil }

// Fill は全体を1色で塗る。
func (c *Canvas) Fill(col color.RGBA) {
	for i := range c.buf {
		c.buf[i] = col
	}
}

// FillRectangle は矩形を塗る。
func (c *Canvas) FillRectangle(x, y, w, h int16, col color.RGBA) {
	for yy := y; yy < y+h; yy++ {
		for xx := x; xx < x+w; xx++ {
			c.SetPixel(xx, yy, col)
		}
	}
}

// Buffer は内部バッファ（row-major, 長さ w*h）を返す。Display.DrawBuffer に渡す。
func (c *Canvas) Buffer() []color.RGBA { return c.buf }
