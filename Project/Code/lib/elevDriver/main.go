// Elevator control logic
/*Elvator functions:
 
*/
package main

import(
    "./lib"
    //"fmt"
    "./elevfunc"
    //"./elevquecontrol"
    //"time"
    //"os"
)

/*Driver interface Signals from controller:
Stop button - 10
Internal buttons - 11 - ...
External Buttons - 50 - ...
Floor sensors output - 30 - ...
*/

func main(){
   driverInterface.Init()
   driverInterface.StopElevator()
   elevfunc.CreateAndListen()
   //elevquecontrol.Create()
   
   
	
	//for{
	//}
   //driverInterface.GetFloorSignal()
   //driverInterface.Test_Run()
	// Create driver with all the channels we need
	//elevfunc.GoToFloor(0)
	/*
	intButtonChannel := make(chan int)
    extButtonChannel := make(chan int)
    floorChannel := make(chan int)
    stopChannel := make(chan int)
	driverInterface.Create(intButtonChannel, floorChannel, stopChannel, extButtonChannel)
	*/
	
	//driverInterface.SetSpeed(-300)
	//elevfunc.GoToFloor()
	//Test of all functions
	//driverInterface.SetFloorLamp(3)   //works for values <0,3>
	//driverInterface.SetStopLamp(1)    //OK!
	//driverInterface.SetDoorLamp(1)    //OK!
	//driverInterface.SetButtonLamp("DOWN",2,1) //OK! (Panic mode also implemented)
/*
	for {
		select {
			case intButtonSignal := <- intButtonChannel :
				// release thread (use a channel or fire a go routine)
				go func() {
					fmt.Println(intButtonSignal)
					//time.Sleep(time.Second * 2) // Just for demonstration why we want to release the thread!
				}()
            case floorSignal := <- floorChannel :
                go func() {
					fmt.Println(floorSignal)
                }()
			case stopSignal := <- stopChannel :
			   go func(){
			        fmt.Println(stopSignal)
			   }()
            case extButtonSignal := <- extButtonChannel :
                go func() {
                    fmt.Println(extButtonSignal)
                }()
		}
	}*/
}
