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
	GOL          map[string][]DataStore.Global_OrderData
	myIP         string
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

// Checks orderlist for element
func contains(a order, list []order) bool {
	for _, i := range list {
		if i == a {
			return true
		}
	}
	return false
}

// Remove element from orderlist
func remove(a order, list []order) []order {
	var ny []order
	for _, i := range list {
		if i != a {
			ny = append(ny, i)
		}
	}
	return ny
}

// Remove element from GOL
func removeGOL(a DataStore.Global_OrderData, list []DataStore.Global_OrderData) []DataStore.Global_OrderData {
	var ny []DataStore.Global_OrderData
	for _, i := range list {
		if i != a {
			ny = append(ny, i)
		}
	}
	return ny
}

// func's for sort interface
func (a ByFloor) Len() int           { return len(a) }
func (a ByFloor) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByFloor) Less(i, j int) bool { return a[i].Floor < a[j].Floor }


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
		time.Sleep(250 * time.Millisecond)
		driverInterface.SetDoorLamp(0)
	}
}

func cost(orderList []order, afterOrderList []order, currPos int, dir_now string, new_order int, new_order_dir string) float64 {
	var squared float64
	squared = 2.0
	switch {
	case dir_now == "UP" && new_order_dir == "UP" && len(orderList)!=0:
		if currPos < new_order {
			return math.Pow(float64((new_order-currPos)+len(orderList)), squared)
		} else {
			return math.Pow(float64(((2*orderList[len(orderList)-1].Floor)-new_order-currPos)+len(orderList)), squared)
		}

	case dir_now == "DOWN" && new_order_dir == "DOWN" && len(orderList)!=0:
		if new_order < currPos {
			return math.Pow(float64((currPos-new_order)+len(orderList)), squared)
		} else {
			return math.Pow(float64(new_order+currPos-orderList[0].Floor+len(orderList)), squared)
		}

	case dir_now == "UP" && new_order_dir == "DOWN" && len(orderList)!=0:
		return math.Pow(float64(2*orderList[len(orderList)-1].Floor-currPos-new_order+len(orderList)), squared)

	case dir_now == "DOWN" && new_order_dir == "UP" && len(orderList)!=0:
		return math.Pow(float64(currPos+new_order-orderList[0].Floor+len(orderList)), squared)

	default:
		return math.Pow(math.Abs(float64(currPos-new_order)), squared)
	}
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

// Clean exit, remove file, go to nearest floor, stop and open the door
func (od *OrderDriver) Exit() {
	driverInterface.SetSpeed(0)
	os.Remove("orderList.txt")
	os.Remove("afterOrders.txt")
}

func (od *OrderDriver) Run(toOne chan DataStore.Order_Message, toAll chan DataStore.Order_Message, recieve chan DataStore.Order_Message, commDisabled chan bool, setLights chan DataStore.ExtButtons_Message, recvLights chan DataStore.ExtButtons_Message, sendGlobal chan DataStore.Global_OrderData, recvGlobal chan DataStore.Global_OrderData, processGOL chan string) {
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

	oldOrders := false
	// Reading Floor and communication status
	go func() {
		for {
			select {
			case floor := <-floorChannel:
				go func() {

					floor = floor - 30 
					od.currentFloor = floor
					fmt.Println(od.currentFloor)
					if od.currentFloor != 0 {
						od.lastFloor = od.currentFloor
						driverInterface.SetFloorLamp(od.currentFloor - 1)
					}
					updatePos <- od.currentOrder
				}()
			case statusChanged := <-commDisabled:
				od.commDisabled = statusChanged
				fmt.Println("Comm status changed to: ", statusChanged, " | ", od.commDisabled) //TODO remove
			}
		}
	}()

	// Reading old Orders if possible
	od.orderList, _ = readOrdersFromFile("orderList.txt")
	od.afterOrders, _ = readOrdersFromFile("afterOrders.txt")

	//if no Old orders goto 1 floor and IDLE
	if od.currentFloor != 1 && (len(od.orderList)+len(od.afterOrders)) == 0 {
		state("DOWN")
		fmt.Println("Going")
		for od.currentFloor != 1 {
			time.Sleep(time.Millisecond * 10)
		}
		state("STOP")
		od.status = "IDLE"

	//else go to nearest floor and update order.
	} else if od.currentFloor == 0 {
		oldOrders = true
		state("DOWN")
		for od.currentFloor == 0 {
			time.Sleep(time.Millisecond * 10)
		}
		state("STOP")
	} else {
		oldOrders = true
		state("STOP")
		
	}

	go func() {
		for{
			new_order := <-ordersChann
				test := order{-1, "NO", false}
				if len(od.afterOrders) != 0 && len(od.orderList) == 0 && new_order==test {

					od.orderList = od.afterOrders
					od.afterOrders = []order(nil)
					writeOrdersToFile("orderList.txt", od.orderList)
					writeOrdersToFile("afterOrders.txt", od.afterOrders)

					updateCurrentOrder <- true
				} else if new_order.Clear {
					od.orderList = remove(order{new_order.Floor, new_order.Dir, false}, od.orderList)
					writeOrdersToFile("orderList.txt", od.orderList)

					updateCurrentOrder <- true

				} else if !contains(new_order, od.orderList) && !contains(new_order, od.afterOrders) && new_order.Dir != "NO" {
					if od.status == "UP" {
						if (new_order.Floor < od.currentFloor && od.currentFloor == od.lastFloor) || (new_order.Floor < od.lastFloor+1 && od.currentFloor == 0) || new_order.Dir == "DOWN" {
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
		}	
	} ()

	go func() {
		for {
			new_stuff := <-updatePos
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
					if od.currentOrder.Dir==od.status || od.currentOrder.Dir=="INT" || len(od.orderList)==1{
						state("STOP")
					
						driverInterface.SetButtonLamp(od.currentOrder.Dir, od.currentFloor-1, 0)
						if od.currentOrder.Dir != "INT" {
							setLights <- DataStore.ExtButtons_Message{Floor: od.currentFloor, Dir: od.currentOrder.Dir, Value: 0}
							sendGlobal <- DataStore.Global_OrderData{Floor: od.currentFloor, Dir: od.currentOrder.Dir, HandlingIP: od.myIP, Clear: true}
						}
						state("OPEN")
						ordersChann <- order{od.currentOrder.Floor, od.currentOrder.Dir, true}
						} else {
							ordersChann <- order{od.currentOrder.Floor, od.currentOrder.Dir, true}
							ordersChann <- order{od.currentOrder.Floor, od.currentOrder.Dir, false}
						}
/*
					if od.currentFloor == 1 {
						od.status = "UP"
					} else if od.currentFloor == od.N_FLOOR {
						od.status = "DOWN"
					}*/
					updateCurrentOrder <- true

				} else if od.currentOrder.Floor > od.lastFloor && od.currentFloor != od.N_FLOOR {
					state("UP")
					od.status = "UP"

				} else if od.currentOrder.Floor < od.lastFloor && od.currentOrder.Floor != 0 {
					state("DOWN")
					od.status = "DOWN"
				}

		}
	} ()

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

					} else if (od.status == "IDLE" && od.lastFloor<3) || (od.status == "UP") {
//						sort.Sort(ByFloor(od.afterOrders)) // TODO
						updatePos <- od.orderList[0]
					} else if (od.status == "IDLE" && od.lastFloor>=3) || (od.status == "DOWN") {
						updatePos <- od.orderList[len(od.orderList)-1]
					}
				}()

			case buttonSignal := <-intButtonChannel:
				go func() {
					incommingInternal := order{Floor: buttonSignal - 10, Dir: "INT"}
					driverInterface.SetButtonLamp(incommingInternal.Dir, incommingInternal.Floor-1, 1) 
					ordersChann <- incommingInternal
				}()
/*
			case new_stuff := <-updatePos:
				go func() {

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
						if od.currentOrder.Dir==od.status || od.currentOrder.Dir=="INT" || len(od.orderList)==1{
							state("STOP")
						
							driverInterface.SetButtonLamp(od.currentOrder.Dir, od.currentFloor-1, 0)
							if od.currentOrder.Dir != "INT" {
								setLights <- DataStore.ExtButtons_Message{Floor: od.currentFloor, Dir: od.currentOrder.Dir, Value: 0}
								sendGlobal <- DataStore.Global_OrderData{Floor: od.currentFloor, Dir: od.currentOrder.Dir, HandlingIP: od.myIP, Clear: true}
							}
							state("OPEN")
							ordersChann <- order{od.currentOrder.Floor, od.currentOrder.Dir, true}
							} else {
								ordersChann <- order{od.currentOrder.Floor, od.currentOrder.Dir, true}
								ordersChann <- order{od.currentOrder.Floor, od.currentOrder.Dir, false}
							}

						if od.currentFloor == 1 {
							od.status = "UP"
						} else if od.currentFloor == od.N_FLOOR {
							od.status = "DOWN"
						}
						updateCurrentOrder <- true

					} else if od.currentOrder.Floor > od.lastFloor && od.currentFloor != od.N_FLOOR {
						state("UP")
						od.status = "UP"

					} else if od.currentOrder.Floor < od.lastFloor && od.currentOrder.Floor != 0 {
						state("DOWN")
						od.status = "DOWN"
					}

				}()
*/
			case extSig := <-extButtonChannel:
				go func() {
					var extOrder order
					if extSig%2 == 0 {
						extOrder = order{((extSig - (extSig % 2) - 30) / 10), "DOWN", false}
						driverInterface.SetButtonLamp("DOWN", ((extSig-(extSig%2)-30)/10)-1, 1) //TODO make nicer. handle floors better

					} else {
						extOrder = order{((extSig - (extSig % 2) - 30) / 10), "UP", false}
						driverInterface.SetButtonLamp("UP", ((extSig-(extSig%2)-30)/10)-1, 1) //TODO make nicer. handle floors better

					}
					setLights <- DataStore.ExtButtons_Message{Floor: extOrder.Floor, Dir: extOrder.Dir, Value: 1}
					ordersAcosted <- extOrder
				}()

			case acosting := <-ordersAcosted:
				go func() {
					min := DataStore.Order_Message{Floor: acosting.Floor, Dir: acosting.Dir, OriginIP: od.myIP, Cost: cost(od.orderList, od.afterOrders, od.lastFloor, od.status, acosting.Floor, acosting.Dir), What: "COST_REQ"}
					req := min
					if !contains(acosting, od.orderList) && !od.commDisabled {
						toAll <- req
						fmt.Println("MASSE CAPS")
						abort := time.After(50 * time.Millisecond)
					loop:
						for {
							select {
							case <-abort:
								break loop
							case got := <-costResponsInternal:
								if got.Cost < min.Cost {
									min.Cost = got.Cost
									min.OriginIP = got.OriginIP
								}
							}
						}


						sendGlobal <- DataStore.Global_OrderData{Floor: min.Floor, Dir: min.Dir, HandlingIP: min.OriginIP, Clear: false}

						if min.OriginIP == od.myIP {
							ordersChann <- order{min.Floor, min.Dir, false}

						} else {
						
							toOne <- DataStore.Order_Message{Floor: min.Floor, Dir: min.Dir, RecipientIP: min.OriginIP, What: "O_REQ"}

						}
					} else if od.commDisabled {
						ordersChann <- order{Floor: min.Floor, Dir: min.Dir, Clear: false}
					}
				}()

			case input := <-recieve:
				go func() {
					if input.What == "COST_REQ" {

						price := cost(od.orderList, od.afterOrders, od.lastFloor, od.status, input.Floor, input.Dir)
						toOne <- DataStore.Order_Message{Floor: input.Floor, Dir: input.Dir, RecipientIP: input.OriginIP, OriginIP: od.myIP, Cost: price, What: "COST_RES"}
					
					} else if input.What == "COST_RES" {

						costResponsInternal <- input

					} else if input.What == "O_REQ" {

						ordersChann <- order{input.Floor, input.Dir, false}

					}
				}()

			case lit := <-recvLights:
				go func() {

					driverInterface.SetButtonLamp(lit.Dir, lit.Floor-1, lit.Value)

				}()

			case ipDown := <-processGOL:
				go func() {
					if od.commDisabled{
						for _, i := range od.GOL[ipDown] {
							ordersChann <- order{Floor: i.Floor, Dir: i.Dir, Clear: false}
						}
					} else {
						for _, i := range od.GOL[ipDown] {
							ordersAcosted <- order{Floor: i.Floor, Dir: i.Dir, Clear: false}
						}
					}
				}()

			case updateGOL := <-recvGlobal:
				go func() {
					if !updateGOL.Clear {

						od.GOL[updateGOL.HandlingIP] = append(od.GOL[updateGOL.HandlingIP], DataStore.Global_OrderData{Floor: updateGOL.Floor, Dir: updateGOL.Dir, HandlingIP: updateGOL.HandlingIP})

					} else {

						od.GOL[updateGOL.HandlingIP] = removeGOL(DataStore.Global_OrderData{Floor: updateGOL.Floor, Dir: updateGOL.Dir}, od.GOL[updateGOL.HandlingIP])
					}
				}()

			}
		}
	}()

	if oldOrders{
	fmt.Println("YAY!")
		updateCurrentOrder <- true
	}
}
