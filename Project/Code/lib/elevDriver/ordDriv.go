package elevDriver

import (
	"./lib"
    "./../DataStore"
	"fmt"
	"math"
	"time"
	"sort"
    "os"
    "io/ioutil"
	"encoding/json"
)


type OrderDriver struct{
    N_FLOOR int
	currentFloor 	int
	lastFloor 		int
	currentOrder	order
	orderList 		[]order
	afterOrders 	[]order
	status			string	
    commDisabled   bool
}

type ByFloor []order

type order struct{
	Floor int
	Dir string
	Clear bool
}

func (od *OrderDriver) Create() {
	od.currentFloor = 0
	od.lastFloor = 0
	od.currentOrder = order{}	
	od.orderList = make([]order, 0)
	od.afterOrders = make([]order, 0)
	od.status = "IDLE"
    od.commDisabled = true
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
    // TODO
}

// TODO implement processGOL
func (od *OrderDriver) Run( toOne chan DataStore.Order_Message, toAll chan DataStore.Order_Message, recieve chan DataStore.Order_Message, commStatus chan bool, setLights chan DataStore.ExtButtons_Message, recvLights chan DataStore.ExtButtons_Message, sendGlobal chan DataStore.Global_OrderData, recvGlobal chan DataStore.Global_OrderData, processGOL chan string){
	driverInterface.Init()
	
	intButtonChannel 	:= make(chan int)
	extButtonChannel 	:= make(chan int)
	floorChannel 		:= make(chan int)
	stopChannel 		:= make(chan int)
	timeoutChannel 		:= make(chan int)

	ordersChann 		:= make(chan order)
	updateCurrentOrder 	:= make(chan bool)
	updatePos 			:= make(chan order)

    costResponsInternal := make(chan DataStore.Order_Message)

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
                case statusChanged := <-commStatus:
                    od.commDisabled = statusChanged
                    fmt.Println("Comm status changed to: ", statusChanged, " | ", od.commDisabled) //TODO remove
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
    go func() {
		for {
			select {

				case a:= <-updateCurrentOrder:
					go func() {
						a=a
						if (len(od.orderList)== 0) && (len(od.afterOrders) == 0){
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
						incommingInternal := order{Floor: buttonSignal-10, Dir: "INT"}
						driverInterface.SetButtonLamp(incommingInternal.Dir, incommingInternal.Floor-1, 1 ) //TODO make nicer. handle floors better
						ordersChann <- incommingInternal
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
							updateCurrentOrder <- true
						}else if new_order.Clear{
							od.orderList = remove(order{new_order.Floor, new_order.Dir, false},od.orderList)
///							fmt.Println("REMOVED: ",new_order, od.orderList)
//							fmt.Println(od.orderList)
							updateCurrentOrder <- true
						
						}else if !contains(new_order, od.orderList) && new_order.Dir != "NO" {
							if od.status == "UP" {
								if (new_order.Floor < od.currentFloor && od.currentFloor == od.lastFloor) || (new_order.Floor < od.lastFloor+1 && od.currentFloor == 0) || new_order.Dir == "DOWN" {
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
								if (new_order.Floor > od.currentFloor && od.currentFloor == od.lastFloor) || (new_order.Floor > od.lastFloor-1 && od.currentFloor == 0) || new_order.Dir=="UP" {
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
						if !new_stuff.Clear || new_stuff.Dir != "NO" {
							od.currentOrder = new_stuff
						}
						if new_stuff.Dir == "NO" {
							for od.currentFloor == 0 {
								time.Sleep(time.Millisecond*250)
							}
							od.status = "IDLE"
//							state("STOP")
						} else if len(od.orderList) == 0 && len(od.afterOrders) == 0 {
							state("IDLE")
						} else if od.currentOrder.Floor == od.currentFloor{
							state("STOP")
							driverInterface.SetButtonLamp(od.currentOrder.Dir, od.currentFloor - 1, 0)
							if od.currentOrder.Dir!="INT"{
								setLights <- DataStore.ExtButtons_Message{Floor: od.currentFloor - 1, Dir: od.currentOrder.Dir, Value: 0}
							}
							state("OPEN")	
//							fmt.Println("GOT TO FLOOR")
							if od.currentFloor == 1 {
								od.status="UP"
							} else if od.currentFloor == od.N_FLOOR {
								od.status="DOWN"
							}
							ordersChann <- order{od.currentOrder.Floor, od.currentOrder.Dir, true}
							updateCurrentOrder <- true
						
						}else if od.currentOrder.Floor > od.lastFloor{
							state("UP")
							od.status="UP"

						}else if (od.currentOrder.Floor < od.lastFloor && od.currentOrder.Floor != 0){
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
						if extSig % 2 == 0 {
							extOrder = order{((extSig - (extSig % 2) - 30) / 10),"DOWN", false}
							driverInterface.SetButtonLamp("DOWN", ((extSig - (extSig % 2) - 30) / 10)-1, 1 ) //TODO make nicer. handle floors better
							//fmt.Println("ExtOrder: ", extOrder)
						}else{
							extOrder = order{((extSig - (extSig % 2) - 30) / 10),"UP", false}
							driverInterface.SetButtonLamp("UP", ((extSig - (extSig % 2) - 30) / 10)-1, 1 ) //TODO make nicer. handle floors better
							//fmt.Println("ExtOrder: ", extOrder)
						}
						setLights <- DataStore.ExtButtons_Message{Floor: extOrder.Floor, Dir: extOrder.Dir, Value: 1}
//						fmt.Println(cost(od.orderList, od.afterOrders, od.lastFloor, od.status, extOrder.floor, extOrder.dir))
				//		ordersChann <- extOrder
					//}()
                        
                        // OriginIP is set in Application Control 
                        min := DataStore.Order_Message{Floor: extOrder.Floor, Dir: extOrder.Dir, RecipientIP: "", Cost: cost(od.orderList, od.afterOrders, od.lastFloor, od.status, extOrder.Floor, extOrder.Dir), What: "COST_REQ"}

                        if !contains(extOrder, od.orderList) && !od.commDisabled {

                            toAll <- min
                            go func(){
                                got := <-costResponsInternal
                                if got.Cost < min.Cost{
                                    min = got
                                }
                            }()

                            time.Sleep(time.Second * 2)

                            if min.OriginIP == "" {
                                ordersChann <- order{min.Floor, min.Dir, false}

                            } else {
                                min.What, min.RecipientIP, min.OriginIP = "O_REQ", min.OriginIP, min.RecipientIP
                                toOne<-min

                            }  
                        } else {
                            ordersChann <- order{min.Floor, min.Dir, false}
                        }
                    }()

				case input := <- recieve:
					go func(){
						if input.What == "COST_REQ" {
							toOne <-DataStore.Order_Message{Floor: input.Floor, Dir: input.Dir, RecipientIP: input.OriginIP, OriginIP: input.RecipientIP, Cost: cost(od.orderList, od.afterOrders, od.lastFloor, od.status, input.Floor, input.Dir), What:"COST_RES"}
						} else if input.What == "COST_RES" {
							costResponsInternal <- input
						} else if input.What == "O_REQ" {
							ordersChann <- order{input.Floor, input.Dir, false}
						}
				}()

				case setLight := <- setLights:
					go func(){
						driverInterface.SetButtonLamp(setLight.Dir, setLight.Floor-1, setLight.Value)
						}()
			}
		}
	}()
}
