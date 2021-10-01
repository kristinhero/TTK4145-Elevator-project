package network

import (
	"Elevator/pkg/network/bcast"
	"Elevator/pkg/elevtypes"
	"time"
	"Elevator/pkg/common"
)

type StateMessage struct {
	Elevator elevtypes.Elevator
	RequestDistribution map[int][common.NumFloors][common.NumButtons]int
	RequestsClearedFloor int
	ID, Iter, RDIter, RCFIter int
}

func Init(id int, 
	enableTransmit <-chan bool,
	localElevState <-chan elevtypes.Elevator, 
	globalElevState chan<- elevtypes.Elevator,
	distributedRequests <-chan map[int][common.NumFloors][common.NumButtons]int,
	newRequestDistribution chan<- map[int][common.NumFloors][common.NumButtons]int,
	localElevReqClearedFloor <-chan int,
	globalElevReqClearedFloor chan<- int,
	globalElevLost chan<- int,
	) {
	go SendState(enableTransmit, localElevState, distributedRequests, localElevReqClearedFloor)
	go ReceiveStates(id, globalElevState, newRequestDistribution, globalElevReqClearedFloor, globalElevLost)
}

/*
Sends the state of the local elevator to other elevators.
*/
func SendState(enableTransmit <-chan bool,
	localElevState <-chan elevtypes.Elevator, 
	distributedRequests <-chan map[int][common.NumFloors][common.NumButtons]int,
	localElevReqClearedFloor <-chan int,
	) {
	stateMessageTx := make(chan StateMessage, 1)
	go bcast.Transmitter(16569, stateMessageTx)
	counter := 0
	RDcounter := 0
	RCFcounter := 0	
	var localElevator elevtypes.Elevator
	requestDistribution :=  make(map[int][common.NumFloors][common.NumButtons]int)
	requestsClearedFloor := -1
	enable := true
	for {
		counter++
		select {
		// Force prioritizes sending the request distribution to other elevators.
		case r := <-distributedRequests:
			requestDistribution = r
			RDcounter++
		default:
			select {
			case e := <-enableTransmit:
				enable = e
				if enable {
					counter = 0
					RDcounter = 0
					RCFcounter = 0
				}
			case r := <-distributedRequests:
				requestDistribution = r
				RDcounter++
			case s := <-localElevState:
				localElevator = s
			case f := <-localElevReqClearedFloor:
				requestsClearedFloor = f
				RCFcounter++
			default:
			}
		}
		message := StateMessage{localElevator, requestDistribution, requestsClearedFloor, localElevator.ID, counter, RDcounter, RCFcounter}
		if enable {
			stateMessageTx <- message
		}
		time.Sleep(25 * time.Millisecond)
	}
}

/*
Receives states from other elevators.
*/
func ReceiveStates(localID int, 
	globalElevState chan<- elevtypes.Elevator,
	newRequestDistribution chan<- map[int][common.NumFloors][common.NumButtons]int,
	globalElevReqClearedFloor chan<- int,
	globalElevLost chan<- int,
	) {
	time.Sleep(25*time.Millisecond)
	stateMessageRx := make(chan StateMessage, 1)
	go bcast.Receiver(16569, stateMessageRx)
	prevs := make(map[int]StateMessage)
	lastSeen := make(map[int]time.Time)
	timeout := 500*time.Millisecond
	for {
		select {
		case message := <- stateMessageRx:
			if message.ID == localID || message.ID == 0 {
				break
			}
			if _, inPrevs := prevs[message.ID]; !inPrevs {
				globalElevState <- message.Elevator
				if message.RDIter > 0 {
					newRequestDistribution <- message.RequestDistribution
				}
			} else {
				if prevs[message.ID].Iter > message.Iter {
					break
				}
				if prevs[message.ID].RDIter < message.RDIter {
					newRequestDistribution <- message.RequestDistribution
				}
				if prevs[message.ID].RCFIter < message.RCFIter {
					globalElevReqClearedFloor <- message.RequestsClearedFloor
				}
				if !elevtypes.ElevatorsAreEqual(prevs[message.ID].Elevator, message.Elevator){
					globalElevState <- message.Elevator
				}
			}
			lastSeen[message.ID] = time.Now()
			prevs[message.ID] = message
		default:	
		}
		for id, lastTime := range lastSeen {
			if time.Now().Sub(lastTime) > timeout {
				globalElevLost <- id
				delete(lastSeen, id)
				delete(prevs, id)
			}
		}
		time.Sleep(5*time.Millisecond)
	}
}
