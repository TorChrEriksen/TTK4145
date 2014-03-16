package driverRunning

import (
	"fmt"
	"sort"
	"math"
	"./"
)


func state(direction string){
	select	{
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

func run(currentOrder int, currentFloor int, dir string){
	select {
	case currentOrder < currentFloor:
		for currentOrder != currentFloor{
			state("DOWN")
			dir = "down"
		}
		STATE("UP")

	case currentOrder > currentFloor:
		for currentOrder != currentFloor{
			state("UP")
			dir = "up"
		}
		state("DOWN")

	case currentOrder == currentFloor:
		state("STOP")
		state("OPEN")
		currentOrder = nil
	}
}

func manageOrders(ordersIn chan int, orderList []int, afterOrders []int, currentOrder int, currentFloor int, direction string){
	select{
	case currentOrder == nil && len(orderList)=!0:
		go func(){
			if direction == "up"{
				orderList = orderList[1:]
				currentOrders = orderList[0]
			}else{
				orderList = orderList[:len(orderList)-1]
				currentOrders = orderList[len(orderList)-1]
			}
		}
	case new_order := <-ordersIn:
		go func(){
			if direction == "up"{
				if new_order < currentFloor{
					afterOrders = append(afterOrders, new_order)
					sort(afterOrders)
				}else{
					orderList = append(orderList, new_order)
					sort(orderList)
				}
			}else{ //if dir == down
				if new_order> currentFloor{
					afterOrders = append(afterOrders, new_order)
					sort(afterOrders)
				}else{
					orderList = append(orderList, new_order)
					sort(orderList)
				}
			}
		}
	case len(orderList)==0 && len(afterOrders)!=0:
		orderList, afterOrders = afterOrders, nil
	case len(orderList==0)&&len(afterOrders)==0:
		state("IDLE")
	}

}

func getSignals(intButtonChannel chan int, extButtonChannel chan int, orderInChan chan int, timeoutChannel chan int, floorChan chan int, stopChan chan int, direction string, CURRENT_ID int, costIN chan gatherCost, costOUT chan gatherCost){
	type gatherCost struct {
		floor int
		direction string
		origin int
		recipient int
	}

	select {
	case floor := <- floorChan:
		go func (){
			currentFloor = floor - 30
		}

	case stop := <- stopChan:
		go func(){
			state("STOP")
			direction = "STOP"
		}
	case iOrder := <-intButtonChannel:
		go func(){
			orderInChan <- iOrder-10
		}
	case exOrder := <-extButtonChannel:
		go func(){
			a := gatherCost{floor:((exOrder-(exOrder%2)-30)/10), origin:CURRENT_ID, recipient:3}
			if exOrder%2==0{
				a.direction="down"
			}else{
				a.direction="up"
			}
			orderInChan <- a.floor
		}
	case floorTimeout := <- timeoutChannel:
		go func(){
			fmt.Println("From the timeout channel: ", floorTimeout)
		}
	}
}

//add afterOrderList to eq.
func cost(orderList []int, afterOrderList int, currPos int, dir_now int, new_order int, new_order_dir int) int{
	select {
		case dir_now==1 && new_order_dir==1:
			if currPos<new_order{
				return math.Sqrt( (new_order-currPos)+len(orderList) )
			}else{
				return math.Sqrt( ((2*orderList[len(orderList)-1])-new_order-currPos)+len(orderList))
			}

		case dir_now==0 && new_order_dir==0:
			if new_order<currPos{
				return math.Sqrt((currPos-new_order)+len(orderList))
			}else{
				return math.Sqrt(new_order+currPos-orderList[0]+len(orderList))
			}

		case dir_now==1&&new_order_dir==0:
			return math.Sqrt(2*orderList[len(orderList)-1]-currPos-new_order+len(orderList))

		case dir_now==0&&new_order_dir==1:
			return math.Sqrt(currPos+new_order-orderList[0]+len(orderList))
	}
}
