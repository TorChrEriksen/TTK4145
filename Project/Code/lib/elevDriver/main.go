package main

import (
	"./lib"
	"fmt"
	"math"
	"time"
)

type ByFloor []Order

type order struct{
	floor int
	dir int
	clear bool
}

type exOrder struct{
	floor		int
	dir 		int
	recipient	int
	origin		int
	cost		float64
}

type exLights struct{
	floor 	int
	dir		int
	value	int
	}

const INT = 2 int


func (p Order) String() string {
	return fmt.Sprintf("%d: %s, Del: %s", p.floor, p.dir, p.clear)
}

func contains(a Order, list []Order) bool {
	for _, i := range list {
		if i == a {
			return true
		}
	}
	return false
}


func (a ByFloor) Len() int           { return len(a) }
func (a ByFloor) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByFloor) Less(i, j int) bool { return a[i].floor < a[j].floor }

func state(direction string) {
	switch {
	case direction == "DOWN":
		driverInterface.SetSpeed(-300)
	case direction == "UP":
		driverInterface.SetSpeed(300)
	case direction == "STOP":
		driverInterface.SetSpeed(300)
		driverInterface.SetSpeed(-300)
		driverInterface.SetSpeed(0)
	case direction == "OPEN":
		driverInterface.SetDoorLamp(1)
		time.Sleep(2 * time.Second)
		driverInterface.SetDoorLamp(0)
	}
}

func cost(orderList []order, afterOrderList []order, currPos int, dir_now string, new_order int, new_order_dir string) float64 {
	var squared float64
	squared = 2.0
	switch {
	case dir_now == "UP" && new_order_dir == "UP":
		if currPos < new_order {
			return math.Pow(float64((new_order-currPos)+len(orderList)), squared)
		} else {
			return math.Pow(float64(((2*orderList[len(orderList)-1])-new_order-currPos)+len(orderList)), squared)
		}

	case dir_now == "DOWN" && new_order_dir == "DOWN":
		if new_order < currPos {
			return math.Pow(float64((currPos-new_order)+len(orderList)), squared)
		} else {
			return math.Pow(float64(new_order+currPos-orderList[0]+len(orderList)), squared)
		}

	case dir_now == "UP" && new_order_dir == "DOWN":
		return math.Pow(float64(2*orderList[len(orderList)-1]-currPos-new_order+len(orderList)), squared)

	case dir_now == "DOWN" && new_order_dir == "UP":
		return math.Pow(float64(currPos+new_order-orderList[0]+len(orderList)), squared)
	}
	return -1.0
}


func main(){
	driverInterface.Init()
	intButtonChannel 	:= make(chan int)
	extButtonChannel 	:= make(chan int)
	floorChannel 		:= make(chan int)
	stopChannel 		:= make(chan int)
	timeoutChannel 		:= make(chan int)
	
	driverInterface.Create(intButtonChannel, floorChannel, stopChannel, extButtonChannel, timeoutChannel)

	ordersChann 		:= make(chan order)
	updateCurrentOrder 	:= make(chan bool)
	updatePos 			:= make(chan order)
	
	costRequestIn 	:= make(chan exOrder)
	costRequestOut 	:= make(chan exOrder)
	costResponsIn 	:= make(chan exOrder)
	costResponsOut 	:= make(chan exOrder)	
	
	delegate := make(chan exOrder)
	
	setOtherLights := make(chan exLights) 

	var currentFloor 	int
	var lastFloor 		int
	var orderList 		[]order
	var afterOrders 	[]order



	//SIGNAL HANDLING
	go func(){
		for{
			select{
				case buttonSignal := <- intButtonChannel:
					go func(){
						incommingI := order{floor: buttonSignal-10, dir: INT}
						ordersChann <- incommingI
					}()

				case new_order := <- ordersChann:
//				fmt.Println(new_order)
					go func (){
						if new_order==order{-1, "NO"}{
							orderList, afterOrders = afterOrders, nil
						}
						if !contains(new_order, orderList) && new_order.dir != "NO" {
							if direction == "UP" {
								if (new_order < currentFloor && currentFloor==lastFloor) || (new_order < lastFloor+1&&currentFloor==0) {
									afterOrders = append(afterOrders, new_order)
									sort.Sort(ByFloor(afterOrders))
								} else {
									orderList = append(orderList, new_order)
									sort.Sort(ByFloor(orderList))
								}
							} else if direction == "DOWN" {
								if (new_order > currentFloor) || (new_order > lastFloor-1) {
									afterOrders = append(afterOrders, new_order)
									sort.Sort(ByFloor(afterOrders))
								} else {
									orderList = append(orderList, new_order)
									sort.Sort(ByFloor(orderList))
								}
							} else {
								orderList = append(orderList, new_order)
								sort.Sort(ByFloor(orderList))
							}
						}else if new_order.clear{
							if new_order.floor == orderList[0]{
								orderList = orderList[1:]
							}else if new_order.floor == orderList[len(orderlist)-1]{
								orderList = orderList[:len(orderList)-1]
							}
						}
					}()

				case a:=<-updateCurrentOrder:
					go func(){
						a=a
						if len(orderList)==0&&len(afterOrders)==0{
							updatePos <- order{-1, "NO"}
						}else if len(orderList)==0{
							ordersChann <- order{-1, "NO"}

						}else if status == "IDLE" || status=="UP"{
							updatePos <- orderList[0]
						}else if status=="DOWN"{
							updatePos <- orderList[len(orderList)-1]
						}
					}()

				case new_stuff := <- updatePos:
					go func(){

						if new_stuff.dir=="NO"{
							for currentFloor==0{
								time.Sleep(time.Millisecond*250)
							}
							status = "IDLE"
							state("STOP")
						}else{
							currentOrder = new_stuff
						}
						if currentOrder.floor<lastFloor{
							state("UP")
							status=="UP"
						}else if currentOrder.floor>lastFloor{
							state("DOWN")
							status=="DOWN"
						}else if currentOrder.floor==currentFloor{
							state("OPEN")
							ordersChann <- order{currentOrder.floor, currentOrder.dir, true}
							updateCurrentOrder <- true
						}
					}()

				case extSig := <- extButtonChannel:
					go func(){
						setOtherLights <- exLights{((extSignal - (extSignal % 2) - 30) / 10), (extSignal % 2), 1}
						if extSignal%2==0{
							extSig = order{((extSignal - (extSignal % 2) - 30) / 10),"DOWN", false}
						}else{
							extSig = order{((extSignal - (extSignal % 2) - 30) / 10),"UP", false}
						}
//						orderChannel <- incommingE				
						costRequestOut <- extOrder{floor: extSig.floor, direction: extSig.dir, origin: THIS_ID}
						min := cost(orderList, afterOrders, lastFloor, status, extSig.floor, extSig.dir)
						//wait for cost responses for 2 sek, whilst updating min
						delegate <-
					}()
				case  costReq := <-costRequestIn:
					go func(){
							costResponsOut <- extOrder{floor: costReq.floor, dir: costReq.dir, recipient: THIS_ID, origin: costReq.origin, cost: cost(orderList, afterOrders, lastFloor, status, costReq.floor, costReq.dir)}
						}()

				case floor := <-floorChannel:
					go func() {
						floor = floor - 30 //REMEMBER TO ADJUST FOR N_FLOORS
						currentFloor = floor
						if currentFloor != 0{
							lastFloor = currentFloor
						}
						driverInterface.SetFloorSignal(lastFloor)
						updatePos <- currentOrder
					}()

			}
		}
	}()
}
