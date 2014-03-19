package main

import (
	"./lib"
	"fmt"
	//"math"
	//"sort"
	"time"
)

type order struct{
	floor 	int
	dir		int	//0: DOWN, 1: UP, 2: INT
	clear	bool
}

const DOWN 		int = 0
const UP		int = 1
const INT		int = 2
const N_FLOORS 	int = 4

func state(status string) {
	switch {
	case status == "DOWN":
		driverInterface.SetSpeed(-300)
	case status == "UP":
		driverInterface.SetSpeed(300)
	case status == "STOP":
		driverInterface.SetSpeed(300)
		driverInterface.SetSpeed(-300)
		driverInterface.SetSpeed(0)
	case status == "OPEN":
		driverInterface.SetDoorLamp(1)
		time.Sleep(2 * time.Second)
		driverInterface.SetDoorLamp(0)
	case status == "IDLE":
		
	}
}

func contains(a order, list []order) bool {
	for _, i := range list {
		if i == a {
			return true
		}
	}
	return false
}

func getSignals(intButtonChannel chan int, xBSignal chan int, orderChannel chan order){
	for {
		select {
			case buttonSignal := <- intButtonChannel:
				go func(){
					incommingI := order{floor: buttonSignal-10, dir: INT}
					orderChannel <- incommingI
				}()
			case extSignal := <- xBSignal:
				go func(){
				var incommingE order
					if extSignal%2==0{
						incommingE = order{floor:((extSignal - (extSignal % 2) - 30) / 10),dir:DOWN}
					}else{
						incommingE = order{floor:((extSignal - (extSignal % 2) - 30) / 10),dir:UP}
					}
					orderChannel <- incommingE
				}()
		}
	}
}

func main(){
	driverInterface.Init()
	
	intButtonChannel := make(chan int)
	extButtonChannel := make(chan int)
	floorChannel := make(chan int)
	stopChannel := make(chan int)
	timeoutChannel := make(chan int)
	
	driverInterface.Create(intButtonChannel, floorChannel, stopChannel, extButtonChannel, timeoutChannel)

	ordersChann := make(chan order)
	updateCurrentOrder := make(chan bool)
	updatePos := make(chan bool)
	
	var orderList 	[N_FLOORS][3]bool
//	var afterOrders	[N_FLOORS][3]bool

	var currentOrder [2] int
	
	var currentFloor 	int
	var lastFloor		int
	
	var status			string
//	var direction		string
	
	go getSignals(intButtonChannel, extButtonChannel, ordersChann)
	go func(){
		for{
			select{
				case floor := <-floorChannel:
					go func() {
						floor = floor - 30 //REMEMBER TO ADJUST FOR N_FLOORS
						currentFloor = floor
						if currentFloor != 0{
							lastFloor = currentFloor
							updatePos <- true
						}
						
//						fmt.Println(currentFloor, lastFloor)
					}() 
			}
		}
	}()
	if currentFloor != 1{
		state("DOWN")
		for currentFloor!= 1{time.Sleep(time.Millisecond*200)}
		state("STOP")
	}
	
	go func(){
		for {
			select{
/*
				case floor := <-floorChannel:
					go func() {
						floor = floor - 30 //REMEMBER TO ADJUST FOR N_FLOORS
						currentFloor = floor
						if currentFloor != 0{
							lastFloor = currentFloor
						}

						fmt.Println(currentFloor, lastFloor)
					}() 
*/
				case stop := <-stopChannel:
					go func(){
						if stop > 0 {
							fmt.Println("STOP")
						}
					}()
				
				case newPos := <-updatePos:
					go func(){
						newPos=newPos
						if currentOrder[0]==0{
							if status == "UP" || status == "DOWN"{
								state("STOP")
							}						
						}
						switch{
							case lastFloor<currentOrder[0]:
								go func(){
									status="UP"
									state("UP")
								}()
							case lastFloor>currentOrder[0]&&currentOrder[0]!=0:
								go func(){
									status="DOWN"
									state("DOWN")
								}()
							case currentFloor==currentOrder[0]:	
									fmt.Println("GOT TO THE ORDER")
									state("STOP")
									state("OPEN")
									fmt.Println(orderList)
									n := order{floor: currentOrder[0], dir:currentOrder[1], clear:true}
									ordersChann <- n
									updateCurrentOrder <- true
									updatePos <- true
						}
					}()
				
				case new_order := <-ordersChann:
					go func(){
						if new_order.clear{
							orderList[new_order.floor-1][new_order.dir]=false
						updateCurrentOrder <- true							
						}else if !orderList[new_order.floor-1][new_order.dir]{
							orderList[new_order.floor-1][new_order.dir]=true
//							fmt.Println(orderList)
						updateCurrentOrder <- true
						}
					}()
				
				case a := <-updateCurrentOrder:
					go func(){
						a = false
						fmt.Println(currentOrder, orderList)
						if status == "DOWN"{
							for i:=lastFloor; i>0;i--{
								if orderList[i-1][DOWN] || orderList[i-1][INT]{
									currentOrder[0], currentOrder[1] = i, DOWN
									break
								}
							}
							for i:=lastFloor; i<=N_FLOORS;i++{
								if orderList[i-1][UP] || orderList[i-1][INT]{
									currentOrder[0], currentOrder[1] = i, 1
									break
								}
							}
						}else if status == "UP"{
							
							for i:=lastFloor; i<=N_FLOORS;i++{
								if orderList[i-1][UP] || orderList[i-1][INT]{
									currentOrder[0], currentOrder[1] = i, 1
									break
								}
							}
							for i:=lastFloor; i>0;i--{
								if orderList[i-1][DOWN] || orderList[i-1][INT]{
									currentOrder[0], currentOrder[1] = i, 0
									break
								}
							}
						}else{
							if lastFloor<2{
								for i:=1; i<=N_FLOORS;i++{
									if orderList[i-1][UP] || orderList[i-1][INT]{
										currentOrder[0], currentOrder[1] = i, 1
										break
									}
								}
								for i:=1; i<=N_FLOORS;i++{
									if orderList[i-1][DOWN] || orderList[i-1][INT]{
										currentOrder[0], currentOrder[1] = i, 0
										break
									}
								}
							}else{
							
								for i:=1; i<=N_FLOORS;i++{
									if orderList[i-1][DOWN] || orderList[i-1][INT]{
										currentOrder[0], currentOrder[1] = i, 0
										break
									}
								}
								for i:=1; i<=N_FLOORS;i++{
									if orderList[i-1][UP] || orderList[i-1][INT]{
										currentOrder[0], currentOrder[1] = i, 1
										break
									}
								}
							}
						}
						fmt.Println(orderList)
				}()
				
			}
		}
		
	}()
	
	
	
	go func(){
		for{
			
			time.Sleep(time.Millisecond*50)
		}
	}()
	
	
	go func(){
		for{
			select {
			/*
				case newst := <-ordersChann:
					go func(){
						fmt.Println("Got order to: ", newst)
					}()
					
				case next_floor := <-floorChannel:
					go func (){//currentFloor = next_floor - 30
						fmt.Println("Current floor: ", next_floor)
					}()
					*/

			}
			
		}
	}()
	
	for{
		time.Sleep(time.Second*1)
	} 
}
