/**
设备集，仅用于后台数据状态缓存和差分处理，不考虑消息格式问题


ver 1.0
*/
package device

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"strconv"
	"sync"
	"time"
)

/**
设备状态清单
*/
type DeviceSet struct {
	DeviceClass string             //设备类
	Devices     map[string]*Device //ID->设备
	//Interval    int64                    //消息时间间隔
	LastModifyTime int64 //上一次消息的时间
	//OnTime      func(*DeviceSet, []*Device)         //定时函数，
	//OnChange func(*DeviceSet, []*Device)
	RWLock sync.RWMutex
}

/**
设备清单
*/
type DeviceList struct {
	DeviceClass string           //设备类
	Devices     map[string]int64 //ID->时间戳
	RWLock      sync.RWMutex
}

/**
新建一个DeviceList对象
*/
func NewDeviceList(deviceClass string) *DeviceList {
	return &DeviceList{
		DeviceClass: deviceClass,
		Devices:     make(map[string]int64),
	}
}

/**
新建一个DeviceSet对象，Now和Last相同
*/
func NewDeviceSet(DeviceClass string) *DeviceSet {
	return &DeviceSet{
		DeviceClass:    DeviceClass,
		Devices:        make(map[string]*Device),
		LastModifyTime: -1,
	}
}

/**
发生修改时的回调函数
*/
type OnChangeFunction func(*DeviceSet, *Device) error

/**
获取设备ID列表
返回 ID数组
lock 是否加锁
*/
func (deviceSet *DeviceSet) GetIDs(lock bool) []string {
	if lock {
		(*deviceSet).RWLock.Lock()
		defer (*deviceSet).RWLock.Unlock()
	}
	ids := make([]string, 0)
	if (*deviceSet).Devices != nil && len((*deviceSet).Devices) > 0 {
		for id, _ := range (*deviceSet).Devices {
			ids = append(ids, id)
		}
	}
	return ids
}

/**
设置对象取值
device: 对象取值
lock 是否加锁
返回 列表（可能对象为空）
*/
func (deviceSet *DeviceSet) List(lock bool) *DeviceList {
	if lock {
		(*deviceSet).RWLock.Lock()
		defer (*deviceSet).RWLock.Unlock()
	}
	list := &DeviceList{
		DeviceClass: deviceSet.DeviceClass,
		Devices:     make(map[string]int64),
	}
	if (*deviceSet).Devices != nil && len((*deviceSet).Devices) > 0 {
		for id, device := range (*deviceSet).Devices {
			t := (*device).T
			(*list).Devices[id] = t
		}
	}
	return list
}

/**
设置对象取值
device: 对象取值
lock 是否加锁
返回（是否是新建对象，是否发生错误）
*/
func (deviceSet *DeviceSet) SetDevice(device *Device, lock bool) (bool, error) {
	if device == nil {
		return false, errors.New("device should not be a null pinter")
	}
	if lock {
		(*deviceSet).RWLock.Lock()
		defer (*deviceSet).RWLock.Unlock()
	}
	var err error = nil
	var id string = (*device).ID
	_, exists := (*deviceSet).Devices[id]
	(*deviceSet).Devices[id] = device
	//更新集合修改时间
	if (*device).T > (*deviceSet).LastModifyTime {
		(*deviceSet).LastModifyTime = (*device).T
	}

	return !exists, err
}

/**
删除超时的对象
currentTime 当前时间
timeout 超时时间
lock 是否加锁
返回（删除个数，删除清单）
*/
func (deviceSet *DeviceSet) RemoveTimeoutDevices(currentTime int64, timeout int64, lock bool) (int, []string) {
	if lock {
		(*deviceSet).RWLock.Lock()
		defer (*deviceSet).RWLock.Unlock()
	}
	var toDelete = make([]string, 0, 10)

	for _, device := range (*deviceSet).Devices {
		t := (*device).T
		id := (*device).ID
		if (currentTime - t) > timeout {
			toDelete = append(toDelete, id)
		}
	}
	for _, id := range toDelete {
		delete((*deviceSet).Devices, id)
	}
	return len(toDelete), toDelete
}

/**
设置超时对象样式
currentTime 当前时间
timeout 超时时间
status 色彩值
lock 是否加锁
返回：修改个数

*/
func (deviceSet *DeviceSet) TagTimeoutDevices(currentTime int64, timeout int64, status int, lock bool) (int, []string) {
	if lock {
		(*deviceSet).RWLock.Lock()
		defer (*deviceSet).RWLock.Unlock()
	}
	var toModify = make([]string, 0, 10)
	for _, device := range (*deviceSet).Devices {
		t := (*device).T
		id := (*device).ID
		if (currentTime - t) > timeout {
			(*device).Color = status
			toModify = append(toModify, id)
		}
	}
	return len(toModify), toModify
}

