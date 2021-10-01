package elevio

import (
	"net"
	"sync"
	"time"
	"Elevator/pkg/common"
	"Elevator/pkg/elevtypes"
)

const _pollRate = 20 * time.Millisecond

var _initialized bool = false
var _mtx sync.Mutex
var _conn net.Conn

func Init(addr string) {
	if _initialized {
		return
	}
	_mtx = sync.Mutex{}
	var err error
	_conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	clearAllLights()
	_initialized = true
}

func GetInitialFloor() int {
	return getFloor()
}

func SetMotorDirection(dir elevtypes.MotorDirection) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{1, byte(dir), 0, 0})
}

func SetButtonLamp(button elevtypes.ButtonType, floor int, value bool) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{2, byte(button), byte(floor), toByte(value)})
}

func SetFloorIndicator(floor int) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{3, byte(floor), 0, 0})
}

func SetDoorOpenLamp(value bool) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{4, toByte(value), 0, 0})
}

func PollButtons(receiver chan<- elevtypes.ButtonEvent) {
	prev := make([][3]bool, common.NumFloors)
	for {
		time.Sleep(_pollRate)
		for f := 0; f < common.NumFloors; f++ {
			for b := elevtypes.ButtonType(0); b < common.NumButtons; b++ {
				v := getButton(b, f)
				if v != prev[f][b] && v != false {
					receiver <- elevtypes.ButtonEvent{f, elevtypes.ButtonType(b)}
				}
				prev[f][b] = v
			}
		}
	}
}

func PollFloorSensor(receiver chan<- int) {
	prev := -1
	for {
		time.Sleep(_pollRate)
		v := getFloor()
		if v != prev && v != -1 {
			receiver <- v
		}
		prev = v
	}
}

func PollObstructionSwitch(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := getObstruction()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func SetAllLights(elevator elevtypes.Elevator) {
	for floor := 0; floor < common.NumFloors; floor++ {
		for btn := elevtypes.ButtonType(0); btn < common.NumButtons; btn++ {
			SetButtonLamp(btn, floor, elevator.Requests[floor][btn] > 0)
		}
	}
}

func clearAllLights() {
	for floor := 0; floor < common.NumFloors; floor++ {
		for btn := elevtypes.ButtonType(0); btn < elevtypes.ButtonType(common.NumButtons); btn++ {
			SetButtonLamp(btn, floor, false)
		}
	}
	SetDoorOpenLamp(false)
}

func getButton(button elevtypes.ButtonType, floor int) bool {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{6, byte(button), byte(floor), 0})
	var buf [4]byte
	_conn.Read(buf[:])
	return toBool(buf[1])
}

func getFloor() int {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{7, 0, 0, 0})
	var buf [4]byte
	_conn.Read(buf[:])
	if buf[1] != 0 {
		return int(buf[2])
	} else {
		return -1
	}
}

func getObstruction() bool {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{9, 0, 0, 0})
	var buf [4]byte
	_conn.Read(buf[:])
	return toBool(buf[1])
}

func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}
