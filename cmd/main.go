package main

import (
	"Elevator/pkg/network"
	"Elevator/pkg/elevfsm"
	"Elevator/pkg/elevio"
	"Elevator/pkg/elevtypes"
	"Elevator/pkg/distributor"
	"Elevator/pkg/common"
	"flag"
)

func main() {
	var id int
	var port string
	flag.IntVar(&id, "id", 0, "id of this peer")
	flag.StringVar(&port, "port", "15657", "port for sim")
	flag.Parse()

	// Initialize channels for main modules
	newLocalRequests := make(chan [common.NumFloors][common.NumButtons]int, 1)
	localElevState := make(chan elevtypes.Elevator, 1)
	globalElevState := make(chan elevtypes.Elevator, 1)
	spamLocalElevState := make(chan elevtypes.Elevator, 1)
	distributedRequests := make(chan map[int][common.NumFloors][common.NumButtons]int, 1)
	newRequestDistribution := make(chan map[int][common.NumFloors][common.NumButtons]int, 1)
	localElevReqClearedFloor := make(chan int, 1)
	globalElevReqClearedFloor := make(chan int, 1)
	globalElevLost := make(chan int, 1)
	enableTransmit := make(chan bool, 1)

	// Initialize main modules and goroutines
	elevio.Init("localhost:" + port)
	network.Init(id, enableTransmit, spamLocalElevState, globalElevState, distributedRequests, newRequestDistribution, 
		localElevReqClearedFloor, globalElevReqClearedFloor, globalElevLost)
	go elevfsm.StateMachine(id, enableTransmit, newLocalRequests, localElevState, localElevReqClearedFloor)
	go distributor.StateMachine(id, localElevState, globalElevState, newLocalRequests, newRequestDistribution, 
		distributedRequests, spamLocalElevState, globalElevReqClearedFloor, globalElevLost)
	for {}
}
