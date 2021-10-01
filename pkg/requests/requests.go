package requests

import (
	"Elevator/pkg/common"
	"Elevator/pkg/elevtypes"
)

/*
Takes in an elevator.
Returns the direction the elevator should move in.
*/
func ChooseDirection(elevator elevtypes.Elevator) elevtypes.MotorDirection {
	switch elevator.Dirn {
	case elevtypes.MD_Up:
		if above(elevator) {
			return elevtypes.MD_Up
		} else if below(elevator) {
			return elevtypes.MD_Down
		} else {
			return elevtypes.MD_Stop
		}
	default:
		if below(elevator) {
			return elevtypes.MD_Down
		} else if above(elevator) {
			return elevtypes.MD_Up
		} else {
			return elevtypes.MD_Stop
		}
	}
}

/*
Takes in an elevator.
Returns true if the elevator should stop at the current floor.
*/
func ShouldStop(elevator elevtypes.Elevator) bool {
	switch elevator.Dirn {
	case elevtypes.MD_Down:
		return elevator.Requests[elevator.Floor][elevtypes.BT_HallDown] == elevator.ID ||
			elevator.Requests[elevator.Floor][elevtypes.BT_Cab] == elevator.ID ||
			!below(elevator)
	default:
		return elevator.Requests[elevator.Floor][elevtypes.BT_HallUp] == elevator.ID ||
			elevator.Requests[elevator.Floor][elevtypes.BT_Cab] == elevator.ID ||
			!above(elevator)
	}
}

/*
Takes in an elevator.
Returns true if there are any requests at its current floor.
*/
func RequestsAtCurrentFloor(elevator elevtypes.Elevator) bool {
	requestsAtCurrentFloor := false
	for btn := elevtypes.ButtonType(0); btn < common.NumButtons; btn++ {
		if elevator.Requests[elevator.Floor][btn] > 0 {
			requestsAtCurrentFloor = true
		}
	}
	return requestsAtCurrentFloor
}

/*
Takes in an elevator. 
Returns an elevator without requests at its current floor.
*/
func ClearAtCurrentFloor(elevator elevtypes.Elevator) elevtypes.Elevator {
	for btn := 0; btn < common.NumButtons; btn++ {
		elevator.Requests[elevator.Floor][btn] = 0
		
	}
	return elevator
}

/*
Takes in an elevator.
Returns true if the elevator has any requests above its current floor.
*/
func above(elevator elevtypes.Elevator) bool {
	for f := elevator.Floor + 1; f < common.NumFloors; f++ {
		for btn := 0; btn < common.NumButtons; btn++ {
			if elevator.Requests[f][btn] == elevator.ID {
				return true
			}
		}
	}
	return false
}

/*
Takes in an elevator.
Returns true if the elevator has any requests below its current floor.
*/
func below(elevator elevtypes.Elevator) bool {
	for f := 0; f < elevator.Floor; f++ {
		for btn := 0; btn < common.NumButtons; btn++ {
			if elevator.Requests[f][btn] == elevator.ID {
				return true
			}
		}
	}
	return false
}