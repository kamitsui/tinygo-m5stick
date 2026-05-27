// cmd/invaders は Invaders を単体で書き込むための薄いラッパー。 make flash PROJ=invaders
package main

import (
	"github.com/kamitsui/tinygo-m5stick/games/invaders"
	"github.com/kamitsui/tinygo-m5stick/pkg/m5stickc"
)

func main() {
	con := m5stickc.NewConsole()
	imu, _ := m5stickc.NewIMU()
	invaders.Run(con, imu)
}
