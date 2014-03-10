// Elevator control logic
/*Elvator functions:
 
*/
package main

import(
    "./lib"
    "fmt"
    "time"
    "sort"
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
//Sets speed and direction. Also stops at signal. TODO: const of 1&&-1 as up and down
func drive(direction string){
	switch	{
	case direction == "DOWN":
		driverInterface.SetSpeed(-300)
	case direction == "UP":
		driverInterface.SetSpeed(300)
	case direction == "STOP":
		driverInterface.SetSpeed(0)
	case direction == "OPEN":
		driverInterface.SetDoorLamp(1)
		time.Sleep(5 * time.Second)
		driverInterface.SetDoorLamp(0)
	}
}

/* One major problem: If the elevator goes to an order that is due in the opposite direction, 
and pick up orders that want to go in opposite direction, beyond the original order floor, it [the elevator] will do so.

For example: Elevator at ground floor[1]. Order to go down come from floor 3. 
Person on 2nd floor wants to go to 4th. The elevator will now go to 2nd, talking an order for 4th, then go to 3rd, taking order to ground, then to 4th, then to ground.

Although this can be fixed with ordersorting
*/
//Må handle med global orderliste, hva skjer med afterOrders. tenk.
func runOffList(currentFloor int, orders []int){
	var maks int
	var next int
	var dir string 			//"DOWN" || "UP"
	var afterOrders []int 	// For orders that does not conform (internals in opposite direction)

	for len(orders)==0{
		driverInterface.SetSpeed(0)
		time.Sleep(100 * time.Millisecond)
	}
	
		//get highest order
	maks = orders[len(orders)-1]
	
		//Get the direction of order and go there
	if maks < currentFloor {
		for len(orders)>0{
			dir = "DOWN"
			go func(){
				maks = orders[len(orders)-1]
				if maks>currentFloor{
					afterOrders, orders = append(afterOrders, maks), orders[:len(orders)-1]
				}else{
					next = orders[len(orders)-1]
				}
			}()

			go func(){
				drive(dir)
				for currentFloor!=next{
					time.Sleep(10 * time.Millisecond)
				}
				drive("STOP")
				orders = orders[:len(orders)-1]
				drive("OPEN")
			}()
		}orders, afterOrders = afterOrders, nil


	}else{
		for len(orders)>0{
			dir = "UP"
			go func(){
				minn = orders[0]
				if minn<currentFloor{
					afterOrders, orders = append(afterOrders, minn), orders[0:]
				}else{
					next = orders[0]
				}
			}()

			go func(){
				drive(dir)
				for currentFloor!=next{
				time.Sleep(10 * time.Millisecond)
				}
				drive("STOP")
				orders = orders[0:]
				drive("OPEN")
			}()
		}
		orders, afterOrders = afterOrders, nil
	}
}


func main(){
   driverInterface.Init()
   //driverInterface.GetFloorSignal()
   //driverInterface.Test_Run()
	// Create driver with all the channels we need
	intButtonChannel := make(chan int)
    extButtonChannel := make(chan int)
    floorChannel := make(chan int)
    stopChannel := make(chan int)
    timeoutChannel := make(chan int)
	driverInterface.Create(intButtonChannel, floorChannel, stopChannel, extButtonChannel, timeoutChannel)
	

	//VARIABLES
	//last known floor
	var last_floor int
	//orderlist
	var orders []int
	var speed int
	var dir string //DOWN || UP


	//driverInterface.SetSpeed(-300)
	
	//Test of all functions
	//driverInterface.SetFloorLamp(3)   //works for values <0,3>
	//driverInterface.SetStopLamp(1)    //OK!
	//driverInterface.SetDoorLamp(1)    //OK!
	//driverInterface.SetButtonLamp("DOWN",2,1) //OK! (Panic mode also implemented)

	//runOffList(last_Floor, orders)

	for {
		select {
			case intButtonSignal := <- intButtonChannel :
				// release thread (use a channel or fire a go routine)
				go func() {
					orders = append(orders, intButtonSignal)
					//can use Sort.IsSorted to check if need sorting
					Sort.Ints(orders)
					fmt.Println(orders)
					//time.Sleep(time.Second * 2) // Just for demonstration why we want to release the thread!
				}()
            case floorSignal := <- floorChannel :
                go func() {
					last_floor = floorSignal - 30
					fmt.Println(a)
					
                	}()
			case stopSignal := <- stopChannel :
			   go func(){
			        fmt.Println(stopSignal)
			        driverInterface.SetSpeed(0)
			        orders=[]	//clear orders
			        time.Sleep(time.Second * 3) //wait for a bit

			   }()
            case extButtonSignal := <- extButtonChannel :
                go func() {
                	//TODO: Change the value from pressing the ext. buttons
                	//to make simpler to sort the signals, a%2==0:UP
			        if extButtonSignal%2==0{

			        	}else{

			        	}
			        /*If in same direction as other orders:
			        	if on a floor in orders: forget
			        	else: get cost from others, send orders to lowest
			        */
                    fmt.Println(extButtonSignal)
                }()
            case floorTimeout := <- timeoutChannel :
                go func() {
                    fmt.Println("From timeout channel: ", floorTimeout)
                }()
		}
	}
}
