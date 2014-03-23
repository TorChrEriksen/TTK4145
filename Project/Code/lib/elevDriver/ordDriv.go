package elevDriver

import (
	"./../DataStore"
	"./lib"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"sort"
	"time"
)

type OrderDriver struct {
	N_FLOOR      int
	currentFloor int
	lastFloor    int
	currentOrder order
	orderList    []order
	afterOrders  []order
	status       string
	commDisabled bool
	GOL			map[string][]DataStore.Global_OrderData
	myIP		string

}

type ByFloor []order

type order struct {
	Floor int
	Dir   string
	Clear bool
}

func (od *OrderDriver) Create(ip string) {
	od.currentFloor = 0
	od.lastFloor = 0
	od.currentOrder = order{}
	od.orderList = make([]order, 0)
	od.afterOrders = make([]order, 0)
	od.status = "IDLE"
	od.commDisabled = true
	od.GOL = make(map[string][]DataStore.Global_OrderData)
	od.myIP = ip



}

func (p order) String() string {
	return fmt.Sprintf("Floor %d, dir: %s, Del: %t", p.Floor, p.Dir, p.Clear)
}

func contains(a order, list []order) bool {
	for _, i := range list {
		if i == a {
			return true
		}
	}
	return false
}

func remove(a order, list []order) []order {
	var ny []order
	for _, i := range list {
		if i != a {
			ny = append(ny, i)
		}
	}
	return ny
}

func removeGOL(a DataStore.Global_OrderData, list []DataStore.Global_OrderData) []DataStore.Global_OrderData {
	var ny []DataStore.Global_OrderData
	for _,i := range list {
		if i != a{
			ny = append(ny, i)
		}
	}
	return ny
}

func (a ByFloor) Len() int           { return len(a) }
func (a ByFloor) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByFloor) Less(i, j int) bool { return a[i].Floor < a[j].Floor }

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
		time.Sleep(250 * time.Millisecond)
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
			return math.Pow(float64(((2*orderList[len(orderList)-1].Floor)-new_order-currPos)+len(orderList)), squared)
		}

	case dir_now == "DOWN" && new_order_dir == "DOWN":
		if new_order < currPos {
			return math.Pow(float64((currPos-new_order)+len(orderList)), squared)
		} else {
			return math.Pow(float64(new_order+currPos-orderList[0].Floor+len(orderList)), squared)
		}

	case dir_now == "UP" && new_order_dir == "DOWN":
		return math.Pow(float64(2*orderList[len(orderList)-1].Floor-currPos-new_order+len(orderList)), squared)

	case dir_now == "DOWN" && new_order_dir == "UP":
		return math.Pow(float64(currPos+new_order-orderList[0].Floor+len(orderList)), squared)
	}
	return -1.0
}

func writeOrdersToFile(filename string, data []order) {

	b, err := json.Marshal(data)
	if err != nil {
		fmt.Println("ElevLogic: error while marshalling ", err.Error())
		return
	}

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("ElevLogic: error while creating file: ", err.Error())
		return
	}

	_, err = file.Write(b)
	if err != nil {
		fmt.Println("ElevLogic: error writing to file: ", err.Error())
		return
	}
	defer file.Close()
}
func readOrdersFromFile(filename string) ([]order, int) {

	ordrList := make([]order, 0)
	file, _ := os.Open(filename)
	if file != nil {

		b, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Println("ElevLogic: error while reading from file: ", err.Error())
			defer file.Close()
			return ordrList, -1
		}

		json.Unmarshal(b, &ordrList)
		fmt.Println(ordrList)
		defer file.Close()
		return ordrList, 1
	}
	defer file.Close()
	return ordrList, -1
}

// Clean exit, remove file
func (od *OrderDriver) Exit() {
	if od.currentFloor==0{
		if od.status == "IDLE"{
			state("DOWN")
			for od.currentFloor==0{
				time.Sleep(time.Millisecond*250)
			}
		}
	}
	os.Remove("orderList.txt")
	os.Remove("afterOrders.txt")
	// TODO
}

