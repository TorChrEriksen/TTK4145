package netCtrl

import (
    "./UDP_BroadcastServer"
    "./UDP_BroadcastClient"
    "./SocketServer"
    "./SocketClient"
    "./NetServices"
    "./ClientCtrl"
    "./../DataStore"
    "./../logger"
    "fmt"
    "time"
    "strings"
    "strconv"
    "encoding/json"
    "net"
)

type NetController struct {
    Identifier string
    TCPPort int
    UDPPort int
    BroadcastPort int
    PacketSize int
    CommDisabled bool
    Timeout time.Duration
    al *logger.AppLogger
    localIP string
    hostList []string
    tcpClients []SocketClient.SocketClient
    udpClients []SocketClient.SocketClient
    clientList []ClientCtrl.ClientInfo
    broadcastChan chan DataStore.Broadcast_Message
    heartbeatChan chan DataStore.Heartbeat_Message
    orderChan chan []byte
    bcChan chan int
    timeoutChan chan string
    monitorConnectionsChan chan int
    iAmTheMaster bool
}

func (nc *NetController) Debug() {
    for {
        fmt.Println("TCP clients: ", nc.tcpClients)
        fmt.Println("UDP clients: ", nc.udpClients)
        fmt.Println("Client list: ", nc.clientList)
        fmt.Println("Host list: ", nc.hostList)
        time.Sleep(time.Second * 2)
    }
}

func (nc *NetController) Create(a *logger.AppLogger) string {
    fileName := fmt.Sprint("log/NetController/NetController_", time.Now().Format(time.RFC3339), ".log")
    logSymLink := "log/NetController.log"

    nc.al = a
    nc.al.SetPackageLog(nc.Identifier, fileName, logSymLink)

    var intErr int
    nc.localIP, intErr = NetServices.FindLocalIP()
    if intErr == 1 {
        nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Local IP found: ", nc.localIP))
    } else {
        nc.al.Send_To_Log(nc.Identifier, logger.ERROR, "Error finding local IP, disabling net communication")
    }

    nc.hostList = make([]string, 0)
    nc.tcpClients = make([]SocketClient.SocketClient, 0)
    nc.udpClients = make([]SocketClient.SocketClient, 0)
    nc.clientList = make([]ClientCtrl.ClientInfo, 0)

    nc.broadcastChan = make(chan DataStore.Broadcast_Message)
    nc.heartbeatChan = make(chan DataStore.Heartbeat_Message)
    nc.bcChan = make(chan int)

    nc.orderChan = make(chan []byte)
    nc.timeoutChan = make(chan string)

    nc.monitorConnectionsChan = make(chan int)

    nc.iAmTheMaster = false

    return nc.localIP
}

// Connect to TCP
func (nc *NetController) connectTCP(tcpAddr string) int {

    for _, tcpConnection := range nc.tcpClients {
        if tcpConnection.GetTCPConn() != nil {
            if strings.EqualFold(tcpConnection.GetTCPConn().RemoteAddr().String(), tcpAddr) {
                result := fmt.Sprint("Already connected to that address: ", tcpAddr, " --> ", tcpConnection.GetTCPConn().RemoteAddr().String())
                nc.al.Send_To_Log(nc.Identifier, logger.ERROR, result)
                return -1
            }
        }
    }

    tcpClient := SocketClient.SocketClient{Identifier: "TCP_SOCKETCLIENT"}
    tcpClient.Create(nc.al)
    tcpErr := tcpClient.ConnectTCP(tcpAddr)

    // Add tcp connection to tcp slice.
    if tcpErr == 1 {
        nc.tcpClients = append(nc.tcpClients, tcpClient) 
        result := fmt.Sprint("Added connection to TCP slice: ", tcpClient.GetTCPConn().RemoteAddr().String())
        nc.al.Send_To_Log(nc.Identifier, logger.INFO, result)
        nc.monitorConnectionsChan <- 1
    } else {
        nc.al.Send_To_Log(nc.Identifier, logger.ERROR, "Error connecting to TCP.")
    }

    return tcpErr
}

