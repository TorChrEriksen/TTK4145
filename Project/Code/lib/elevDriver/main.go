package main
/*
// #cgo CFLAGS: -std=c99 -g -Wall -O2 -I .
// #cgo LDFLAGS: -lcomedi -g -lm
// #include "./lib/C/io.c"
// #include "./lib/C/elev.c"
// #include "./lib/C/runner.c"
import "C"
*/

import (
	"./lib"
	"fmt"
	"math"
	"time"
	"sort"
)

type ByFloor []order

type order struct{
	floor int
	dir string
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

const DOWN 	int = 0
const UP 	int = 1
const INT 	int = 2
const N_FLOOR int = 4
const THIS_ID int = 1337


func (p order) String() string {
	return fmt.Sprintf("Floor %d, dir: %s, Del: %t", p.floor, p.dir, p.clear)
}

func contains(a order, list []order) bool {
	for _, i := range list {
		if i == a {
			return true
		}
	}
	return false
}

func remove(a order, list []order)[]order{
	var ny []order
	for _, i := range list {
		if i != a {
			ny = append(ny, i)
		}
	}
	return ny
}


func (a ByFloor) Len() int           { return len(a) }
func (a ByFloor) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByFloor) Less(i, j int) bool { return a[i].floor < a[j].floor }

