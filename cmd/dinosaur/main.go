// cmd/dinosaur は Dinosaur を単体で書き込むための薄いラッパー。 make flash PROJ=dinosaur
package main

import (
	"github.com/kamitsui/tinygo-m5stick/games/dinosaur"
	"github.com/kamitsui/tinygo-m5stick/pkg/m5stickc"
)

func main() {
	con := m5stickc.NewConsole()
	imu, _ := m5stickc.NewIMU()
	dinosaur.Run(con, imu)
}