// Connect to UDP
func (nc *NetController) connectUDP(udpAddr string) int {

    for _, udpConnection := range nc.udpClients {
        if udpConnection.GetUDPConn() != nil {
            if strings.EqualFold(udpConnection.GetUDPConn().RemoteAddr().String(), udpAddr) {
                result := fmt.Sprint("Already connected to that address: ", udpAddr, " --> ", udpConnection.GetUDPConn().RemoteAddr().String())
                nc.al.Send_To_Log(nc.Identifier, logger.ERROR, result)
                return -1
            }
        }
    }

    udpClient := SocketClient.SocketClient{Identifier: "UDP_SOCKETCLIENT"}
    udpClient.Create(nc.al)
    udpErr := udpClient.ConnectUDP(udpAddr)

    // Add udp connection to udp slice.
    if udpErr == 1 {
        nc.udpClients = append(nc.udpClients, udpClient)
        result := fmt.Sprint("Added connection to UDP slice: ", udpClient.GetUDPConn().RemoteAddr().String())
        nc.al.Send_To_Log(nc.Identifier, logger.INFO, result)
        udpClient.SendHeartbeat()
    } else {
        nc.al.Send_To_Log(nc.Identifier, logger.ERROR, "Error connecting to UDP.")
    }

    return udpErr
}

func (nc *NetController) Run(notifyCommChan chan bool, orderCallbackChan chan DataStore.Order_Message, processGOLChan chan string, extButtonCallbackChan chan DataStore.ExtButtons_Message, globalOrderListCallbackChan chan DataStore.Global_OrderData, masterChan chan bool) {

    UDP_BroadcastServer.Run(nc.broadcastChan, nc.BroadcastPort, nc.PacketSize)
    UDP_BroadcastClient.Run(nc.bcChan, nc.BroadcastPort)
    SocketServer.Run(nc.orderChan, nc.heartbeatChan, nc.TCPPort, nc.PacketSize)
    go nc.validateConnections(nc.timeoutChan, nc.monitorConnectionsChan, processGOLChan)
    go nc.monitorCommStatus(nc.monitorConnectionsChan, notifyCommChan, masterChan)

    go func() {
        for {
            select {

            case bClient := <-nc.bcChan :
                nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Sent a broadcast message with ", bClient, " bytes."))

            // Received a broadcast, check if its a new elevator or old
            case broadcastMessage := <-nc.broadcastChan :
                go func() {
                    if strings.EqualFold(string(nc.localIP), string(broadcastMessage.IP)) {
                        nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Ignoring broadcast from local IP: ", broadcastMessage.IP))
                        return
                    }

                    for _, host := range nc.hostList {
                        if strings.EqualFold(string(host), string(broadcastMessage.IP)) {
                            nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Already part of host list: ", broadcastMessage.IP))
                            return
                        }
                    }
                    nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Appending to host list: ", broadcastMessage.IP))
                    nc.hostList = append(nc.hostList, broadcastMessage.IP)


                    nc.connectTCP(fmt.Sprint(broadcastMessage.IP, ":", nc.TCPPort)) 
                    nc.connectUDP(fmt.Sprint(broadcastMessage.IP, ":", nc.UDPPort)) 
                }()

            // Received a heartbeat
            case heartbeat := <-nc.heartbeatChan :
                go func() {
                    for _, client := range nc.clientList {
                        if strings.EqualFold(client.GetIP(), heartbeat.IP) { 
                            nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Already part of UDP client list: ", heartbeat.IP))

                            // Reset timer for this connection
                            client.SetAlive()
                            return
                        }
                    }

                    nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Appending to client list: ", heartbeat.IP))
                    newClient := ClientCtrl.ClientInfo{}
                    newClient.Create(heartbeat.IP, nc.Timeout)
                    go newClient.RunCtrl(nc.timeoutChan)
                    nc.clientList = append(nc.clientList, newClient)
                }()

            // Received data
            case orderMsg := <-nc.orderChan :
                go func() {
                    convData, errInt := nc.unmarshal(orderMsg)
                    if errInt == -1 {
                        nc.al.Send_To_Log(nc.Identifier, logger.ERROR, fmt.Sprint("Cannot read message, somrthing went wrong unmarshaling."))
                        return
                    }
                    nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Message received from a client"))

                    m := convData.(map[string]interface{})

                    id := int(m["MessageID"].(float64))

                    if id == 1 {
                        fmt.Println("Order message")
                        var result DataStore.Order_Message
                        for k, v := range m {
                            switch k {
                            case "MessageID" :
                                result.MessageID = int(v.(float64))
                            case "Floor" :
                                result.Floor = int(v.(float64))
                            case "Dir" :
                                result.Dir = v.(string)
                            case "RecipientIP" :
                                result.RecipientIP = v.(string)
                            case "OriginIP" :
                                result.OriginIP = v.(string)
                            case "Cost" :
                                result.Cost = v.(float64)
                            case "What" :
                                result.What = v.(string)
                            default :
                                fmt.Println("Error: ", k, " | ", v)
                            }
                        }

                        fmt.Println("Result: ", result)
                        orderCallbackChan <- result

                    } else if id == 2 {
                        fmt.Println("Lights message")
                        var result DataStore.ExtButtons_Message
                        for k,v := range m {
                            switch k {
                            case "MessageID" :
                                result.MessageID = int(v.(float64))
                            case "Floor" :
                                result.Floor = int(v.(float64))
                            case "Dir" :
                                result.Dir = v.(string)
                            case "Value" :
                                result.Value = int(v.(float64))
                            default :
                                fmt.Println("Error: ", k, " | ", v)
                            }
                        }

                        fmt.Println("Result: ", result)
                        extButtonCallbackChan <- result

                    } else if id == 3 {
                        fmt.Println("Global order queue message")
                        var result DataStore.Global_OrderData
                        for k,v := range m {
                            switch k {
                            case "MessageID" :
                                result.MessageID = int(v.(float64))
                            case "Floor" :
                                result.Floor = int(v.(float64))
                            case "Dir" :
                                result.Dir = v.(string)
                            case "HandlingIP" :
                                result.HandlingIP = v.(string)
                            case "Clear" :
                                result.Clear = v.(bool)
                            default :
                                fmt.Println("Error: ", k, " | ", v)
                            }
                        }

                        fmt.Println("Result: ", result)
                        globalOrderListCallbackChan <- result

                    } else {
                        fmt.Println("Unknown message received")
                        return
                    }
                }()
            }
        }
    }()
}

