package elevfunc

import (
    //"fmt"
    "../lib/"
    "../elevquecontrol/"
)

//func Create(intBtChan chan int, floorChan chan int, stopChan chan int, extBtChan chan int) {
func CreateAndListen(){
	
	intButtonChannel := make(chan int)
    extButtonChannel := make(chan int)
    floorChannel := make(chan int)
    stopChannel := make(chan int)
    queintButtonChannel := make(chan int)
    queextButtonChannel := make(chan int)
    quefloorChannel := make(chan int)
    questopChannel := make(chan int)
    elevquecontrol.Create(queintButtonChannel, quefloorChannel, questopChannel, queextButtonChannel)
	driverInterface.Create(intButtonChannel, floorChannel, stopChannel, extButtonChannel)
	for {
		select {
			case intButtonSignal := <- intButtonChannel :
				// release thread (use a channel or fire a go routine)
				go func() {
					//fmt.Println(intButtonSignal)
					queintButtonChannel <- intButtonSignal
					//time.Sleep(time.Second * 2) // Just for demonstration why we want to release the thread!
				}()
            case floorSignal := <- floorChannel :
                go func() {
					//fmt.Println(floorSignal)
					quefloorChannel <- floorSignal
                }()
			case stopSignal := <- stopChannel :
			   go func(){
			        //fmt.Println(stopSignal)
			        questopChannel <- stopSignal
			   }()
            case extButtonSignal := <- extButtonChannel :
                go func() {
                    //fmt.Println(extButtonSignal)
                    queextButtonChannel <- extButtonSignal
                }()
		}
	}
}
/*
func GoUp(){
	driverInterface.SetSpeed(C.int(300))
}

func GoDown(){
	driverInterface.SetSpeed(C.int(-300))
}

func Wait(){
	
}
func Stop(){
	driverInterface.StopElevator()
}
*/
/*
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
*/
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