/**
设置超时对象样式 1，2，3档次
currentTime 当前时间
timeout 超时时间，要求降序排列
color 色彩值
lock 是否加锁
返回：（修改个数，修改后的对象，修改详情）

*/
func (deviceSet *DeviceSet) TagTimeoutDevices2(currentTime int64, timeout []int64, color []int, lock bool) (int, []*Device, []string) {
	if lock {
		(*deviceSet).RWLock.Lock()
		defer (*deviceSet).RWLock.Unlock()
	}
	toSet := make([]*Device, 0)
	if len(timeout) != len(color) {
		return 0, nil, nil
	}
	var n = len(timeout)
	var toModify = make([]string, 0, 10)

	for _, device := range (*deviceSet).Devices {
		t := (*device).T
		id := (*device).ID
		var i = 0
		for i = 0; i < n; i++ {
			//超时降序排列，超时多先判断，只记录真正修改的
			if (currentTime - t) >= timeout[i] {
				if (*device).Color != color[i] {
					(*device).Color = color[i]
					toSet = append(toSet, device)
					toModify = append(toModify, id+"-->"+strconv.Itoa(color[i]))
				}
				break
			}
		}
	}
	if len(toSet) == 0 {
		toSet = nil
	}
	return len(toModify), toSet, toModify
}

/**
删除超时的对象

timeout 超时时间
lock 是否锁定对象
返回 删除个数
*/
func (deviceList *DeviceList) RemoveTimeoutDevices(timeout int64, lock bool) int {
	if lock {
		(*deviceList).RWLock.Lock()
		defer (*deviceList).RWLock.Unlock()
	}
	var toDelete = make([]string, 0, 10)
	now := time.Now().Unix()
	for id, t := range (*deviceList).Devices {
		if now-t > timeout {
			toDelete = append(toDelete, id)
		}
	}
	for _, id := range toDelete {
		delete((*deviceList).Devices, id)
	}
	return len(toDelete)
}

/**
设置对象取值
devices: 对象取值
lock: 是否锁定对象
返回（新建对象数，修改对象数，是否发生错误）
*/
func (deviceSet *DeviceSet) SetDevices(devices []*Device, lock bool) (int, int, error) {
	if len(devices) == 0 {
		return 0, 0, nil
	}
	newCnt := 0
	modifyCnt := 0
	if lock {
		(*deviceSet).RWLock.Lock()
		defer (*deviceSet).RWLock.Unlock()
	}
	var err error = nil
	for _, device := range devices {
		var id string = (*device).ID
		_, exists := (*deviceSet).Devices[id]
		if exists {
			modifyCnt++
		} else {
			newCnt++
		}
		(*deviceSet).Devices[id] = device
		//更新集合修改时间
		if (*device).T > (*deviceSet).LastModifyTime {
			(*deviceSet).LastModifyTime = (*device).T
		}
	}
	return newCnt, modifyCnt, err
}

/**
删除对象
device: 对象id
lock: 是否锁定对象
返回：（是否找到，是否出错）
*/
func (deviceSet *DeviceSet) RemoveDevice(id string, lock bool) (bool, error) {
	if len(id) == 0 {
		return false, errors.New("len(id) should not be 0")
	}
	if lock {
		(*deviceSet).RWLock.Lock()
		defer (*deviceSet).RWLock.Unlock()
	}
	var err error = nil
	_, exists := (*deviceSet).Devices[id]
	if exists {
		delete((*deviceSet).Devices, id)
	}

	return !exists, err
}

/**
删除对象
devices: 对象id清单
lock: 是否锁定对象
返回：（删除成功个数，是否出错）
*/
func (deviceSet *DeviceSet) RemoveDevices(ids []string, lock bool) (int, error) {
	if lock {
		(*deviceSet).RWLock.Lock()
		defer (*deviceSet).RWLock.Unlock()
	}
	n := 0
	var err error = nil
	for _, id := range ids {
		_, exists := (*deviceSet).Devices[id]
		if exists {
			delete((*deviceSet).Devices, id)
			n++
		}
	}
	return n, err
}

/**
获取指定ID的device数据
lock: 是否锁定对象
*/
func (deviceSet *DeviceSet) GetDevice(id string, lock bool) *Device {
	if lock {
		(*deviceSet).RWLock.Lock()
		defer (*deviceSet).RWLock.Unlock()
	}
	var device, ok = (*deviceSet).Devices[id]
	if !ok {
		return nil
	} else {
		return device
	}
}

/**
获取当前的Device数组
lock: 是否锁定对象
*/
func (deviceSet *DeviceSet) GetDevices(lock bool) []*Device {
	if lock {
		(*deviceSet).RWLock.Lock()
		defer (*deviceSet).RWLock.Unlock()
	}
	var devices []*Device
	devices = make([]*Device, 0)
	for _, device := range (*deviceSet).Devices {
		devices = append(devices, device)
	}
	return devices
}

/**
清空当前的设备数组
lock: 是否锁定对象
*/
func (deviceSet *DeviceSet) Clear(lock bool) {
	if lock {
		(*deviceSet).RWLock.Lock()
		defer (*deviceSet).RWLock.Unlock()
	}
	(*deviceSet).Devices = make(map[string]*Device)
}

/**
保存当前状态到文件
lock: 是否锁定对象
*/
func (deviceSet *DeviceSet) Save(file string, lock bool) error {
	if lock {
		(*deviceSet).RWLock.Lock()
	}
	data, err := json.Marshal(*deviceSet)
	if lock { //尽快释放
		(*deviceSet).RWLock.Unlock()
	}

	if err != nil {
		return err
	}
	err = ioutil.WriteFile(file, data, 0777)
	return err
}

/**
读取状态到文件
*/
func Load(file string) (*DeviceSet, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	dSet := &DeviceSet{}
	err = json.Unmarshal(data, dSet)
	if err != nil {
		return nil, err
	} else {
		return dSet, nil
	}
}
