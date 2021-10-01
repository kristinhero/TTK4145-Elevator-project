package elevtypes

import "Elevator/pkg/common"

type MotorDirection int

const (
	MD_Up   MotorDirection = 1
	MD_Down                = -1
	MD_Stop                = 0
)

type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown            = 1
	BT_Cab                 = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

type ElevatorState int

const (
	ES_Idle     ElevatorState = 0
	ES_DoorOpen               = 1
	ES_Moving                 = 2
)

type Elevator struct {
	ID int
	Floor          int
	Dirn           MotorDirection
	Requests       [common.NumFloors][common.NumButtons] int
	State          ElevatorState
	DoorObstructed bool
	Stuck 		   bool
	Config         struct {
		DoorOpenDuration_s int
	}
}

func Uninitialized(ID int) Elevator {
	elevator := Elevator {
		ID: ID,
		Floor: -1,
		Dirn:  MD_Stop,
		State: ES_Idle,
	}
	elevator.Config.DoorOpenDuration_s = 3
	return elevator
}

/*
Returns true if two elevators are identical.
*/
func ElevatorsAreEqual(A Elevator, B Elevator) bool {
	if A.ID != B.ID { return false}
	if A.Floor != B.Floor {return false}
	if A.Dirn != B.Dirn {return false}
	if !requestsAreEqual(A.Requests, B.Requests) {return false}
	if A.State != B.State {return false}
	if A.DoorObstructed != B.DoorObstructed {return false}
	if A.Config.DoorOpenDuration_s != B.Config.DoorOpenDuration_s {return false}
	return true
}

/*
Returns true if two request matrices are identical.
*/
func requestsAreEqual(A, B [common.NumFloors][common.NumButtons]int) bool {
	for i:=0; i< common.NumFloors; i++ {
		for j:=0; j< common.NumButtons; j++ {
			if A[i][j] != B[i][j] {
				return false
			}
		}
	}
	return true
}