func state(direction string) {
	switch {
	case direction == "DOWN":
//		fmt.Println("DOWN")
		driverInterface.SetSpeed(-300)
	case direction == "UP":
//		fmt.Println("UP")
		driverInterface.SetSpeed(300)
	case direction == "STOP":
//		fmt.Println("STOP")
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
			return math.Pow(float64(((2*orderList[len(orderList)-1].floor)-new_order-currPos)+len(orderList)), squared)
		}

	case dir_now == "DOWN" && new_order_dir == "DOWN":
		if new_order < currPos {
			return math.Pow(float64((currPos-new_order)+len(orderList)), squared)
		} else {
			return math.Pow(float64(new_order+currPos-orderList[0].floor+len(orderList)), squared)
		}

	case dir_now == "UP" && new_order_dir == "DOWN":
		return math.Pow(float64(2*orderList[len(orderList)-1].floor-currPos-new_order+len(orderList)), squared)

	case dir_now == "DOWN" && new_order_dir == "UP":
		return math.Pow(float64(currPos+new_order-orderList[0].floor+len(orderList)), squared)
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
/*	
	costRequestIn 	:= make(chan exOrder)
	costRequestOut 	:= make(chan exOrder)
	costResponsIn 	:= make(chan exOrder)
	costResponsOut 	:= make(chan exOrder)	
	
	delegate := make(chan exOrder)
	
	setOtherLights := make(chan exLights) 
*/
	var currentFloor 	int
	var lastFloor 		int
	var currentOrder	order
	var orderList 		[]order
	var afterOrders 	[]order
//	var direction		string
	var status			string
	
	go func(){
		for{
			select{	
				case floor := <-floorChannel:
					go func() {

						floor = floor - 30 //REMEMBER TO ADJUST FOR N_FLOORS
						currentFloor = floor
//						fmt.Println(currentFloor)
						if currentFloor != 0{
							lastFloor = currentFloor
						}
//						driverInterface.SetFloorLamp(driverInterface._Ctype_int(lastFloor))
						updatePos <- currentOrder
					}()		
				}
			}
	}()

	if currentFloor != 1{
		state("DOWN")
		for currentFloor!= 1{
			time.Sleep(time.Millisecond*200)
			if currentFloor!=0{
//				driverInterface.SetFloorLamp(currentFloor-1)
			}
		}
		state("STOP")
		status = "IDLE"
	}

	//SIGNAL HANDLING
	go func(){
		for{
			select{
				case a:=<-updateCurrentOrder:
					go func(){
						a=a
						if len(orderList)==0&&len(afterOrders)==0{
							state("STOP")
							updatePos <- order{-1, "NO", false}
						}else if len(orderList)==0{
							ordersChann <- order{-1, "NO", false}

						}else if status == "IDLE" || status=="UP"{
							updatePos <- orderList[0]
						}else if status=="DOWN"{
							updatePos <- orderList[len(orderList)-1]
						}
						
					}()
			
				case buttonSignal := <- intButtonChannel:
					go func(){
//						fmt.Println("INT: ", buttonSignal)
						incommingI := order{floor: buttonSignal-10, dir: "INT"}
						ordersChann <- incommingI
					}()

				case new_order := <- ordersChann:
					go func (){
//						fmt.Println("new order: ",new_order)
						test := order{-1, "NO", false}
						if new_order==test&&len(orderList)==0{
//							fmt.Println(orderList, afterOrders)
							orderList = afterOrders
							afterOrders = nil
//							fmt.Println(orderList, afterOrders)
							updateCurrentOrder <- true
						}else if new_order.clear{
//							fmt.Println("REMOVING ",orderList)
							orderList = remove(order{new_order.floor, new_order.dir, false},orderList)
///							fmt.Println("REMOVED: ",new_order, orderList)
//							fmt.Println(orderList)
							updateCurrentOrder <- true
						
						}else if !contains(new_order, orderList) && new_order.dir != "NO" {
							if status == "UP" {
								if (new_order.floor < currentFloor && currentFloor==lastFloor) || (new_order.floor < lastFloor+1&&currentFloor==0)||new_order.dir=="DOWN" {
									afterOrders = append(afterOrders, new_order)
									sort.Sort(ByFloor(afterOrders))
//									fmt.Println("yay!")
								} else {
									orderList = append(orderList, new_order)
									sort.Sort(ByFloor(orderList))
									updateCurrentOrder <- true
///									fmt.Println("nan!")
								}
							} else if status == "DOWN" {
								if (new_order.floor > currentFloor) || (new_order.floor > lastFloor-1)||new_order.dir=="UP" {
									afterOrders = append(afterOrders, new_order)
									sort.Sort(ByFloor(afterOrders))
								} else {
									orderList = append(orderList, new_order)
									sort.Sort(ByFloor(orderList))
									updateCurrentOrder <- true
								}
							} else {
								
								orderList = append(orderList, new_order)
								sort.Sort(ByFloor(orderList))
								updateCurrentOrder <- true
							}
						}
//						fmt.Println("OL: ",orderList)						
					}()



				case new_stuff := <- updatePos:
					go func(){
//						fmt.Println("proposed order: ",new_stuff)
						if !new_stuff.clear||new_stuff.dir!="NO"{
							currentOrder = new_stuff
						}
						if new_stuff.dir=="NO"{
							for currentFloor==0{
								time.Sleep(time.Millisecond*250)
							}
							status = "IDLE"
//							state("STOP")
						}else if currentOrder.floor>lastFloor{
							state("UP")
							status="UP"
						}else if currentOrder.floor<lastFloor && currentOrder.floor!=0{
							state("DOWN")
							status="DOWN"
						}else if currentOrder.floor==currentFloor{
							state("STOP")
							state("OPEN")
//							fmt.Println("GOT TO FLOOR")
							ordersChann <- order{currentOrder.floor, currentOrder.dir, true}
							updateCurrentOrder <- true
						}
//						fmt.Println("current order:", currentOrder)						
					}()

				case extSig := <- extButtonChannel:
					go func(){
//						setOtherLights <- exLights{((extSignal - (extSignal % 2) - 30) / 10), (extSignal % 2), 1}
						
						var extOrder order
						if extSig%2==0{
							extOrder = order{((extSig - (extSig % 2) - 30) / 10),"DOWN", false}
						}else{
							extOrder = order{((extSig - (extSig % 2) - 30) / 10),"UP", false}
						}
						fmt.Println(cost(orderList, afterOrders, lastFloor, status, extOrder.floor, extOrder.dir))
						ordersChann <- extOrder				
//						costRequestOut <- extOrder{floor: extSig.floor, direction: extSig.dir, origin: THIS_ID}
//						min := cost(orderList, afterOrders, lastFloor, status, extSig.floor, extSig.dir)
						//wait for cost responses for 2 sek, whilst updating min
//						delegate <-
					}()
/*				case  costReq := <-costRequestIn:
					go func(){
							costResponsOut <- extOrder{floor: costReq.floor, dir: costReq.dir, recipient: THIS_ID, origin: costReq.origin, cost: cost(orderList, afterOrders, lastFloor, status, costReq.floor, costReq.dir)}
						}()
*/


			}
		}
	}()
	
	for {
		time.Sleep(time.Second)
	}
}
