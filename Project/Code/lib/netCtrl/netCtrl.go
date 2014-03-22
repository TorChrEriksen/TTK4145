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
    "encoding/json"
    "net"
//    "bytes"
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
    localIP string // TODO: change this to net.IP and do byte compare
    hostList []string //TODO : we are doing string compare, do it with bytes instead in some way
    tcpClients []SocketClient.SocketClient
    udpClients []SocketClient.SocketClient
    clientList []ClientCtrl.ClientInfo
    broadcastChan chan DataStore.Broadcast_Message
    heartbeatChan chan DataStore.Heartbeat_Message
    orderChan chan []byte
    bcChan chan int
    timeoutChan chan string
    monitorConnectionsChan chan int

//    sendOrderChannel chan []byte
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

    return nc.localIP

//    nc.sendOrderChannel = make(chan []byte)
}

func (nc *NetController) connectTCP(tcpAddr string) int {

    for _, tcpConnection := range nc.tcpClients {
        if tcpConnection.GetTCPConn() != nil { //TODO Verify that this works
            if strings.EqualFold(tcpConnection.GetTCPConn().RemoteAddr().String(), tcpAddr) {
                result := fmt.Sprint("Already connected to that address: ", tcpAddr, " --> ", tcpConnection.GetTCPConn().RemoteAddr().String())
                nc.al.Send_To_Log(nc.Identifier, logger.ERROR, result)
                return -1
            }
        }
    }

    tcpClient := SocketClient.SocketClient{Identifier: "TCP_SOCKETCLIENT"}
    tcpClient.Create(nc.al)
    tcpErr := tcpClient.ConnectTCP(tcpAddr) //TODO: FIX

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

func (nc *NetController) connectUDP(udpAddr string) int {

    for _, udpConnection := range nc.udpClients {
        if udpConnection.GetUDPConn() != nil { //TODO Verify that this works
            if strings.EqualFold(udpConnection.GetUDPConn().RemoteAddr().String(), udpAddr) {
                result := fmt.Sprint("Already connected to that address: ", udpAddr, " --> ", udpConnection.GetUDPConn().RemoteAddr().String())
                nc.al.Send_To_Log(nc.Identifier, logger.ERROR, result)
                return -1
            }
        }
    }

    udpClient := SocketClient.SocketClient{Identifier: "UDP_SOCKETCLIENT"}
    udpClient.Create(nc.al) // TODO: one struct for UDP, one for TCP
    udpErr := udpClient.ConnectUDP(udpAddr)

    // Add udp connection to udp slice.
    if udpErr == 1 {
        nc.udpClients = append(nc.udpClients, udpClient)
        result := fmt.Sprint("Added connection to UDP slice: ", udpClient.GetUDPConn().RemoteAddr().String())
        nc.al.Send_To_Log(nc.Identifier, logger.INFO, result)
        udpClient.SendHeartbeat()
//        nc.monitorConnectionsChan <- 1
    } else {
        nc.al.Send_To_Log(nc.Identifier, logger.ERROR, "Error connecting to UDP.")
    }

    return udpErr
}

func (nc *NetController) Run(notifyCommChan chan bool, orderChanCallback chan DataStore.Order_Message) {

    UDP_BroadcastServer.Run(nc.broadcastChan, nc.BroadcastPort, nc.PacketSize)
    UDP_BroadcastClient.Run(nc.bcChan, nc.BroadcastPort)
    SocketServer.Run(nc.orderChan, nc.heartbeatChan, nc.TCPPort, nc.PacketSize)
    go nc.validateConnections(nc.timeoutChan, nc.monitorConnectionsChan)
    go nc.monitorCommStatus(nc.monitorConnectionsChan, notifyCommChan)

    go func() {
        for {
            select {

            case bClient := <-nc.bcChan :
                // TODO: when to stop broadcasting, never? :)
                nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Sent a heartbeat with ", bClient, " bytes."))

            // Received a broadcast, check if its a new elevator or old
            case broadcastMessage := <-nc.broadcastChan :
                //fmt.Println([]byte(nc.localIP))
                //fmt.Println([]byte(broadcastMessage.IP))

                go func() {
                    if strings.EqualFold(string(nc.localIP), string(broadcastMessage.IP)) {
//                    if bytes.Equal(nc.localIP, broadcastMessage.IP) {
                        nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Ignoring broadcast from local IP: ", broadcastMessage.IP))
                        return
                    }

                // Check if we are connected to this computer
                // Connect if not
                // Need more logic here
                // Or maybe its enough? What happens if the heartbeat is still running but we loose the TCP connection?

                    for _, host := range nc.hostList {
                        if strings.EqualFold(string(host), string(broadcastMessage.IP)) {
//                    if bytes.Equal(host, broadcastMessage.IP) {
                            nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Already part of host list: ", broadcastMessage.IP))
                            return
                        }
                    }
                    nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Appending to host list: ", broadcastMessage.IP))
                    nc.hostList = append(nc.hostList, broadcastMessage.IP)


                    nc.connectTCP(fmt.Sprint(broadcastMessage.IP, ":", nc.TCPPort)) //TODO fix?
                    nc.connectUDP(fmt.Sprint(broadcastMessage.IP, ":", nc.UDPPort)) //TODO fix?

                    /*
                    if (tcpErr == 1) && (udpErr == 1) {
                        newElevIdChan <- broadcastMessage.IP
                    } else {
                        fmt.Println("Error connecting")
                    }
                    */
                }()

            // Received a heartbeat
            case heartbeat := <-nc.heartbeatChan :
                go func() {
                    for _, client := range nc.clientList {
                        if strings.EqualFold(client.GetIP(), heartbeat.IP) { //TODO: fix!
//                     if bytes.Equal(client.IP, heartbeat.IP) {
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

            // Received an order
            case orderMsg := <-nc.orderChan :
                go func() {
                    convData, errInt := nc.unmarshal(orderMsg)
                    if errInt == -1 {
                        //nc.al.Send_To_Log(nc.Identifier, logger.ERROR, fmt.Sprint("Cannot read message, somrthing went wrong unmarshaling."))
                        return
                    }

                    nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Message received from a client"))
                    orderChanCallback <- convData

                }()
            }
        }
    }()
}

// Checks if we have any connections available, if not we are "offline"
func (nc *NetController) monitorCommStatus(ch chan int, notifyChan chan bool) {

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
    }
}

// Validate our connections, remove those that has timed out
// TODO: Can we use this to detect if we are without network comm?
func (nc *NetController) validateConnections(timeoutChan chan string, monitorConnectionsChan chan int) {

    // Check if we have net comm.
    // Like if the slice is empty or so?
    // TODO Remove from host list

    for {
        timedOutClient := <-timeoutChan

        for n, client := range nc.clientList {
            if strings.EqualFold(client.GetIP(), timedOutClient) {
                fmt.Println("Found a client in our list that has timed out: ", timedOutClient)
                
                // Update our connection monitor
                monitorConnectionsChan <- -1

                fmt.Println("Client list: ", nc.clientList)

                // Grow the slice by one
                nc.clientList = append(nc.clientList, ClientCtrl.ClientInfo{})
                fmt.Println("Client list: ", nc.clientList)

                // Swap the element that timed out with the last element (always nil), and delete shrink slice
                nc.clientList = append(nc.clientList[:n], nc.clientList[n + 1])
                fmt.Println("Client list: ", nc.clientList)

                // Shrink the slice by one
//                nc.clientList = nc.clientList[0:len(nc.clientList) - 1]
//                fmt.Println("Client list: ", nc.clientList)

                // Remove the connection from TCP client list
                for k, tcpClient := range nc.tcpClients {
                    if tcpClient.GetTCPConn() != nil { //TODO Verify that this works

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

                            // Shrink the slice by one
//                            nc.tcpClients = nc.tcpClients[0:len(nc.tcpClients) - 1]
//                            fmt.Println("TCP clients: ", nc.tcpClients)
                        }
                    }
                }

                // Remove the connection from UDP client list
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

                            // Shrink the slice by one
//                            nc.udpClients = nc.udpClients[0:len(nc.udpClients) - 1]
//                            fmt.Println("UDP clients: ", nc.udpClients)
                        }
                    }
                }
            }
        }
    }
    /*
    for {
        fmt.Println("Validating connections")
        for _, client := range nc.clientList {
 //           clientsAvailable = true
            fmt.Println("inside a client")

            // Client timed out, remove it from our list
            fmt.Println(client.GetIP(), " : ", client.GetStatus())
            if client.GetStatus() {
            nc.al.Send_To_Log(nc.Identifier, logger.INFO, "Connection timed out")

                // Remove the connection from TCP client list
                for _, tcpClient := range nc.tcpClients {
                    if tcpClient.GetTCPConn() != nil { //TODO Verify that this works
                        if strings.EqualFold(client.GetIP(), tcpClient.GetTCPConn().LocalAddr().String()) {

                            fmt.Println("Shrinking TCP conn slice")
                            nc.al.Send_To_Log(nc.Identifier, logger.INFO, "Attempting to kill tcp connection")
                            tcpClient.KillTCPConnection()
                            //nc.tcpClients = append(nc.tcpClients[:k], nc.tcpClients[k+1])
                            nc.tcpClients = nc.tcpClients[0:len(nc.tcpClients) - 1]
                            //nc.tcpClients = append([]SocketClient.SocketClient, nc.tcpClients[:len(nc.tcpClients) - 1])
                        }
                    }
                }

                // Remove the connection from UDP client list
                for _, udpClient := range nc.udpClients {
                    if udpClient.GetUDPConn() != nil { //TODO Verify that this works
                        if strings.EqualFold(client.GetIP(), udpClient.GetUDPConn().LocalAddr().String()) {

                            fmt.Println("Shrinking UDP conn slice")
                            nc.al.Send_To_Log(nc.Identifier, logger.INFO, "Attempting to kill udp connection")
                            udpClient.KillUDPConnection()
                            //nc.udpClients = append(nc.udpClients[:j], nc.udpClients[j + 1])
                            nc.udpClients = nc.udpClients[0:len(nc.udpClients) -1]
                            //nc.udpClients = append([]SocketClient.SocketClient(nil), nc.udpClients[:len(nc.udpClients) - 1])
                        }
                    }
                }

                fmt.Println("Shrinking clients slice")
                //nc.clientList = append(nc.clientList[:n], nc.clientList[n + 1])
                nc.clientList = nc.clientList[0:len(nc.clientList) - 1]
                //nc.clientList = append([]ClientCtrl.ClientInfo, nc.clientList[:len(nc.clientList) - 1])
            }
        }

        time.Sleep(time.Second * 1)
    }
    */
}

