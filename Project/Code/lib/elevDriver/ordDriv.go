package ordDriv

import (
	"./lib"
	"fmt"
	"math"
	"time"
	"sort"
)


type OrderDriver struct{




	currentFloor 	int
	lastFloor 		int
	currentOrder	order
	orderList 		[]order
	afterOrders 	[]order
	status			string	

}

type ByFloor []order

type order struct{
	floor int
	dir string
	clear bool
}
//TODO: DATASTORE
type exOrder struct{
	floor		int
	dir 		int
	recipient	int
	origin		int
	cost		float64
	what		string
}

type exLights struct{
	floor 	int
	dir		int
	value	int
	}


func (od *OrderDriver) Create() {
	od.currentFloor = 0
	od.lastFloor = 0
	od.currentOrder = 0	
	od.orderList = make([]order, 0)
	od.afterOrders = make([]order, 0)
	od.status = "IDLE"
}

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


func (od *OrderDriver) Run( toOne chan exOrder, toAll chan exOrder, recieved chan exOrder){
	driverInterface.Init()
	
	intButtonChannel 	:= make(chan int)
	extButtonChannel 	:= make(chan int)
	floorChannel 		:= make(chan int)
	stopChannel 		:= make(chan int)
	timeoutChannel 		:= make(chan int)

	ordersChann 		:= make(chan order)
	updatecurrentOrder 	:= make(chan bool)
	updatePos 			:= make(chan order)


	driverInterface.Create(intButtonChannel, floorChannel, stopChannel, extButtonChannel, timeoutChannel)
	


	go func(){
		for{
			select{	
				case floor := <-floorChannel:
					go func() {

						floor = floor - 30 //REMEMBER TO ADJUST FOR N_FLOORS
						od.currentFloor = floor
						fmt.Println(od.currentFloor)
						if od.currentFloor != 0{
							od.lastFloor = od.currentFloor
							driverInterface.SetFloorLamp(od.currentFloor-1) //TODO Find nicer way of setting the lamp?
						}
						updatePos <- od.currentOrder
					}()		
				}
			}
	}()

	if od.currentFloor != 1{
		state("DOWN")
		for od.currentFloor!= 1{
			time.Sleep(time.Millisecond*200)
			if od.currentFloor!=0{
			}
		}
		state("STOP")
		od.status = "IDLE"
	}

	//SIGNAL HANDLING
	go func(){
		for{
			select{

				case a:=<-updateCurrentOrder:
					go func(){
						a=a
						if len(od.orderList)==0&&len(od.afterOrders)==0{
							state("STOP")
							updatePos <- order{-1, "NO", false}
						}else if len(od.orderList)==0{
							ordersChann <- order{-1, "NO", false}

						}else if od.status == "IDLE" || od.status=="UP"{
							sort.Sort(ByFloor(od.afterOrders))
							updatePos <- od.orderList[0]
						}else if od.status=="DOWN"{
							updatePos <- od.orderList[len(od.orderList)-1]
						}
						
					}()
			
				case buttonSignal := <- intButtonChannel:
					go func(){
//						fmt.Println("INT: ", buttonSignal)
						incommingI := order{floor: buttonSignal-10, dir: "INT"}
						driverInterface.SetButtonLamp(incommingI.dir, incommingI.floor-1, 1 ) //TODO make nicer. handle floors better
						ordersChann <- incommingI
					}()

				case new_order := <- ordersChann:
					go func (){
//						fmt.Println("new order: ",new_order)
						test := order{-1, "NO", false}
						if new_order==test&&len(od.orderList)==0{
//							fmt.Println(od.orderList, od.afterOrders)
							od.orderList = od.afterOrders
							od.afterOrders = nil
//							fmt.Println(od.orderList, od.afterOrders)
							updateod.currentOrder <- true
						}else if new_order.clear{
							od.orderList = remove(order{new_order.floor, new_order.dir, false},od.orderList)
///							fmt.Println("REMOVED: ",new_order, od.orderList)
//							fmt.Println(od.orderList)
							updateod.currentOrder <- true
						
						}else if !contains(new_order, od.orderList) && new_order.dir != "NO" {
							if od.status == "UP" {
								if (new_order.floor < od.currentFloor && od.currentFloor==od.lastFloor) || (new_order.floor < od.lastFloor+1&&od.currentFloor==0)||new_order.dir=="DOWN" {
									od.afterOrders = append(od.afterOrders, new_order)
									sort.Sort(ByFloor(od.afterOrders))
//									fmt.Println("yay!")
								} else {
									od.orderList = append(od.orderList, new_order)
									sort.Sort(ByFloor(od.orderList))
									updateCurrentOrder <- true
///									fmt.Println("nan!")
								}
							} else if od.status == "DOWN" {
								if (new_order.floor > od.currentFloor && od.currentFloor==od.lastFloor) || (new_order.floor > od.lastFloor-1&&od.currentFloor==0)||new_order.dir=="UP" {
									od.afterOrders = append(od.afterOrders, new_order)
									sort.Sort(ByFloor(od.afterOrders))
								} else {
									od.orderList = append(od.orderList, new_order)
									sort.Sort(ByFloor(od.orderList))
									updateCurrentOrder <- true
								}
							} else {
								
								od.orderList = append(od.orderList, new_order)
								sort.Sort(ByFloor(od.orderList))
								updateCurrentOrder <- true
							}
						}
//						fmt.Println("OL: ",od.orderList)						
					}()

				case new_stuff := <- updatePos:
					go func(){
//						fmt.Println("proposed order: ",new_stuff)
//						fmt.Println(od.status)
						if !new_stuff.clear||new_stuff.dir!="NO"{
							od.currentOrder = new_stuff
						}
						if new_stuff.dir=="NO"{
							for od.currentFloor==0{
								time.Sleep(time.Millisecond*250)
							}
							od.status = "IDLE"
//							state("STOP")
						}else if len(od.orderList)==0 && len(od.afterOrders)==0{
							state("IDLE")
						}else if od.currentOrder.floor==od.currentFloor{
							state("STOP")
							driverInterface.SetButtonLamp(od.currentOrder.dir, od.currentFloor-1, 0)
							state("OPEN")	
//							fmt.Println("GOT TO FLOOR")
							if od.currentFloor==1{
								od.status="UP"
							}else if od.currentFloor==N_FLOOR{
								od.status="DOWN"
							}
							ordersChann <- order{od.currentOrder.floor, od.currentOrder.dir, true}
							updateod.currentOrder <- true
						
						}else if od.currentOrder.floor>od.lastFloor{
							state("UP")
							od.status="UP"

						}else if (od.currentOrder.floor<od.lastFloor && od.currentOrder.floor!=0){
							state("DOWN")
							od.status="DOWN"
						
						}
//						fmt.Println("current order:", od.currentOrder)						
					}()

				case extSig := <- extButtonChannel:

					// TODO Queue push?

					go func(){

						// TODO QUeue pop and loop untill queue empty?

//						setOtherLights <- exLights{((extSignal - (extSignal % 2) - 30) / 10), (extSignal % 2), 1}
						
						var extOrder order
						if extSig%2==0{
							extOrder = order{((extSig - (extSig % 2) - 30) / 10),"DOWN", false}
							driverInterface.SetButtonLamp("DOWN", ((extSig - (extSig % 2) - 30) / 10)-1, 1 ) //TODO make nicer. handle floors better
							//fmt.Println("ExtOrder: ", extOrder)
						}else{
							extOrder = order{((extSig - (extSig % 2) - 30) / 10),"UP", false}
							driverInterface.SetButtonLamp("UP", ((extSig - (extSig % 2) - 30) / 10)-1, 1 ) //TODO make nicer. handle floors better
							//fmt.Println("ExtOrder: ", extOrder)
						}
//						fmt.Println(cost(od.orderList, od.afterOrders, od.lastFloor, od.status, extOrder.floor, extOrder.dir))
				//		ordersChann <- extOrder
					}()
				
						
						if !contains(extOrder, od.orderList)&& THIS_ID!=-1{
							min:=exOrder(floor:extOrder.floor, dir:extOrder.dir, origin:THIS_ID, cost:cost(od.orderList, od.afterOrders, od.lastFloor, od.status, extOrder.floor, extOrder.dir), recipient:THIS_ID, what: "COST_REQ") 
							var got exOrder
							toAll <- min
							go func(){
								got =<- costResponsIn
								if got.cost<min.cost{
									min=got
								}
							}()

							time.Sleep(time.Second * 2)

							if min.recipient==THIS_ID{
								ordersChann <- order(min.floor, floor.dir, false)	

							}else{
								min.what, min.recipient, min.origin = "O_REQ", min.origin, min.recipient
								toOne<-min
							}

						}else{
							ordersChann <- order(min.floor, floor.dir, false)
						}

					}()


				case input := <- recieve:
					go func(){
						if input.what=="COST_REQ"{
							toOne <-extOrder{floor: input.floor, dir: input.dir, recipient: input.origin, origin: input.recipient, cost: cost(od.orderList, od.afterOrders, od.lastFloor, od.status, costReq.floor, costReq.dir), what:"COST_RES"}
						}else if input.what=="COST_RES"{
							costResponsIn <- input
						}else if input.what=="O_REQ"{
							ordersChann <- order(input.floor, input.dir, false)
						}
						}()
			}()
			}
		}()


}
