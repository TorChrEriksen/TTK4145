// Elevator control logic
package main

//EXT orders to chann
//EXT orders struct
//EXT orders light struct and handle

import (
	"./lib"
	"fmt"
	"math"
	"sort"
	"time"
	//"os"
)

const THIS_ID = 1

type gatherCost struct {
	floor     int
	direction string
	origin    int
	recipient int
}

//works
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

func contains(a int, list []int) bool {
	for _, i := range list {
		if i == a {
			return true
		}
	}
	return false
}

func manageNrun(ordersIn chan int, floorChan chan int, stopChan chan int) {
	var orderList []int
	var afterOrders []int

	//var currentOrder 	int
	var currentFloor int
	var lastFloor int

	var status string
	var direction string //UP, DOWN, IDLE
	direction = "IDLE"
	fmt.Println("Starting mNr")

	go func() {
		for {
			//get and manage signals from channels
			select {
			//Handle incomming orders
			//REMEMBER TO LOCKDOWN ORDERLIST WHEN CHANGING, OTHERWISE USE CURRENTORDER
			case new_order := <-ordersIn:
				fmt.Println(new_order)
				if !contains(new_order, orderList) {
					if direction == "UP" {
						if new_order < currentFloor || new_order < lastFloor+1 {
							afterOrders = append(afterOrders, new_order)
							sort.Ints(afterOrders)
						} else {
							orderList = append(orderList, new_order)
							sort.Ints(orderList)
						}
					} else if direction == "DOWN" {
						if (new_order > currentFloor) || (new_order > lastFloor-1) {
							afterOrders = append(afterOrders, new_order)
							sort.Ints(afterOrders)
						} else {
							orderList = append(orderList, new_order)
							sort.Ints(orderList)
						}
					} else {
						orderList = append(orderList, new_order)
						sort.Ints(orderList)
					}
				}
				//Update current and last floor
			case floor := <-floorChan:
				go func() {
					floor = floor - 30 //REMEMBER TO ADJUST FOR N_FLOORS
					if currentFloor != 0 && currentFloor != lastFloor {
						lastFloor = currentFloor
					}
					currentFloor = floor
					//fmt.Println(currentFloor)
				}()
			//Listen for STOP signal
			case stop := <-stopChan:
				go func() {
					if stop > 0 {
						status = "STOP"
					}
				}()
			}
		}
	}()
	/*
		go func(){
			for{
				//fmt.Println(len(orderList), orderList)
				switch{

					case (currentFloor == 4) || (currentFloor == 1):
						state("STOP")
					case len(orderList)%2==0:
						state("UP")
					case len(orderList)%2==1:
						state("DOWN")
				}
				time.Sleep(time.Millisecond*200)
			}
		}()*/
	//FOR EN IDIOTI!! MAN MÅ HUSKE Å GI DEN TIIIIIDDDDD!!!!!!
	go func() {
		for {
			fmt.Println("HTO")
			switch {
			//STOP and WAIT
			case len(orderList) == 0 && len(afterOrders) == 0:
				fmt.Println("Case 1")
				//for currentFloor==0{}
				status = "IDLE"
				state("STOP")

				//FIGHT THE TROLLS AND GET THAT BACKLOG!
			case len(orderList) == 0 && len(afterOrders) != 0:
				orderList, afterOrders = afterOrders, nil

			//Bør bruke currentOrder som placeholder slik at det ikke blir fucka med sanntid, og vil gjøre det enklere å skru av på lys
			case len(orderList) != 0:
				fmt.Println("Case 2")
				for len(orderList) > 0 {
					// go UP
					if orderList[0] > lastFloor {
						status = "UP"
						state("UP")
						for orderList[0] != currentFloor {
							time.Sleep(10 * time.Millisecond)
						}
						state("STOP")
						state("OPEN")
						orderList = orderList[1:]
						// STOP and GET IT
					} else if orderList[len(orderList)-1] == currentFloor || currentFloor == orderList[0] {
						state("STOP")
						state("OPEN")
						// go DOWN
					} else {
						status = "DOWN"
						state("DOWN")
						for currentFloor != orderList[len(orderList)-1] {
							time.Sleep(10 * time.Millisecond)
						}
						state("STOP")
						state("OPEN")
						orderList = orderList[:len(orderList)-1]
					}
				}
			}
			time.Sleep(time.Millisecond * 250)
		}
	}()
}

