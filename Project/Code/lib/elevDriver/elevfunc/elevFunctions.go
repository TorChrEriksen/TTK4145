package elevfunc

import (
    //"fmt"
    "time"
    "../lib/"
)

func GoToFloor(requestedFloor int){
//func GoToFloor(requestedFloor int, ch chan int)
	currentFloor := int(driverInterface.GetFloorSignal())
	
	if currentFloor < requestedFloor{
		driverInterface.SetSpeed(300)
	}else if currentFloor > requestedFloor{
		driverInterface.SetSpeed(-300)
	}
	
	//go func()
	for{
		time.Sleep(time.Millisecond * 10)
		currentFloor := int(driverInterface.GetFloorSignal())
		if currentFloor == requestedFloor{
			driverInterface.StopElevator()
			break
			//ch <- 0
		}
	}
}

/*
func dummy() {
	ch := make(chan int)
	
	GoToFloor(ch)
	
	go func() {
		for {
			select {
			case result <- ch :
				break;
				// Handle elevator arrival
				
			case default :
				continue
			}
		}	
	}
	
	// Elevator arrived, do something.
}*/
