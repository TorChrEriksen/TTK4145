// Elevator control logic
/*Elvator functions:
 
*/
package main

import(
    "./lib"
    //"fmt"
    //"time"
    //"os"
)

func main(){
   driverInterface.Init()
   //driverInterface.GetFloorSignal()
   driverInterface.Test_Run()
	// Create driver with all the channels we need
	//buttonChannel := make(chan int)
	//stopChannel := make(chan int)
	//driverInterface.Create(buttonChannel)
	//stopChannel := make(chan int)
	
	//driverInterface.SetSpeed(-300)
	
	//Test of all functions
	//driverInterface.SetFloorLamp(3)   //works for values <0,3>
	//driverInterface.SetStopLamp(1)    //does not seem to be working on arbeidsplass 16
	//driverInterface.SetDoorLamp(1)    //OK!
	//driverInterface.SetButtonLamp("DOWN",2,1) //OK! (Panic mode also implemented)

	/*
	for {
		select {
			case buttonSignal := <-buttonChannel :
				// release thread (use a channel or fire a go routine)
				go func() {
					fmt.Println(buttonSignal)
					time.Sleep(time.Second * 2) // Just for demonstration why we want to release the thread!
				}()
			case stopSignal := <-stopChannel :
			   go func(){
			      fmt.Println(stopSignal)
			   }()

			// more cases
//			case floorSignal := <-floorChannel :
		}
	}*/
}