//this method works, send orders through chann's, TOMORROW: extHandling
func getSignals(intButtonChannel chan int, xBSignal chan int, orderChannel chan int) {
	for {
		select {
		case buttonSignal := <-intButtonChannel:
			go func() {
				buttonSignal = buttonSignal - 10
				//fmt.Println(buttonSignal)
				orderChannel <- buttonSignal
			}()
		case extOrder := <-xBSignal:
			go func() {
				extOrder = ((extOrder - (extOrder % 2) - 30) / 10)
				//fmt.Println(extOrder)
				orderChannel <- extOrder
			}()
		}
	}
}

func test(ordersChann chan int, floorChannel chan int, stopChannel chan int) {
	select {
	case newst := <-ordersChann:
		fmt.Println("Got order to: ", newst)
	case next_floor := <-floorChannel:
		currentFloor := next_floor - 30
		fmt.Println(currentFloor)
	case stop := <-stopChannel:
		if stop > 0 {
			fmt.Println("STOP")
		}
	}
}

//add afterOrderList to eq.
func cost(orderList []int, afterOrderList []int, currPos int, dir_now int, new_order int, new_order_dir int) float64 {
	var squared float64
	squared = 2.0
	switch {
	case dir_now == 1 && new_order_dir == 1:
		if currPos < new_order {
			return math.Pow(float64((new_order-currPos)+len(orderList)), squared)
		} else {
			return math.Pow(float64(((2*orderList[len(orderList)-1])-new_order-currPos)+len(orderList)), squared)
		}

	case dir_now == 0 && new_order_dir == 0:
		if new_order < currPos {
			return math.Pow(float64((currPos-new_order)+len(orderList)), squared)
		} else {
			return math.Pow(float64(new_order+currPos-orderList[0]+len(orderList)), squared)
		}

	case dir_now == 1 && new_order_dir == 0:
		return math.Pow(float64(2*orderList[len(orderList)-1]-currPos-new_order+len(orderList)), squared)

	case dir_now == 0 && new_order_dir == 1:
		return math.Pow(float64(currPos+new_order-orderList[0]+len(orderList)), squared)
	}
	return 0.0
}

func main() {
	fmt.Println("Interface init starting.")
	driverInterface.Init()
	fmt.Println("Interface init finished.")
	intButtonChannel := make(chan int)
	extButtonChannel := make(chan int)
	floorChannel := make(chan int)
	stopChannel := make(chan int)
	timeoutChannel := make(chan int)
	fmt.Println("Interface creating channels")
	driverInterface.Create(intButtonChannel, floorChannel, stopChannel, extButtonChannel, timeoutChannel)
	fmt.Println("Done")

	//var orderList 		[]int
	//var afterOrders 	[]int
	//var status 			string
	//var currentFloor 	int
	//var lastFloor       int
	//var direction       string

	//reapCost 	:= make(chan gatherCost)
	//replyCost	:= make(chan gatherCost)
	/*
		pollCost	:= make(chan struct)
		recieveCost	:= make(chan struct)
	*/
	ordersChann := make(chan int)
	driverInterface.SetSpeed(0)
	driverInterface.SetDoorLamp(1)
	fmt.Println("testing, testing")
	//lastFloor = 0
	//go manageNrun(ordersChann, floorChannel, stopChannel)
	go getSignals(intButtonChannel, extButtonChannel, ordersChann)
	for {
		time.Sleep(1 * time.Second)

		//go test(ordersChann, floorChannel, stopChannel)
		go func() {
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
		}()

	}
}

//TOMORROW FIRST THING: MERGE RUN AND MANAGE, ADD STOP AND CURRENT FLOOR SIGNALS, SETUP INFO CHANN STRUCT.
/*
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
}*/

/*
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
*/
