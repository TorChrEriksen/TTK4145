// Elevator control logic
/*Elvator functions:
 
*/
package main

import(
    "./lib"
    "fmt"
    //"time"
    //"os"
)

/*Driver interface Signals from controller:
STOP = 10
INT_BTN_1 = 11
INT_BTN_2 = 12
INT_BTN_3 = 13
INT_BTN_4 = 14
EXT_BTN_1_UP = 21
EXT_BTN_2_UP = 22
EXT_BTN_2_DOWN = 23
EXT_BTN_3_UP = 24
EXT_BTN_3_DOWN = 25
EXT_BTN_4_DOWN = 26
FLOOR_NO = 30
FLOOR_1 = 31
FLOOR_2 = 32
FLOOR_3 = 33
FLOOR_4 = 34
*/

func main(){
   driverInterface.Init()
   //driverInterface.GetFloorSignal()
   //driverInterface.Test_Run()
	// Create driver with all the channels we need
	intButtonChannel := make(chan int)
    extButtonChannel := make(chan int)
    floorChannel := make(chan int)
    stopChannel := make(chan int)
//intBtChan chan int, floorChan chan int, stopChan chan int, extBtChan chan int
	driverInterface.Create(intButtonChannel, floorChannel, stopChannel, extButtonChannel)
	
	//driverInterface.SetSpeed(-300)
	
	//Test of all functions
	//driverInterface.SetFloorLamp(3)   //works for values <0,3>
	//driverInterface.SetStopLamp(1)    //does not seem to be working on arbeidsplass 16
	//driverInterface.SetDoorLamp(1)    //OK!
	//driverInterface.SetButtonLamp("DOWN",2,1) //OK! (Panic mode also implemented)

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
	}
}
