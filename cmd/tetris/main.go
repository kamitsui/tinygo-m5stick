// cmd/tetris は Tetris を単体で書き込むための薄いラッパー。 make flash PROJ=tetris
package main

import (
	"github.com/kamitsui/tinygo-m5stick/games/tetris"
	"github.com/kamitsui/tinygo-m5stick/pkg/m5stickc"
)

func main() {
	con := m5stickc.NewConsole()
	imu, _ := m5stickc.NewIMU()
	tetris.Run(con, imu)
}