// Checks if we have any connections available, if not we are "offline"
func (nc *NetController) monitorCommStatus(ch chan int, notifyChan chan bool, masterChan chan bool) {

    numberOfConnections := 0

    for {
        value := <-ch

        if value == -1 {
            numberOfConnections -= 1
        } else if value == 1 {
            numberOfConnections += 1
        } else {
            continue
        }

        fmt.Println("Number of connections: ", numberOfConnections)

        if numberOfConnections > 0 {
            nc.CommDisabled = false
            notifyChan <- false
        } else {
            nc.CommDisabled = true
            notifyChan <- true
        }

        // See if we are the master
        me, err := strconv.Atoi(nc.localIP[strings.LastIndex(nc.localIP, ".") + 1:])
        if err != nil {
            fmt.Println("Error converting local IP last segment to int", err.Error())
        } else if numberOfConnections == 0 {
            nc.iAmTheMaster = false
            fmt.Println("im no longer ZE MASTAH! No connections :<<<")
            masterChan <- nc.iAmTheMaster
        } else {
             for _, candidate := range nc.hostList {
                if candidate != "" {
                    a, err := strconv.Atoi(candidate[strings.LastIndex(candidate, ".") + 1:])
                    if err != nil {
                        fmt.Println("Error converting host IP last segment to int", err.Error())
                    } else {
                        if me < a {
                            nc.iAmTheMaster = false
                            fmt.Println("im not longer ZE MASTAH :<<<<<<<<<<")
                            masterChan <- nc.iAmTheMaster
                            break;
                        } else {
                            nc.iAmTheMaster = true
                            fmt.Println("I AM ZE MASTAAAAAH!")
                        }
                    }
                }
            }

            // Notify that we are the master
            masterChan <- nc.iAmTheMaster
        }
    }
}

