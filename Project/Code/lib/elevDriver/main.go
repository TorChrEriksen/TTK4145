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
EXT_BTN_1_UP = 41
EXT_BTN_2_UP = 51
EXT_BTN_2_DOWN = 50
EXT_BTN_3_UP = 61
EXT_BTN_3_DOWN = 60
EXT_BTN_4_DOWN = 70
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
//Må handle med global orderliste, hva skjer med afterOrders. tenk. // Added another for loop that should execute the false additions.

func contains(s []int, e int) bool {
    for _, a := range s { if a == e { return true } }
    return false
}

func runOffList(ordersIn chan int, currentFloor chan int){
	var maks int
	var next int
	var dir string 			//"DOWN" || "UP"
	var orders []int
	var afterOrders []int 	// For orders that does not conform (internals in opposite direction)
	var last_floor int

	go func (){
		ny := <- ordersIn
		orders = append(orders, ny - 10)
		sort.Ints(orders)
	}

	go func (){
		new_floor := <- currentFloor
		last_floor = new_floor - 30

	}

	go func (){
		for {
			if len(orders)!=0{

				maks = orders[len(orders)-1]

				if maks< last_floor{
					for len(orders)>0{
						dir = "DOWN"

						go func(){
							maks = orders[len(orders)-1]
							if maks>last_floor{
								afterOrders, orders = append(afterOrders, maks), orders[:len(orders)-1]
								sort.Ints(afterOrders)
							}else{
								next=orders[len(orders)-1]
							}
						}()
						go func(){
							drive(dir)
							for last_floor!=next{
								time.Sleep(10 * time.Millisecond)
							}
							drive("STOP")
							orders = orders[:len(orders)-1]
							drive("OPEN")
						}()
					}
				}else{
					for len(orders)>0{
						dir = "UP"
						go func(){
							minn = orders[0]
							if minn<last_floor{
								afterOrders, orders = append(afterOrders, minn), orders[1:]
								sort.Ints(afterOrders)
							}else{
								next = orders[0]
							}
						}()

						go func(){
							drive(dir)
							for last_floor!=next{
								time.Sleep(10 * time.Millisecond)
							}
							drive("STOP")
							orders = orders[1:]
							drive("OPEN")
						}()
					}
				}
				orders, afterOrders = afterOrders, nil
			}
		}
	}
}


//func(distance to order + number of orders on the way) think about adding +len(afterOrderList[sort.Search(afterOrderList, new_order)]) to some
func cost(orderList []int, afterOrderList int, currPos int, dir_now int, new_order int, new_order_dir int) int{
	select {
		case dir_now==1 && new_order_dir==1:
			if currPos<new_order{
				return math.Sqrt( (new_order-currPos)+len(orderList[:sort.Search(orderList, new_order)]))
			}else{
				return math.Sqrt( ((2*orderList[len(orderList)-1])-new_order-currPos)+len(orderList))
			}

		case dir_now==0 && new_order_dir==0:
			if new_order<currPos{
				return math.Sqrt((currPos-new_order)+len(orderList[sort.Search(orderList, new_order):]))
			}else{
				return math.Sqrt(new_order+currPos-orderList[0]+len(orderList))
			}

		case dir_now==1&&new_order_dir==0:
			return math.Sqrt(2*orderList[len(orderList)-1]-currPos-new_order+len(orderList))

		case dir_now==0&&new_order_dir==1:
			return math.Sqrt(currPos+new_order-orderList[0]+len(orderList))
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
	ordersIn := make(chan int)
	currentFloor := make(chan int)
	var speed int
	var dir string //DOWN || UP


	//driverInterface.SetSpeed(-300)
	
	//Test of all functions
	//driverInterface.SetFloorLamp(3)   //works for values <0,3>
	//driverInterface.SetStopLamp(1)    //OK!
	//driverInterface.SetDoorLamp(1)    //OK!
	//driverInterface.SetButtonLamp("DOWN",2,1) //OK! (Panic mode also implemented)

	//runOffList(ordersIn, currentFloor)

	for {
		select {
			case intButtonSignal := <- intButtonChannel :
				// release thread (use a channel or fire a go routine)
				go func() {
					ordersIn <- intButtonSignal
					//can use Sort.IsSorted to check if need sorting
					//time.Sleep(time.Second * 2) // Just for demonstration why we want to release the thread!
				}()
            case floorSignal := <- floorChannel :
                go func() {
                	currentFloor <- floorSignal					
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
                	//to make simpler to sort the signals, a%2==1:UP
                	
                	// lowest[cost , elevatorID]
                	var lowest [2]int

                	// incomming[floor , direction]
                	var incomming  := [2]int{((extButtonChannel-extButtonChannel%2)-30)/10, extButtonChannel%2}	

                	/*
					if 
                	*/

                	/*
					lowest = [thisElev.cost , thisElev.ID]
					get the cost from the others
						if Other.Cost < lowest[0]:
							lowest = [Other.cost, Other.ID]
					send incomming[0] to Other.ID's OrderCHANL
                	*/
                }()
            case floorTimeout := <- timeoutChannel :
                go func() {
                    fmt.Println("From timeout channel: ", floorTimeout)
                }()
		}
	}
}




	/*
	for len(orders)==0{
		driverInterface.SetSpeed(0)
		time.Sleep(100 * time.Millisecond)
	}
	
		//get highest order

	maks = orders[len(orders)-1]
	for len(orders){
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

				go func(){}

		}else{
			for len(orders)>0{
				dir = "UP"
				go func(){
					minn = orders[0]
					if minn<currentFloor{
						afterOrders, orders = append(afterOrders, minn), orders[1:]
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
					orders = orders[1:]
					drive("OPEN")
				}()
			}
		orders, afterOrders = afterOrders, nil
		}
	}
}
*/
