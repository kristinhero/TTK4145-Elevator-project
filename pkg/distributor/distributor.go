package distributor

import (
	"Elevator/pkg/elevtypes"
	"Elevator/pkg/common"
	"Elevator/pkg/elevio"
	"Elevator/pkg/assigner"
)

/*
The distributor holds the states of the active and non active elevators on the network.
It updates the states based on messages from the local elevator and the network.
It (re)distributes the current requests when there is a new button press, 
an elevator is lost or an elevator has come back.
*/
func StateMachine(localID int, 
	localElevState <-chan elevtypes.Elevator, 
	globalElevState <-chan elevtypes.Elevator,
	newLocalRequests chan<- [common.NumFloors][common.NumButtons]int,
	newRequestDistribution <-chan map[int][common.NumFloors][common.NumButtons]int,
	distributedRequests chan<- map[int][common.NumFloors][common.NumButtons]int,
	spamLocalElevState chan<- elevtypes.Elevator,
	globalElevReqClearedFloor <-chan int,
	globalElevLost <-chan int,
	) {
	activeElevators := make(map[int]elevtypes.Elevator)
	nonActiveElevators := make(map[int]elevtypes.Elevator)
	newButtonPress := make(chan elevtypes.ButtonEvent, 1)
	go elevio.PollButtons(newButtonPress)
	for {
		select {
		case id := <- globalElevLost:
			nonActiveElevators[id] = activeElevators[id]
			delete(activeElevators, id)
			requestDistribution := assigner.AssignHallRequests(activeElevators)
			distributedRequests <- requestDistribution
			newLocalRequests <- requestDistribution[localID]
		case s := <- globalElevState:
			if _, nonActive := nonActiveElevators[s.ID]; nonActive {
				oldRequests := nonActiveElevators[s.ID].Requests
				updatedRequests := updateHallRequests(oldRequests, activeElevators[localID].Requests)
				mergedRequests := mergeRequests(s.Requests, updatedRequests)
				delete(nonActiveElevators, s.ID)
				s.Requests = mergedRequests
				activeElevators[s.ID] = s
				requestDistribution := assigner.AssignHallRequests(activeElevators)
				distributedRequests <- requestDistribution
				newLocalRequests <- requestDistribution[localID]	
			} else {
				activeElevators[s.ID] = s
			}
		case f := <- globalElevReqClearedFloor:
			localRequests := activeElevators[localID].Requests
			for btn := elevtypes.ButtonType(0); btn < common.NumButtons - 1; btn++ {
				localRequests[f][btn] = 0
			}
			newLocalRequests <- localRequests
		case s := <- localElevState:
			activeElevators[s.ID] = s
			spamLocalElevState <- s
		case r:= <- newRequestDistribution:
			if _,forUs := r[localID]; forUs {
				newLocalRequests <- r[localID]
			}
		case b := <- newButtonPress:
			temp := activeElevators[localID]
			temp.Requests[b.Floor][b.Button] = localID
			activeElevators[localID] = temp
			requestDistribution := assigner.AssignHallRequests(activeElevators)
			distributedRequests <- requestDistribution
			newLocalRequests <- requestDistribution[localID]
		}
	}
}

/*
Takes in an old request array and a new request array.
Returns an updated request array with the old cab requests and new hall requests.
*/
func updateHallRequests(requests[common.NumFloors][common.NumButtons]int,
	newRequests[common.NumFloors][common.NumButtons]int,
	) [common.NumFloors][common.NumButtons]int {	
	updatedRequests := requests
	for floor := 0; floor < common.NumFloors; floor++ {
		for btn := elevtypes.ButtonType(0); btn < common.NumButtons - 1; btn++ {
			updatedRequests[floor][btn] = newRequests[floor][btn]
		}
	}
	return updatedRequests
}

/*
Takes in an old request array and a new request array.
Returns a merged request array of the two. 
*/
func mergeRequests(oldRequests[common.NumFloors][common.NumButtons]int,
newRequests[common.NumFloors][common.NumButtons]int,
) [common.NumFloors][common.NumButtons]int {
	mergedRequests := oldRequests
	for floor := 0; floor < common.NumFloors; floor++ {
		for btn := elevtypes.ButtonType(0); btn < common.NumButtons; btn++ {
			if newRequests[floor][btn] > 0 {
				mergedRequests[floor][btn] = newRequests[floor][btn]
			}
		}
	}
	return mergedRequests
}