// Validate our connections, remove when we receive a timeout on the timeout channel
func (nc *NetController) validateConnections(timeoutChan chan string, monitorConnectionsChan chan int, processGOLChan chan string) {
    for {
        timedOutClient := <-timeoutChan
        processGOLChan <- timedOutClient

        for n, client := range nc.clientList {
            if strings.EqualFold(client.GetIP(), timedOutClient) {
                fmt.Println("Found a client in our list that has timed out: ", timedOutClient)

                fmt.Println("Client list: ", nc.clientList)

                // Grow the slice by one
                nc.clientList = append(nc.clientList, ClientCtrl.ClientInfo{})
                fmt.Println("Client list: ", nc.clientList)

                // Swap the element that timed out with the last element (always nil), and delete shrink slice
                nc.clientList = append(nc.clientList[:n], nc.clientList[n + 1])
                fmt.Println("Client list: ", nc.clientList)

                // Remove the connection from TCP client list
                go func() {
                    for k, tcpClient := range nc.tcpClients {
                        if tcpClient.GetTCPConn() != nil {

                            host, _, err := net.SplitHostPort(tcpClient.GetTCPConn().RemoteAddr().String())
                            if err != nil {
                                nc.al.Send_To_Log(nc.Identifier, logger.ERROR, fmt.Sprint("Error spliting TCP IP: ", err.Error()))
                                return
                            }

                            fmt.Println("TCP IP: ", host)

                            if strings.EqualFold(timedOutClient, host) {

                                // Kill TCP connection
                                nc.al.Send_To_Log(nc.Identifier, logger.INFO, "Attempting to kill tcp connection")
                                tcpClient.KillTCPConnection()
                                fmt.Println("TCP clients: ", nc.tcpClients)

                                // Grow the slice by one
                                nc.tcpClients = append(nc.tcpClients, SocketClient.SocketClient{})
                                fmt.Println("TCP clients: ", nc.tcpClients)

                                // Swap the element that timed out with the last element (always nil), and delete shrink slice
                                nc.tcpClients = append(nc.tcpClients[:k], nc.tcpClients[k + 1])
                                fmt.Println("TCP clients: ", nc.tcpClients)
                            }
                        }
                    }
                }()

                // Remove the connection from UDP client list
                go func() {
                    for j, udpClient := range nc.udpClients {
                        if udpClient.GetUDPConn() != nil { 
                            
                            host, _, err := net.SplitHostPort(udpClient.GetUDPConn().RemoteAddr().String())
                            if err != nil {
                                nc.al.Send_To_Log(nc.Identifier, logger.ERROR, fmt.Sprint("Error spliting UDP IP: ", err.Error()))
                                return
                            }

                            fmt.Println("UDP IP: ", host)

                            if strings.EqualFold(timedOutClient, host) {

                                // Kill UDP connection
                                nc.al.Send_To_Log(nc.Identifier, logger.INFO, "Attempting to kill udp connection")
                                udpClient.KillUDPConnection()
                                fmt.Println("UDP clients: ", nc.udpClients)

                                // Grow the slice by one
                                nc.udpClients = append(nc.udpClients, SocketClient.SocketClient{})
                                fmt.Println("UDP clients: ", nc.udpClients)

                                // Swap the element that timed out with the last element (always nil), and delete shrink slice
                                nc.udpClients = append(nc.udpClients[:j], nc.udpClients[j + 1])
                                fmt.Println("UDP clients: ", nc.udpClients)
                            }
                        }
                    }
                }()

                // Remove the connection from host list
                go func() {
                    for m, host := range nc.hostList {
                        if strings.EqualFold(timedOutClient, host) {

                            // Grow the slice by one
                            nc.hostList = append(nc.hostList, "")
                            fmt.Println("Host list: ", nc.hostList)

                            // Swap the element that timed out with the last element (always nil), and delete shrink slice
                            nc.hostList = append(nc.hostList[:m], nc.hostList[m + 1])
                            fmt.Println("Host list: ", nc.hostList)

                            // Update our connection monitor
                            monitorConnectionsChan <- -1
                        }
                    }
                }()
            }
        }
    }
}

