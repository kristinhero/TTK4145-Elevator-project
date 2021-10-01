Elevator Project
================


Modules
-------
### Assigner

Optimally assigns hall requests to n elevators, based on total time used to serve the requests. Uses an executable that is linked under "Requests".

### Common

Contains global constants.

### Distributor

The distributor holds the states of the active and non active elevators on the network.
It updates the states based on messages from the local elevator and the network.
It (re)distributes the current requests when there is a new button press, 
an elevator is lost or an elevator has come back.

### Elevfsm

Finite state machine for a single elevator. 

### Elevio

Driver to communicate with the elevator hardware/simulator. Link to git repo is under "Resources".

### Elevtypes

Elevator types and functions used on these types. 

### Network

Communication with other elevators over UDP using broadcast. Sends a single message type every 25th millisecond with the elevator state, including any new request distribution and any new cleared requests at a floor. Link to git repo for starter code is under "Resources". 

### Requests

Returns decisions for a single elevator based on current requests.

Resources
---------
- Hall request assigner: https://github.com/TTK4145/Project-resources/tree/master/cost_fns/hall_request_assigner
- Network: https://github.com/TTK4145/Network-go
- Driver: https://github.com/TTK4145/driver-go