package elevquecontrol

import (
    "fmt"
    "time"
    "../lib/"
    //"../elevfunc/"
    "math"
)

func Create(intBtChan chan int, floorChan chan int, stopChan chan int, extBtChan chan int) {
	floorReqChan := make(chan int)
	currentFloorChan := make(chan int)
	go OrderGrab (intBtChan, floorChan, stopChan, extBtChan, floorReqChan, currentFloorChan)
	go QueMaker(floorReqChan, currentFloorChan)
}

func OrderGrab(intBtChan chan int, floorChan chan int, stopChan chan int, extBtChan chan int, floorReqChan chan int, currentFloorChan chan int){
	for {
		select {
			case intButtonSignal := <- intBtChan :
				// release thread (use a channel or fire a go routine)
				go func() {
					fmt.Println(intButtonSignal)
					IntBtnResolver(intButtonSignal, floorReqChan)
					//time.Sleep(time.Second * 2) // Just for demonstration why we want to release the thread!
				}()
            case floorSignal := <- floorChan :
                go func() {
					fmt.Println(floorSignal)
					var floor int = int(math.Mod(float64(floorSignal),float64(driverInterface.FLOOR_NO_BASE-1)))
					if floor == 0{
						//currentFloorChan <- -1
					}else{
						currentFloorChan <- floor-1 //making sure lowest floor is 0
					}
                }()
			case stopSignal := <- stopChan :
			   go func(){
			        fmt.Println(stopSignal)
			   }()
            case extButtonSignal := <- extBtChan :
                go func() {
                    fmt.Println(extButtonSignal)
                }()
		}
	}
}

func IntBtnResolver(signal int, floorReqChan chan int){
	//var convSignal float64= float64(driverInterface.FLOOR_NO_BASE-1)
	var floor int= int(math.Mod(float64(signal),float64(driverInterface.INT_BTN_BASE-1)))-1
	fmt.Println(floor)
	floorReqChan <- floor
}

func QueMaker(intFloorReqChan chan int, currentFloorChan chan int){
	var intQueArray [driverInterface.FLOORS]int
	for i := range intQueArray{
		intQueArray[i] = 0
	}
	currentFloor := -1
	intFloorRequest := currentFloor
	stateChanged := 0
	dir := 0
	//int maxFloor := nil
	for{
		time.Sleep(time.Millisecond * 10)
		select{
			case tempFloor := <- currentFloorChan:
				go func(){
					currentFloor = tempFloor
					fmt.Println("Current floor is: ", currentFloor)
					stateChanged = 0
				}()
			case tempFloorRequest := <- intFloorReqChan:
				go func(){
					fmt.Println("Hopefully your request will be processed")
					intFloorRequest = tempFloorRequest
					dir= intFloorRequest - currentFloor
					fmt.Println("Dir is: ", dir)
					intQueArray[intFloorRequest] = 1
					stateChanged = 0
				}()
			default:
				break
				//fmt.Println(currentFloor)
		}
		//TODO protect stuff from overshooting the sensor
		//Most of this this stuff is to be moved to driverInterface!!
		if dir < 0{
			elevfunc.GoDown()
			//driverInterface.SetSpeed(-100)
		}else if dir >0{
			//elevfunc.GoUp()
			driverInterface.SetSpeed(100)
		}else if dir == 0{
			driverInterface.StopElevator()
		}
		if currentFloor == intFloorRequest && stateChanged == 0{
			time.Sleep(time.Millisecond * 10)
			fmt.Println("We have a match!")
			driverInterface.StopElevator()
			dir = 0
			stateChanged = 1
		}
		//fmt.Println(currentFloor)
		
	}
	//var currentFloor int
	//currentFloor = -1
	/*
	for{
		go func(){
			intFloorRequest := <- intFloorReqChan
			intQueArray[intFloorRequest] = 1
			fmt.Println("Que array: ", intQueArray)
		}()
		go func(){
			currentFloor := <- currentFloorChan
			fmt.Println("Current floor:", currentFloor)
		}()
		//fmt.Println("Que array: ", intQueArray)
	}*/
}

