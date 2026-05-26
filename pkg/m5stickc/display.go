package m5stickc

import (
	"image/color"
	"machine"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/st7789"
)

// 表示器（ST7789V2, 1.14" 135x240）の論理サイズ。
const (
	DisplayWidth  = 135
	DisplayHeight = 240
)

// Display は M5StickC Plus2 の表示器を「自然な向き」（USB-C を下にした持ち方で
// 左上原点・y 下向き）で扱えるようにするラッパ。
//
// 背景: tinygo.org/x/drivers の st7789 ドライバは Rotation0 のとき
// RowOffset/ColumnOffset を 0 に潰してしまい、このパネル（GRAM 原点からずれて
// 実装されている）では位置がずれる。オフセットが効くのは Rotation180 系のみ
// だが、それだと USB-C を上にして持つ必要があり、ボタンと向きが噛み合わない。
//
// そこで内部は実績のある Rotation180 + ColumnOffset 53 / RowOffset 40 のまま、
// 論理座標を 180° 反転して公開する。これで USB-C を下にした自然な持ち方で、
// 文字も図形も正立して表示される。
type Display struct {
	dev st7789.Device
}

// NewDisplay は SPI2 と ST7789V2 を初期化し、自然な向きの *Display を返す。
//
// 例:
//
//	d := m5stickc.NewDisplay()
//	d.FillScreen(color.RGBA{0, 0, 0, 255})
//	tinyfont.WriteLine(d, &freemono.Bold9pt7b, 8, 20, "Hello", white)
func NewDisplay() *Display {
	machine.SPI2.Configure(machine.SPIConfig{
		Frequency: 40e6,
		SCK:       tftSCK,
		SDO:       tftSDO,
		SDI:       machine.NoPin,
	})

	dev := st7789.New(machine.SPI2, tftRST, tftDC, tftCS, tftBL)
	dev.Configure(st7789.Config{
		Width:        DisplayWidth,
		Height:       DisplayHeight,
		Rotation:     drivers.Rotation180,
		RowOffset:    40,
		ColumnOffset: 53,
	})
	return &Display{dev: dev}
}

// Size は論理サイズ (135x240) を返す。drivers.Displayer 互換（tinyfont 等で使える）。
func (d *Display) Size() (int16, int16) {
	return DisplayWidth, DisplayHeight
}

// SetPixel は自然な向きの論理座標に1点を打つ。
func (d *Display) SetPixel(x, y int16, c color.RGBA) {
	d.dev.SetPixel(DisplayWidth-1-x, DisplayHeight-1-y, c)
}

// FillRectangle は自然な向きの論理座標で矩形を塗る。
func (d *Display) FillRectangle(x, y, w, h int16, c color.RGBA) error {
	return d.dev.FillRectangle(DisplayWidth-x-w, DisplayHeight-y-h, w, h, c)
}

// FillScreen は全画面を塗る（全面なので反転は不要）。
func (d *Display) FillScreen(c color.RGBA) {
	d.dev.FillScreen(c)
}

// Display は drivers.Displayer 互換のためのメソッド（直接描画なので実質 no-op）。
func (d *Display) Display() error {
	return d.dev.Display()
}