// TODO implement processGOL
func (od *OrderDriver) Run( toOne chan DataStore.Order_Message, toAll chan DataStore.Order_Message, recieve chan DataStore.Order_Message, commStatus chan bool, setLights chan DataStore.ExtButtons_Message, recvLights chan DataStore.ExtButtons_Message, sendGlobal chan DataStore.Global_OrderData, recvGlobal chan DataStore.Global_OrderData, processGOL chan string){
	driverInterface.Init()

	intButtonChannel := make(chan int)
	extButtonChannel := make(chan int)
	floorChannel := make(chan int)
	stopChannel := make(chan int)
	timeoutChannel := make(chan int)

	ordersChann := make(chan order)
	ordersAcosted := make(chan order)
	updateCurrentOrder := make(chan bool)
	updatePos := make(chan order)

	costResponsInternal := make(chan DataStore.Order_Message)

	driverInterface.Create(intButtonChannel, floorChannel, stopChannel, extButtonChannel, timeoutChannel)

	go func() {
		for {
			select {
			case floor := <-floorChannel:
				go func() {

					floor = floor - 30 //REMEMBER TO ADJUST FOR N_FLOORS
					od.currentFloor = floor
					fmt.Println(od.currentFloor)
					if od.currentFloor != 0 {
						od.lastFloor = od.currentFloor
						driverInterface.SetFloorLamp(od.currentFloor - 1) //TODO Find nicer way of setting the lamp?
					}
					updatePos <- od.currentOrder
				}()
			case statusChanged := <-commStatus:
				od.commDisabled = statusChanged
				fmt.Println("Comm status changed to: ", statusChanged, " | ", od.commDisabled) //TODO remove
			}
		}
	}()

	od.orderList, _ = readOrdersFromFile("orderList.txt")
	od.afterOrders, _ = readOrdersFromFile("afterOrders.txt")

	if od.currentFloor != 1 && (len(od.orderList)+len(od.afterOrders)) == 0{
		state("DOWN")
		fmt.Println("Going")
		for od.currentFloor != 1 {
			if od.currentFloor != 1 {
				time.Sleep(time.Millisecond * 200)
			}
		}
		state("STOP")
		od.status = "IDLE"
	} else if od.currentFloor==0{
		state("DOWN")
		for od.currentFloor ==0{time.Sleep(time.Millisecond*250)}
			updateCurrentOrder <- true
	}else {
		updateCurrentOrder <- true
	}
	//SIGNAL HANDLING
	go func() {
		for {
			select {

			case a := <-updateCurrentOrder:
				go func() {
					a = a
					if (len(od.orderList) == 0) && (len(od.afterOrders) == 0) {
						state("STOP")
						updatePos <- order{-1, "NO", false}
					} else if len(od.orderList) == 0 {
						ordersChann <- order{-1, "NO", false}

					} else if od.status == "IDLE" || od.status == "UP" {
						sort.Sort(ByFloor(od.afterOrders))
						updatePos <- od.orderList[0]
					} else if od.status == "DOWN" {
						updatePos <- od.orderList[len(od.orderList)-1]
					}

				}()

			case buttonSignal := <-intButtonChannel:
				go func() {
					//						fmt.Println("INT: ", buttonSignal)
					incommingInternal := order{Floor: buttonSignal - 10, Dir: "INT"}
					driverInterface.SetButtonLamp(incommingInternal.Dir, incommingInternal.Floor-1, 1) //TODO make nicer. handle floors better
					ordersChann <- incommingInternal
				}()

			case new_order := <-ordersChann:
				go func() {
					//						fmt.Println("new order: ",new_order)
					test := order{-1, "NO", false}
					if new_order == test && len(od.orderList) == 0 {
						//							fmt.Println(od.orderList, od.afterOrders)
						od.orderList = od.afterOrders
						od.afterOrders = nil
						writeOrdersToFile("orderList.txt", od.orderList)
						writeOrdersToFile("afterOrders.txt", od.afterOrders)
						//							fmt.Println(od.orderList, od.afterOrders)
						updateCurrentOrder <- true
					} else if new_order.Clear {
						od.orderList = remove(order{new_order.Floor, new_order.Dir, false}, od.orderList)
						writeOrdersToFile("orderList.txt", od.orderList)
						///							fmt.Println("REMOVED: ",new_order, od.orderList)
						//							fmt.Println(od.orderList)
						updateCurrentOrder <- true

					} else if !contains(new_order, od.orderList) && new_order.Dir != "NO" {
						if od.status == "UP" {
							if (new_order.Floor < od.currentFloor && od.currentFloor == od.lastFloor) || (new_order.Floor < od.lastFloor+1 && od.currentFloor == 0) || new_order.Dir == "DOWN" {
								od.afterOrders = append(od.afterOrders, new_order)
								sort.Sort(ByFloor(od.afterOrders))
								writeOrdersToFile("afterOrders.txt", od.afterOrders)
								updateCurrentOrder <- true
								//									fmt.Println("yay!")
							} else {
								od.orderList = append(od.orderList, new_order)
								sort.Sort(ByFloor(od.orderList))
								writeOrdersToFile("orderList.txt", od.orderList)
								updateCurrentOrder <- true
								///									fmt.Println("nan!")
							}
						} else if od.status == "DOWN" {
							if (new_order.Floor > od.currentFloor && od.currentFloor == od.lastFloor) || (new_order.Floor > od.lastFloor-1 && od.currentFloor == 0) || new_order.Dir == "UP" {
								od.afterOrders = append(od.afterOrders, new_order)
								sort.Sort(ByFloor(od.afterOrders))
								writeOrdersToFile("afterOrders.txt", od.afterOrders)
								updateCurrentOrder <- true
							} else {
								od.orderList = append(od.orderList, new_order)
								sort.Sort(ByFloor(od.orderList))
								writeOrdersToFile("orderList.txt", od.orderList)
								updateCurrentOrder <- true
							}
						} else {

							od.orderList = append(od.orderList, new_order)
							sort.Sort(ByFloor(od.orderList))
							writeOrdersToFile("orderList.txt", od.orderList)
							updateCurrentOrder <- true
						}
					}
					//						fmt.Println("OL: ",od.orderList)						
				}()

			case new_stuff := <-updatePos:
				go func() {
					//						fmt.Println("proposed order: ",new_stuff)
					//						fmt.Println(od.status)
					if !new_stuff.Clear || new_stuff.Dir != "NO" {
						od.currentOrder = new_stuff
					}
					if new_stuff.Dir == "NO" {
						for od.currentFloor == 0 {
							time.Sleep(time.Millisecond * 250)
						}
						od.status = "IDLE"
						//							state("STOP")
					} else if len(od.orderList) == 0 && len(od.afterOrders) == 0 {
						state("IDLE")
					} else if od.currentOrder.Floor == od.currentFloor {
						state("STOP")
						driverInterface.SetButtonLamp(od.currentOrder.Dir, od.currentFloor-1, 0)
						if od.currentOrder.Dir != "INT" {
							setLights <- DataStore.ExtButtons_Message{Floor: od.currentFloor - 1, Dir: od.currentOrder.Dir, Value: 0}
							sendGlobal <- DataStore.Global_OrderData{Floor:od.currentFloor, Dir:od.currentOrder.Dir,HandlingIP:od.myIP, Clear:true}
						}
						state("OPEN")
						//							fmt.Println("GOT TO FLOOR")
						if od.currentFloor == 1 {
							od.status = "UP"
						} else if od.currentFloor == od.N_FLOOR {
							od.status = "DOWN"
						}
						
						ordersChann <- order{od.currentOrder.Floor, od.currentOrder.Dir, true}
						
						updateCurrentOrder <- true
						
					} else if od.currentOrder.Floor > od.lastFloor && od.currentFloor != od.N_FLOOR {
						state("UP")
						od.status = "UP"

					
					} else if od.currentOrder.Floor < od.lastFloor && od.currentOrder.Floor != 0 {
						state("DOWN")
						od.status = "DOWN"
					}
					//						fmt.Println("current order:", od.currentOrder)						
				}()

			case extSig := <-extButtonChannel:

				// TODO Queue push?


				go func() {

					// TODO QUeue pop and loop untill queue empty?

					//						setOtherLights <- exLights{((extSignal - (extSignal % 2) - 30) / 10), (extSignal % 2), 1}

					var extOrder order
					if extSig%2 == 0 {
						extOrder = order{((extSig - (extSig % 2) - 30) / 10), "DOWN", false}
						driverInterface.SetButtonLamp("DOWN", ((extSig-(extSig%2)-30)/10)-1, 1) //TODO make nicer. handle floors better
						//fmt.Println("ExtOrder: ", extOrder)
					} else {
						extOrder = order{((extSig - (extSig % 2) - 30) / 10), "UP", false}
						driverInterface.SetButtonLamp("UP", ((extSig-(extSig%2)-30)/10)-1, 1) //TODO make nicer. handle floors better
						//fmt.Println("ExtOrder: ", extOrder)
					}
					setLights <- DataStore.ExtButtons_Message{Floor: extOrder.Floor, Dir: extOrder.Dir, Value: 1}
					ordersAcosted <- extOrder
				}()
				
					//						fmt.Println(cost(od.orderList, od.afterOrders, od.lastFloor, od.status, extOrder.floor, extOrder.dir))
					//		ordersChann <- extOrder
					//}()


			case acosting := <- ordersAcosted:
				go func(){
					// OriginIP is set in Application Control 
				
					min := DataStore.Order_Message{Floor: acosting.Floor, Dir: acosting.Dir, RecipientIP: od.myIP, Cost: cost(od.orderList, od.afterOrders, od.lastFloor, od.status, acosting.Floor, acosting.Dir), What: "COST_REQ"}
					req := min
					if !contains(acosting, od.orderList) && !od.commDisabled {

						toAll <- req
						go func() {
							got := <-costResponsInternal
								fmt.Println("GOT A RESPONSE!!")
							if got.Cost < min.Cost {
								min = got
							}
						}()

						time.Sleep(time.Second * 2)

						sendGlobal <- DataStore.Global_OrderData{Floor:min.Floor, Dir:min.Dir, HandlingIP:min.OriginIP, Clear:false}

						if min.OriginIP == od.myIP {
							ordersChann <- order{min.Floor, min.Dir, false}
							

						} else {
//							min.What min.RecipientIP, min.OriginIP = "O_REQ", min.OriginIP, min.RecipientIP
							toOne <- DataStore.Order_Message{Floor: min.Floor, Dir: min.Dir, RecipientIP: min.OriginIP, What: "O_REQ"}

						}
					} else if od.commDisabled {
						ordersChann <- order { Floor:min.Floor, Dir:min.Dir, Clear:false }
					}
				}()

			case input := <-recieve:
				go func() {
					if input.What == "COST_REQ" {
						fmt.Println("Getting a cost request", input)
						toOne <- DataStore.Order_Message{Floor: input.Floor, Dir: input.Dir, RecipientIP: input.OriginIP, OriginIP: input.RecipientIP, Cost: cost(od.orderList, od.afterOrders, od.lastFloor, od.status, input.Floor, input.Dir), What: "COST_RES"}
					} else if input.What == "COST_RES" {
						fmt.Println("Getting a cost respons")
						costResponsInternal <- input
					} else if input.What == "O_REQ" {
						fmt.Println(("Getting an Order from OUTSIDE"))
						ordersChann <- order{input.Floor, input.Dir, false}
					}
				}()

			case lit := <-recvLights:
				go func() {
					fmt.Println("Oh God we have to set lights")
					driverInterface.SetButtonLamp(lit.Dir, lit.Floor-1, lit.Value)
				}()

// se om vi skal sette den utenfor, og bruke buffer på chan for å unngå forvirring i ordre.
			case ipDown := <-processGOL:
				go func() {
					fmt.Println("OH NO! A computer is down")
					for _,i := range od.GOL[ipDown]{
						ordersAcosted <- order{Floor:i.Floor, Dir:i.Dir, Clear:false}
					}
				}()
			
			case updateGOL := <- recvGlobal:
				go func(){
//					_, exists := od.GOL[updateGOL.HandlingIP]					
	//				if !exists{
		//				od.GOL[updateGOL.HandlingIP] = DataStore.Received_OrderData{}
			//			od.GOL[updateGOL.HandlingIP].OrderList = make([]DataStore.Global_OrderData,0)
				//	}	else
					if !updateGOL.Clear{
						od.GOL[updateGOL.HandlingIP] = append(od.GOL[updateGOL.HandlingIP], DataStore.Global_OrderData{Floor: updateGOL.Floor, Dir:updateGOL.Dir, HandlingIP: updateGOL.HandlingIP})
					} else {
						od.GOL[updateGOL.HandlingIP] = removeGOL(DataStore.Global_OrderData{Floor: updateGOL.Floor, Dir:updateGOL.Dir}, od.GOL[updateGOL.HandlingIP])
					}
				}()

			}
		}
	}()
}
