package elevfsm

import (
	"Elevator/pkg/common"
	"Elevator/pkg/elevio"
	"Elevator/pkg/elevtypes"
	"Elevator/pkg/requests"
	"time"
)

/*
Finite state machine for a single elevator.
*/
func StateMachine(ID int, 
	enableTransmit chan<- bool,
	newRequests <-chan [common.NumFloors][common.NumButtons]int,  
	elevState chan<- elevtypes.Elevator,
	requestsClearedFloor chan<- int,
	) {
	elevator := elevtypes.Uninitialized(ID)
	doorTimerTimeout := time.Duration(elevator.Config.DoorOpenDuration_s) * time.Second
	doorTimer := time.NewTimer(doorTimerTimeout)
	doorTimer.Stop()
	motorStopTimeout := 3*time.Second
	motorStopTimer := time.NewTimer(motorStopTimeout)
	motorStopTimer.Stop()
	obstructionTimeout := 10*time.Second
	obstructionTimer := time.NewTimer(obstructionTimeout)
	obstructionTimer.Stop()
	
	if elevio.GetInitialFloor() == -1 {
		elevator.Dirn = elevtypes.MD_Down
		elevio.SetMotorDirection(elevator.Dirn)
		elevator.State = elevtypes.ES_Moving
		motorStopTimer.Reset(motorStopTimeout)
	} 
	floorArrival := make(chan int, 1)
	newObstrSignal := make(chan bool, 1)
	go elevio.PollFloorSensor(floorArrival)
	go elevio.PollObstructionSwitch(newObstrSignal)

	for {
		select {
		// Force prioritizes detecting a floor arrival for safety reasons.
		case f := <- floorArrival:
			if elevator.Stuck {
				elevator.Stuck = false
				enableTransmit <- true
			}
			elevator.Floor = f
			elevio.SetFloorIndicator(elevator.Floor)
			motorStopTimer.Stop()
			switch elevator.State {
			case elevtypes.ES_Moving:
				if requests.ShouldStop(elevator) {
					elevio.SetMotorDirection(elevtypes.MD_Stop)
					elevio.SetDoorOpenLamp(true)
					elevator = requests.ClearAtCurrentFloor(elevator)
					requestsClearedFloor <- elevator.Floor
					doorTimer.Reset(doorTimerTimeout)
					elevio.SetAllLights(elevator)
					elevator.State = elevtypes.ES_DoorOpen
				} else {
					motorStopTimer.Reset(motorStopTimeout)
				}
			}
		default: 
			select {
			case f := <- floorArrival:
				if elevator.Stuck {
					elevator.Stuck = false
					enableTransmit <- true
				}
				elevator.Floor = f
				elevio.SetFloorIndicator(elevator.Floor)
				motorStopTimer.Stop()
				switch elevator.State {
				case elevtypes.ES_Moving:
					if requests.ShouldStop(elevator) {
						elevio.SetMotorDirection(elevtypes.MD_Stop)
						elevio.SetDoorOpenLamp(true)
						elevator = requests.ClearAtCurrentFloor(elevator)
						requestsClearedFloor <- elevator.Floor
						doorTimer.Reset(doorTimerTimeout)
						elevio.SetAllLights(elevator)
						elevator.State = elevtypes.ES_DoorOpen
					} else {
						motorStopTimer.Reset(motorStopTimeout)
					}
				}
			case r := <- newRequests:
				elevator.Requests = r
				switch elevator.State {
				case elevtypes.ES_DoorOpen:
					if requests.RequestsAtCurrentFloor(elevator){
						doorTimer.Reset(doorTimerTimeout)
						elevator = requests.ClearAtCurrentFloor(elevator)
						requestsClearedFloor <- elevator.Floor
					}
				case elevtypes.ES_Idle:
					if requests.RequestsAtCurrentFloor(elevator){
						elevio.SetDoorOpenLamp(true)
						doorTimer.Reset(doorTimerTimeout)
						elevator = requests.ClearAtCurrentFloor(elevator)
						requestsClearedFloor <- elevator.Floor
						elevator.State = elevtypes.ES_DoorOpen
					} else {
						elevator.Dirn = requests.ChooseDirection(elevator)
						elevio.SetMotorDirection(elevator.Dirn)
						if elevator.Dirn != elevtypes.MD_Stop {
							elevator.State = elevtypes.ES_Moving
							motorStopTimer.Reset(motorStopTimeout)
						}
					}
				}
				elevio.SetAllLights(elevator)
			case o := <- newObstrSignal:
				elevator.DoorObstructed = o
				switch elevator.State {
				case elevtypes.ES_DoorOpen:
					if elevator.DoorObstructed {
						obstructionTimer.Reset(obstructionTimeout)
					} else {
						doorTimer.Reset(doorTimerTimeout)
						obstructionTimer.Stop()
						if elevator.Stuck {
							elevator.Stuck = false
							enableTransmit <- true
						}
					}
				}
			case <-doorTimer.C:
				switch elevator.State {
				case elevtypes.ES_DoorOpen:
					if !elevator.DoorObstructed {
						elevator.Dirn = requests.ChooseDirection(elevator)
						elevio.SetDoorOpenLamp(false)
						elevio.SetMotorDirection(elevator.Dirn)
						if elevator.Dirn == elevtypes.MD_Stop {
							elevator.State = elevtypes.ES_Idle
						} else {
							elevator.State = elevtypes.ES_Moving
							motorStopTimer.Reset(motorStopTimeout)
						}
					}
				}
			case <-motorStopTimer.C:
				elevator.Stuck = true
				enableTransmit <- false
			case <-obstructionTimer.C:
				elevator.Stuck = true
				enableTransmit <- false
			}
		}
		elevState <- elevator
	}
}
