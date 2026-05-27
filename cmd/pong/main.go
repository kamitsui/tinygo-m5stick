// cmd/pong は Pong を単体で書き込むための薄いラッパー。 make flash PROJ=pong
package main

import (
	"github.com/kamitsui/tinygo-m5stick/games/pong"
	"github.com/kamitsui/tinygo-m5stick/pkg/m5stickc"
)

func main() {
	con := m5stickc.NewConsole()
	imu, _ := m5stickc.NewIMU()
	pong.Run(con, imu)
}
