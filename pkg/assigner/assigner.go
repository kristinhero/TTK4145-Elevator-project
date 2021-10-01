package assigner

import (
	"Elevator/pkg/common"
	"Elevator/pkg/elevtypes"
	"encoding/json"
	"os/exec"
	"os"
	"strconv"
)

type elevator struct {
	Behaviour string `json:"behaviour"`
	Floor int `json:"floor"`
	Direction string `json:"direction"`
	CabRequests [common.NumFloors]bool `json:"cabRequests"`
}

type input struct {
	HallRequests [common.NumFloors][common.NumButtons - 1]bool `json:"hallRequests"`
	States map[string]elevator `json:"states"`
}

/*
Takes a map of n elevators and optimally assigns hall requests. 
Returns a map of request distribution.
*/
func AssignHallRequests(elevators map[int]elevtypes.Elevator,
	) map[int][common.NumFloors][common.NumButtons]int {
	jsoninput,_ := json.Marshal(encodeInput(elevators))
	var jsonoutput []byte
	path,_ := os.Getwd()
	jsonoutput,_ = exec.Command(path + "/hall_request_assigner","-e","--input",string(jsoninput),"--clearRequestType","all").Output()
	output := make(map[string][][]bool)
	json.Unmarshal(jsonoutput, &output)
	requests := decodeOutput(output, elevators)
	return requests
} 

/*
Takes a map of n elevators.
Returns a struct with a map of the states of the elevators and matrix of hall requests.
*/
func encodeInput(elevators map[int]elevtypes.Elevator) input {
	input := input{States: make(map[string]elevator) }
	for ID, elevator := range elevators {
		input.States[strconv.Itoa(ID)] = newElevatorStruct(elevator)
		for floor := 0; floor < common.NumFloors; floor++ {
			for btn := 0; btn < common.NumButtons - 1; btn++ {
				if elevator.Requests[floor][btn] > 0 {
					input.HallRequests[floor][btn] = true
				}
			}
		}
	}
	return input
}

/*
Takes a map of hall requests for n elevators.
Returns a map of all requests for the elevators.
*/
func decodeOutput(output map[string][][]bool, 
	elevators map[int]elevtypes.Elevator,
	) map[int][common.NumFloors][common.NumButtons]int {
	requestDistribution := make(map[int][common.NumFloors][common.NumButtons]int)
	var requestsNoCab [common.NumFloors][common.NumButtons]int
	for IDstring, hallRequests := range output {
		for floor := 0; floor < common.NumFloors; floor++ {
			for btn := 0; btn < common.NumButtons -1; btn++ {
				if hallRequests[floor][btn] {
					requestsNoCab[floor][btn],_ = strconv.Atoi(IDstring) 
				}
			}
		}
	}
	for ID, _ := range elevators {
		requestDistribution[ID] = addCabOrders(ID, requestsNoCab, elevators)
	}
	return requestDistribution
}

/*
Takes a matrix of requests without cab requests, a map of elevators and an elevator ID.
Returns all the requests of the elevator.
*/
func addCabOrders(ID int, 
	requestsNoCab [common.NumFloors][common.NumButtons]int, 
	elevators map[int]elevtypes.Elevator,
	) [common.NumFloors][common.NumButtons]int {
	newRequests := requestsNoCab
	for floor := 0; floor < common.NumFloors; floor++ {
		if elevators[ID].Requests[floor][elevtypes.BT_Cab] == ID {
			newRequests[floor][elevtypes.BT_Cab] = ID
		}
	}
	return newRequests
}

/*
Takes an elevator from elevtypes.
Returns an elevator of the local type.
*/
func newElevatorStruct(oldElevator elevtypes.Elevator) elevator {
	newElevator := elevator{Floor: oldElevator.Floor}
	for floor := 0; floor < common.NumFloors; floor++ {
		if oldElevator.Requests[floor][elevtypes.BT_Cab] == oldElevator.ID {
			newElevator.CabRequests[floor] = true
		}
	}
	switch oldElevator.State {
	case elevtypes.ES_Idle:
		newElevator.Behaviour = "idle"
	case elevtypes.ES_Moving:
		newElevator.Behaviour = "moving"
	case elevtypes.ES_DoorOpen:
		newElevator.Behaviour = "doorOpen"
	}
	switch oldElevator.Dirn {
	case elevtypes.MD_Up:
		newElevator.Direction = "up"
	case elevtypes.MD_Down:
		newElevator.Direction = "down"
	case elevtypes.MD_Stop:
		newElevator.Direction = "stop"
	}
	return newElevator
}