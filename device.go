/*
设备信息
ver 1.0
*/
package device

import (
	"encoding/json"
	"errors"
	"math"
)

/*
存储的设备信息，不考虑渲染问题，T采用东八区字符串格式
*/
type Device1 struct {
	ID     string  `json:"id"`
	X      float32 `json:"x"`
	Y      float32 `json:"y"`
	R      float32 `json:"r"`
	Status int     `json:"s"`
	Color  int     `json:"c"`
	T      string  `json:"t"`
}

/*
后台存储的设备信息，不考虑渲染问题
*/
type Device struct {
	ID     string  `json:"id"`
	X      float32 `json:"x"`
	Y      float32 `json:"y"`
	R      float32 `json:"r"`
	Status int     `json:"s"`
	Color  int     `json:"c"`
	T      int64   `json:"t"`
}

/**
判断对象是否移动（包括旋转角度）
device0: 对比目标
*/
func (device *Device) Moved(device0 *Device) bool {
	if device == nil || device0 == nil {
		return false
	}
	return (device.X == device0.X && device.Y == device0.Y && device.R == device0.R)
}

/**
计算对象变化量(距离, 角度, 时间,err)
距离单位近似为米
*/
func (device *Device) GetDelta(device0 *Device) (float64, float32, int64, error) {
	if device == nil || device0 == nil {
		return 0, 0, 0, errors.New("null pointer")
	}
	deltaX := device.X - device0.X
	deltaY := device.Y - device0.Y
	dist := math.Sqrt(float64(deltaX*deltaX+deltaY*deltaY)) * 100000
	deltaR := device.R - device0.R
	deltaT := device.T - device0.T
	return dist, deltaR, deltaT, nil
}

/**
根据坐标变化计算两个对象之间的夹角
*/
func (device *Device) GetAngle(device0 *Device) (float32, error) {
	if device == nil || device0 == nil {
		return 0, errors.New("null pointer")
	}
	deltaX := device.X - device0.X
	deltaY := device.Y - device0.Y
	if deltaX == 0 && deltaY == 0 {
		return 0, errors.New("not moved")
	}
	r := float32(math.Atan2(float64(deltaY), float64(deltaX)))
	return r, nil
}

/**
获取对象的JSON
*/
func (device *Device) JSON() ([]byte, error) {
	return json.Marshal(device)
}

/**
解析对象的JSON
*/
func Parse(jsonBytes []byte) (Device, error) {
	var device Device
	err := json.Unmarshal(jsonBytes, &device)
	if err != nil {
		return device, err
	} else {
		return device, nil
	}
}

/**
解析对象的JSON，缺失值使用指定的默认值
jsonBytes: JSON数据
defaultVlue: 默认值
*/
func ParseWithDefaults(jsonBytes []byte, defaultValue Device) (Device, error) {
	var device = defaultValue
	err := json.Unmarshal(jsonBytes, &device)
	if err != nil {
		return device, err
	} else {
		return device, nil
	}
}