func (nc *NetController) SendGlobalOrderList(data DataStore.Global_OrderData) {
    convData := nc.marshal(data)
    if convData != nil {
        for _, client := range nc.tcpClients {
            if client.GetTCPConn() != nil {
                client.SendData(convData)
            }
        }
        return
    }
    nc.al.Send_To_Log(nc.Identifier, logger.ERROR, fmt.Sprint("Error while sending data: *NetController.SendGlobalOrderList()."))
}

func (nc *NetController) SendLights(data DataStore.ExtButtons_Message) {
    convData := nc.marshal(data)
    if convData != nil {
        for _, client := range nc.tcpClients {
            if client.GetTCPConn() != nil {
                client.SendData(convData)
            }
        }
        return
    }
    nc.al.Send_To_Log(nc.Identifier, logger.ERROR, fmt.Sprint("Error while sending data: *NetController.SendLights()."))
}

// Send data to all available elevators
func (nc *NetController) SendData(data DataStore.Order_Message) {
    convData := nc.marshal(data)
    if convData != nil {
        for _, client := range nc.tcpClients {
            if client.GetTCPConn() != nil {
                client.SendData(convData)
            }
        }
        return
    }
    nc.al.Send_To_Log(nc.Identifier, logger.ERROR, fmt.Sprint("Error while sending data: *NetController.SendData()."))
}

// Sending data to single recipient
func (nc *NetController) SendData_SingleRecepient(data DataStore.Order_Message, destIP string) {

    convData := nc.marshal(data)
    if convData != nil {

        // See if we are connected to the client
        for _, tcpClient := range nc.tcpClients {
            if tcpClient.GetTCPConn() != nil {
                host, _, err := net.SplitHostPort(tcpClient.GetTCPConn().RemoteAddr().String())
                if err != nil {
                    nc.al.Send_To_Log(nc.Identifier, logger.ERROR, fmt.Sprint("Error spliting TCP IP: ", err.Error()))
                    return
                }

                if strings.EqualFold(destIP, host) {
                    tcpClient.SendData(convData)
                }
            }
        }
        return
    }
    nc.al.Send_To_Log(nc.Identifier, logger.ERROR, fmt.Sprint("Error while sending data: *NetController.SendData_SingleRecepient()."))
}

// Serialize data to send
func (nc *NetController) marshal(data interface{}) []byte {
    convData, err := json.Marshal(data)
    if err != nil {
        nc.al.Send_To_Log(nc.Identifier, logger.ERROR, fmt.Sprint("Error while marshalling: ", err.Error()))
        return nil
    }
    return convData
}

// Unserialize data received
func (nc *NetController) unmarshal(data []byte) (interface{}, int) {
    var convData interface{}
    //convData := DataStore.Order_Message{}
    err := json.Unmarshal(data, &convData)
    if err != nil {
        nc.al.Send_To_Log(nc.Identifier, logger.ERROR, fmt.Sprint("Error while unmarshalling: ", err.Error()))
        return convData, -1
    }

    return convData, 1
}

func (nc *NetController) Exit() {
    // Close TCP connections
    for _, tcpClient := range nc.tcpClients {
        if tcpClient.GetTCPConn() != nil {
            tcpClient.KillTCPConnection()
        }
    }

    // Close UDP connections
    for _, udpClient := range nc.udpClients {
        if udpClient.GetUDPConn() != nil {
            udpClient.KillUDPConnection()
        }
    }
}
