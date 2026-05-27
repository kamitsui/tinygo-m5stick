// cmd/2048 は 2048 を単体で書き込むための薄いラッパー。 make flash PROJ=2048
package main

import (
	g2048 "github.com/kamitsui/tinygo-m5stick/games/2048"
	"github.com/kamitsui/tinygo-m5stick/pkg/m5stickc"
)

func main() {
	con := m5stickc.NewConsole()
	imu, _ := m5stickc.NewIMU()
	g2048.Run(con, imu)
}
