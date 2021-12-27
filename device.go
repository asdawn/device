/*
设备信息
ver 1.2 拆分Color为高亮+颜色代码
ver 1.1 增加额外时间戳、原始坐标
ver 1.0
*/
package device

import (
	"encoding/json"
	"errors"
	"math"
	"strconv"
)

/*
存储的设备信息，不考虑渲染问题，T采用东八区字符串格式
*/
type Device1 struct {
	ID     string  `json:"id"`
	ORGX   float32 `json:"orgx"`
	ORGY   float32 `json:"orgy"`
	X      float32 `json:"x"`
	Y      float32 `json:"y"`
	R      float32 `json:"r"`
	Status int     `json:"s"`
	Color  int     `json:"c"`
	T      string  `json:"t"`  //设备消息产生时间
	TM     int64   `json:"tm"` //最终移动时间
	T1     int64   `json:"t1"` //设备消息接收时间1
	T2     int64   `json:"t2"` //设备消息转发时间1
	T3     int64   `json:"t3"` //设备消息接收时间2
	T4     int64   `json:"t4"` //设备消息转发时间2
	T5     int64   `json:"t5"` //设备消息接收时间5
}

/*
后台存储的设备信息，不考虑渲染问题
*/
type Device struct {
	ID     string  `json:"id"`
	ORGX   float32 `json:"orgx"`
	ORGY   float32 `json:"orgy"`
	X      float32 `json:"x"`
	Y      float32 `json:"y"`
	R      float32 `json:"r"`
	Status int     `json:"s"`
	Color  int     `json:"c"`
	T      int64   `json:"t"`  //设备消息产生时间
	TM     int64   `json:"tm"` //最终移动时间
	T1     int64   `json:"t1"` //设备消息接收时间1
	T2     int64   `json:"t2"` //设备消息转发时间1
	T3     int64   `json:"t3"` //设备消息接收时间2
	T4     int64   `json:"t4"` //设备消息转发时间2
	T5     int64   `json:"t5"` //设备消息接收时间5
}

/**
计算状态值（替换s的指定位）
digit: 位数
value: 值
*/
func SetStatus(s int, digit int, value int) (int, error) {
	if s < 0 {
		return 0, errors.New("invalid status code: " + strconv.Itoa(s))
	}
	if value > 10 || value < 0 {
		return 0, errors.New("invalid status value: " + strconv.Itoa(value))
	}
	if digit < 0 || digit > 9 { //预留10个状态位
		return 0, errors.New("invalid status digit: " + strconv.Itoa(digit))
	}
	var exp = int(math.Pow10(digit))

	s_rem := s % exp         //余数
	s_main := s / (exp * 10) //前边几位
	//s_digit := s/exp - s_main*10 //指定位
	newS := (s_main*10+value)*exp + s_rem
	return newS, nil
}

/**
计算状态值（替换s的指定位）
digit: 位数
value: 值
*/
func GetStatus(s int, digit int) (int, error) {
	if s < 0 {
		return 0, errors.New("invalid status code: " + strconv.Itoa(s))
	}
	if digit < 0 || digit > 9 { //预留10个状态位
		return 0, errors.New("invalid status digit: " + strconv.Itoa(digit))
	}
	var exp = int(math.Pow10(digit))
	s_main := s / (exp * 10)     //前边几位
	s_digit := s/exp - s_main*10 //指定位
	return s_digit, nil
}

/**
设置是否高亮（Color的倒数第二位）
highlight: 是否高亮
返回：错误
*/
func (device *Device) SetHighlight(highlight bool) error {
	var value int = 0
	if highlight {
		value = 1
	}
	newS, err := SetStatus(device.Color, 1, value)
	if err != nil {
		return err
	}
	device.Color = newS
	return nil
}

/**
设置颜色代码（Color的最低位）
color: 颜色代码，0-9，0一般为默认色
返回：错误
*/
func (device *Device) SetColor(color int) error {
	if color < 0 || color > 9 {
		return errors.New("invalid color code: " + strconv.Itoa(color))
	}
	newS, err := SetStatus(device.Color, 0, color)
	if err != nil {
		return err
	}
	device.Color = newS
	return nil
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
