// cmd/snake は Snake を単体で書き込むための薄いラッパー。
//
//	make flash PROJ=snake
package main

import (
	"github.com/kamitsui/tinygo-m5stick/games/snake"
	"github.com/kamitsui/tinygo-m5stick/pkg/m5stickc"
)

func main() {
	con := m5stickc.NewConsole()
	imu, _ := m5stickc.NewIMU() // Snake は使わないが共通ラッパーに合わせる
	snake.Run(con, imu)
}
