package m5stickc

import (
	"machine"

	"tinygo.org/x/drivers/mpu6886"
)

// Dir は画面座標系での方向（チルトやカーソルに使う）。
type Dir uint8

const (
	DirNone Dir = iota
	DirUp
	DirDown
	DirLeft
	DirRight
)

// IMU は内蔵6軸センサ MPU6886 を、画面の自然な向き（[[Display]] と同じ）に
// 合わせて扱うラッパ。加速度の X 軸は画面向きに対して反転しているため、
// ここで符号を補正し「+X=右 / +Y=下」（画面論理座標と同じ）で返す。
type IMU struct {
	dev *mpu6886.Device
}

// NewIMU は内部 I2C を設定し MPU6886 を初期化する。
func NewIMU() (*IMU, error) {
	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: 400 * machine.KHz,
		SDA:       imuSDA,
		SCL:       imuSCL,
	})
	dev := mpu6886.New(machine.I2C0)
	if err := dev.Configure(mpu6886.Config{}); err != nil {
		return nil, err
	}
	return &IMU{dev: dev}, nil
}

// Acceleration は加速度を画面座標系（+X=右, +Y=下）の µg で返す（1g≈1e6）。
func (i *IMU) Acceleration() (x, y, z int32) {
	ax, ay, az, err := i.dev.ReadAcceleration()
	if err != nil {
		return 0, 0, 0
	}
	return -ax, ay, az // X 軸は画面向きに合わせて反転
}

// Tilt は支配的な傾き方向を画面座標系で返す。水平に近い（どの軸も
// thresholdMicroG 未満）場合は DirNone。閾値の目安: 300000 (=0.3g)。
func (i *IMU) Tilt(thresholdMicroG int32) Dir {
	x, y, _ := i.Acceleration()
	ax, ay := abs32(x), abs32(y)
	if ax < thresholdMicroG && ay < thresholdMicroG {
		return DirNone
	}
	if ax >= ay {
		if x > 0 {
			return DirRight
		}
		return DirLeft
	}
	if y > 0 {
		return DirDown
	}
	return DirUp
}

func abs32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}
