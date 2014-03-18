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
	dir		string
}

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
	}
}

func contains(a int, list []int) bool {
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
					incommingI := order{floor: buttonSignal-10, dir: "INT"}
					orderChannel <- incommingI
				}()
			case extSignal := <- xBSignal:
				go func(){
				var incommingE order
					if extSignal%2==0{
						incommingE = order{floor:((extSignal - (extSignal % 2) - 30) / 10),dir:"UP"}
					}else{
						incommingE = order{floor:((extSignal - (extSignal % 2) - 30) / 10),dir:"DOWN"}
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
	
	go getSignals(intButtonChannel, extButtonChannel, ordersChann)
	
	go func(){
		for{
			select {
				case newst := <-ordersChann:
					fmt.Println("Got order to: ", newst)
				case next_floor := <-floorChannel:
					//currentFloor = next_floor - 30
					fmt.Println("Current floor: ", next_floor)
				case stop := <-stopChannel:
					if stop > 0 {
						fmt.Println("STOP")
					}
			}
			
		}
	}()
	
	for{
		time.Sleep(time.Second*1)
	}go run 
}