// TODO Waiting for structure.
// TODO Do we want to check the tcpClient list vs the clientList?
func (nc *NetController) SendData(data DataStore.Order_Message) {

    fmt.Println(nc.tcpClients)

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

func (nc *NetController) SendData_SingleRecepient(data DataStore.Order_Message, destIP string) {

    // TODO: Send to single recepient

    convData := nc.marshal(data)
    if convData != nil {

        // See if we are connected to the client
        for _, tcpClient := range nc.tcpClients {
            host, _, err := net.SplitHostPort(tcpClient.GetTCPConn().RemoteAddr().String())
            if err != nil {
                nc.al.Send_To_Log(nc.Identifier, logger.ERROR, fmt.Sprint("Error spliting TCP IP: ", err.Error()))
                return
            }

            if strings.EqualFold(destIP, host) {
                tcpClient.SendData(convData)
            }
        }
        return
    }

    nc.al.Send_To_Log(nc.Identifier, logger.ERROR, fmt.Sprint("Error while sending data: *NetController.SendData_SingleRecepient()."))

    // TODO Reconnect and resend message?
}

// Serialize data to send
func (nc *NetController) marshal(data DataStore.Order_Message) []byte {
    convData, err := json.Marshal(data)
    if err != nil {
        nc.al.Send_To_Log(nc.Identifier, logger.ERROR, fmt.Sprint("Error while marshalling: ", err.Error()))
        return nil
    }
    return convData
}

// Unserialize data received
func (nc *NetController) unmarshal(data []byte) (DataStore.Order_Message, int) {
    convData := DataStore.Order_Message{}
    err := json.Unmarshal(data, &convData)
    if err != nil {
        nc.al.Send_To_Log(nc.Identifier, logger.ERROR, fmt.Sprint("Error while unmarshalling: ", err.Error()))
        return convData, -1
    }
    return convData, 1
}

func (nc *NetController) Exit() {
    // Save stuff to file for the last time?

    // Close connections
    for _, tcpClient := range nc.tcpClients {
        if tcpClient.GetTCPConn() != nil { // TODO test this by letting client be timed out for a long time
            tcpClient.KillTCPConnection()
        }
    }

    for _, udpClient := range nc.udpClients {
        if udpClient.GetUDPConn() != nil { // TODO test this by letting client be timed out for a long time
            udpClient.KillUDPConnection()
        }
    }
}
