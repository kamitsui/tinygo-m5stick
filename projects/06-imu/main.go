// Project 06: IMU (MPU6886) 検証
//
// M5StickC Plus2 内蔵の6軸IMU (MPU6886) を I2C で読み、加速度(傾き)を画面表示する。
// 2048 / Tetris のチルト操作が実現できるかの確認用。
//
// 内蔵 I2C: SDA=GPIO21, SCL=GPIO22。MPU6886 アドレスは 0x68（ドライバ既定）。
//
// 書き込み: make flash PROJ=06-imu
package main

import (
	"image/color"
	"machine"
	"strconv"
	"time"

	"github.com/kamitsui/tinygo-m5stick/pkg/m5stickc"
	"tinygo.org/x/drivers/mpu6886"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
)

var (
	black = color.RGBA{0, 0, 0, 255}
	white = color.RGBA{255, 255, 255, 255}
	green = color.RGBA{0, 220, 0, 255}
	red   = color.RGBA{230, 40, 40, 255}
	blue  = color.RGBA{0, 0, 160, 255}
)

func main() {
	m5stickc.HoldPower()
	display := m5stickc.NewDisplay()
	display.FillScreen(black)
	display.FillRectangle(0, 0, 135, 28, blue)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 8, 20, "IMU test", white)

	// 内蔵 I2C を設定。
	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: 400 * machine.KHz,
		SDA:       machine.GPIO21,
		SCL:       machine.GPIO22,
	})

	imu := mpu6886.New(machine.I2C0)
	err := imu.Configure(mpu6886.Config{})

	if err != nil || !imu.Connected() {
		tinyfont.WriteLine(display, &freemono.Bold9pt7b, 8, 80, "NOT FOUND", red)
		for {
			time.Sleep(time.Second)
		}
	}
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 8, 60, "connected", green)

	var prevLabel string
	for {
		ax, ay, az, e := imu.ReadAcceleration()
		if e != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		// µg を milli-g に丸めて表示（静止時、ある軸が約 ±1000）。
		drawVal(display, 100, "X", ax/1000)
		drawVal(display, 125, "Y", ay/1000)
		drawVal(display, 150, "Z", az/1000)

		// 支配的な傾き方向（しきい値 0.3g）。
		label := tilt(ax, ay)
		if label != prevLabel {
			display.FillRectangle(0, 185, 135, 30, black)
			tinyfont.WriteLine(display, &freemono.Bold12pt7b, 8, 210, label, green)
			prevLabel = label
		}
		time.Sleep(80 * time.Millisecond)
	}
}

func tilt(ax, ay int32) string {
	const th = 300000 // 0.3g (µg)
	if abs(ax) < th && abs(ay) < th {
		return "FLAT"
	}
	if abs(ax) >= abs(ay) {
		// X 軸は画面の自然な向きに対して反転している（実機確認）。
		if ax > 0 {
			return "LEFT"
		}
		return "RIGHT"
	}
	if ay > 0 {
		return "DOWN"
	}
	return "UP"
}

func abs(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

func drawVal(d *m5stickc.Display, y int16, label string, mg int32) {
	d.FillRectangle(0, y-16, 135, 22, black)
	tinyfont.WriteLine(d, &freemono.Bold9pt7b, 8, y, label+": "+strconv.Itoa(int(mg)), white)
}
