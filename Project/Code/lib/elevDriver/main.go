// Elevator control logic

package main

import (
	"./lib"
	"fmt"
	"sort"
	"time"
	"math"
	//"os"
)
const THIS_ID = 1

type gatherCost struct {
	floor int
	direction string
	origin int
	recipient int
}
//works
func state(direction string){
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

//TOMORROW FIRST THING: MERGE RUN AND MANAGE, ADD STOP AND CURRENT FLOOR SIGNALS, SETUP INFO CHANN STRUCT.
func run(currentOrder int, currentFloor int, dir string){
	switch {
	case currentOrder < currentFloor:
		for currentOrder != currentFloor{
			state("DOWN")
			dir = "down"
		}
		state("UP")

	case currentOrder > currentFloor:
		for currentOrder != currentFloor{
			state("UP")
			dir = "up"
		}
		state("DOWN")

	case currentOrder == currentFloor:
		state("STOP")
		state("OPEN")
		currentOrder = 0
	}
}

func manageOrders(ordersIn chan int, orderList []int, afterOrders []int, currentOrder int, currentFloor int, direction string){
	switch{
	case (currentOrder == 0) && (len(orderList)!=0):
		go func(){
			if direction == "up"{
				orderList = orderList[1:]
				currentOrder = orderList[0]
			}else{
				orderList = orderList[:len(orderList)-1]
				currentOrder = orderList[len(orderList)-1]
			}
		}()
	case len(orderList)==0 && len(afterOrders)!=0:
		orderList, afterOrders = afterOrders, nil
	case len(orderList)==0 && len(afterOrders)==0:
		state("IDLE")
	}
	
	select{
	case new_order := <-ordersIn:
		go func(){
			if direction == "up"{
				if new_order < currentFloor{
					afterOrders = append(afterOrders, new_order)
					sort.Ints(afterOrders)
				}else{
					orderList = append(orderList, new_order)
					sort.Ints(orderList)
				}
			}else{ //if dir == down
				if new_order> currentFloor{
					afterOrders = append(afterOrders, new_order)
					sort.Ints(afterOrders)
				}else{
					orderList = append(orderList, new_order)
					sort.Ints(orderList)
				}
			}
		}()
	}

}


//this method works, send orders through chann's, TOMORROW: extHandling
func getSignals(intButtonChannel chan int, xBSignal chan int, orderChannel chan int){
   	select {
        case buttonSignal := <-intButtonChannel :
        go func() {
            buttonSignal = buttonSignal-10
            fmt.Println(buttonSignal)
            orderChannel <- buttonSignal
        }()
        case extOrder := <- xBSignal:
            go func(){
                extOrder = ((extOrder-(extOrder%2)-30)/10)
                fmt.Println(extOrder)
                orderChannel <- extOrder
            }()
       	    
    }
}
   	
   	
   	
   	/*
   	select {
       	case floor := <- floorChan:
       		go func (){
       			fmt.Println("Current floor: ", currentFloor)
       			currentFloor = floor - 30
       		}()
       
   	    case stop := <- stopChan:
   	    	go func(){
       		    if stop != 0{
           			fmt.Println("STOPPING")
           			state("STOP")
       	    		direction = "STOP"
       	    		
       	    		}
       		}()
       	case iOrder := <-intButtonChannel:
       		go func(){
       			fmt.Println("Internal order: ", iOrder - 10)
       			orderInChan <- iOrder-10
       			
       		}()
       	case exOrder := <-extButtonChannel:
       		go func(){
       			a := gatherCost{floor:((exOrder-(exOrder%2)-30)/10), origin:CURRENT_ID, recipient:3}
       			if exOrder%2==0{
       				a.direction="down"
       			}else{
       				a.direction="up"
       			}
       			orderInChan <- a.floor
       		}()
       	case floorTimeout := <- timeoutChannel:
       		go func(){
       			fmt.Println("From the timeout channel: ", floorTimeout)
      		}()
	}

}
*/
//add afterOrderList to eq.
func cost(orderList []int, afterOrderList int, currPos int, dir_now int, new_order int, new_order_dir int) float64{
    var squared float64
    squared = 2.0
	switch {
		case dir_now==1 && new_order_dir==1:
			if currPos<new_order{
				return math.Pow(float64((new_order-currPos)+len(orderList)), squared)
			}else{
				return math.Pow(float64(((2*orderList[len(orderList)-1])-new_order-currPos)+len(orderList)),squared)
			}

		case dir_now==0 && new_order_dir==0:
			if new_order<currPos{
				return math.Pow(float64((currPos-new_order)+len(orderList)),squared)
			}else{
				return math.Pow(float64(new_order+currPos-orderList[0]+len(orderList)),squared)
			}

		case dir_now==1&&new_order_dir==0:
			return math.Pow(float64(2*orderList[len(orderList)-1]-currPos-new_order+len(orderList)),squared)

		case dir_now==0&&new_order_dir==1:
			return math.Pow(float64(currPos+new_order-orderList[0]+len(orderList)),squared)
	}
	return 0.0
}





func main() {
	driverInterface.Init()

	intButtonChannel 	:= make(chan int)
	extButtonChannel 	:= make(chan int)
	floorChannel 		:= make(chan int)
	stopChannel 		:= make(chan int)
	timeoutChannel 		:= make(chan int)

	driverInterface.Create(intButtonChannel, floorChannel, stopChannel, extButtonChannel, timeoutChannel)
    
	//var orderList 		[]int
	//var afterOrders 	[]int
	//var status 			string
	var currentFloor 	int
	//var lastFloor       int
    //var direction       string
    
	//reapCost 	:= make(chan gatherCost)
	//replyCost	:= make(chan gatherCost)
    /*
	pollCost	:= make(chan struct)
	recieveCost	:= make(chan struct)
    */
	ordersChann := make(chan int)
	
    fmt.Println("testing, testing")
    //lastFloor = 0
    for {
        time.Sleep(10*time.Millisecond)
        go getSignals(intButtonChannel, extButtonChannel, ordersChann)
        
        go func(){
            select{
                case newst := <- ordersChann:
                    fmt.Println("Got order to: ", newst)
                case next_floor:= <- floorChannel:
                    currentFloor = next_floor - 30
                    fmt.Println(currentFloor)
                case stop := <-stopChannel:
                    fmt.Println("STOP")
            }
        }()
    }
}
