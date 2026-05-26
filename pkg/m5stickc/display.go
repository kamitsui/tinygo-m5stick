package m5stickc

import (
	"machine"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/st7789"
)

// 表示器（ST7789V2, 1.14" 135x240）のサイズ。
const (
	DisplayWidth  = 135
	DisplayHeight = 240
)

// NewDisplay は SPI2 を設定し ST7789V2 を初期化して返す。
//
// 重要: tinygo.org/x/drivers の st7789 は Rotation0 のとき RowOffset/ColumnOffset を
// 0 に潰すため、このパネルでは位置がずれる。実機検証済みの Rotation180 +
// ColumnOffset 53 / RowOffset 40 を使う（USB-C を上にすると正立する）。
//
// 例:
//
//	display := m5stickc.NewDisplay()
//	display.FillScreen(color.RGBA{0, 0, 0, 255})
func NewDisplay() st7789.Device {
	machine.SPI2.Configure(machine.SPIConfig{
		Frequency: 40e6,
		SCK:       tftSCK,
		SDO:       tftSDO,
		SDI:       machine.NoPin,
	})

	display := st7789.New(machine.SPI2, tftRST, tftDC, tftCS, tftBL)
	display.Configure(st7789.Config{
		Width:        DisplayWidth,
		Height:       DisplayHeight,
		Rotation:     drivers.Rotation180,
		RowOffset:    40,
		ColumnOffset: 53,
	})
	return display
}
