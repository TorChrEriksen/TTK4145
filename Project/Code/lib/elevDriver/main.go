// Elevator control logic
/*Elvator functions:

*/
package main

import (
	"./lib"
	"fmt"
	"sort"
	"time"
	//"os"
)
const THIS_ID = 1


func main() {
	driverInterface.Init()

	intButtonChannel 	:= make(chan int)
	extButtonChannel 	:= make(chan int)
	floorChannel 		:= make(chan int)
	stopChannel 		:= make(chan int)
	timeoutChannel 		:= make(chan int)

	driverInterface.Create(intButtonChannel, floorChannel, stopChannel, extButtonChannel, timeoutChannel)

	var orderList 		[]int
	var afterOrders 	[]int
	var status 			string
	var currentFloor 	int

	gatherCost 	:= make(chan struct)
	replyCost	:= make(chan struct)

	pollCost	:= make(chan struct)
	recieveCost	:= make(chan struct)

	ordersChann := make(chan int)


	//test 1: 		Get signals correct
	//test 1.1	 	Create internal orders correctly
	//test 1.2:		Create ext. orders correctly

	//test 2: 		Drive
	//test 2.1: 		Exchange afterOrders

	//test 3:		Send for external orders
	//test 3.1:		Cost function
	//test 3.2:		Pollhandling for cost

	//test 4:		Stop routine
	//test 5:		Init routine
	//test 6: 		Obstruct routine



	// test 3:		
}
