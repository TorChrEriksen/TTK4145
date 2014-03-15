package netCtrl

import (
    "./UDP_BroadcastServer"
    "./UDP_BroadcastClient"
    "./SocketServer"
    "./SocketClient"
    "./NetServices"
    "./../DataStore"
    "./../logger"
    "fmt"
    "time"
    "strings"
    "encoding/json"
//    "net"
//    "bytes"
)

type NetController struct {
    // What if this is uninitialized when calling SendData
    // eg. calling SendData before Create(), or something....
    Identifier string
    TCPPort int
    UDPPort int
    BroadcastPort int
    al *logger.AppLogger
    localIP string // TODO: change this to net.IP and do byte compare
    hostList []string //TODO : we are doing string compare, do it with bytes instead in some way
    tcpClients []SocketClient.SocketClient
    udpClients []SocketClient.SocketClient
    clientList []DataStore.Client
    broadcastChan chan DataStore.Broadcast_Message
    heartbeatChan chan DataStore.Heartbeat_Message
    orderChan chan []byte
    bcChan chan int

    sendOrderChannel chan []byte
}

func (nc *NetController) Create(a *logger.AppLogger) {
    fileName := fmt.Sprint("log/NetController/NetController_", time.Now().Format(time.RFC3339), ".log")
    logSymLink := "log/NetController.log"

    nc.al = a
    nc.al.SetPackageLog(nc.Identifier, fileName, logSymLink)

    var intErr int
    nc.localIP, intErr = NetServices.FindLocalIP()
    if intErr == 1 {
        nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Local IP found: ", nc.localIP))
    } else {
        nc.al.Send_To_Log(nc.Identifier, logger.ERROR, "Error finding local IP")
        // TODO: We will ahve to disable the net ctrl when we have no valid local IP
        // Enough for detecting that we have no connection?
        // Or do we use the heartbeat for that as well, and just ignore the local IP?

    }

    nc.hostList = make([]string, 10)
    nc.tcpClients = make([]SocketClient.SocketClient, 10)
    nc.udpClients = make([]SocketClient.SocketClient, 10)
    nc.clientList = make([]DataStore.Client, 10)

    nc.broadcastChan = make(chan DataStore.Broadcast_Message)
    nc.heartbeatChan = make(chan DataStore.Heartbeat_Message)
    nc.bcChan = make(chan int)

    nc.orderChan = make(chan []byte)

    nc.sendOrderChannel = make(chan []byte)
}

func (nc *NetController) connectTCP(tcpAddr string) {

    // Check if socket is already connected to tcpAddr
    // TODO: need to have a flag for connected or not.
    for _, tcpConnection := range nc.tcpClients {
        if tcpConnection.GetTCPConn() != nil { //TODO Verify that this works
            if strings.EqualFold(tcpConnection.GetTCPConn().LocalAddr().String(), tcpAddr) {
                result := fmt.Sprint("Already connected to that address: ", tcpAddr, " --> ", tcpConnection.GetTCPConn().LocalAddr().String())
                nc.al.Send_To_Log(nc.Identifier, logger.ERROR, result)
                return
            }
        }
    }

    tcpClient := SocketClient.SocketClient{Identifier: "TCP_SOCKETCLIENT"}
    tcpClient.Create(nc.al, nc.sendOrderChannel)
    tcpErr := tcpClient.ConnectTCP(tcpAddr) //TODO: FIX

    // Add tcp connection to tcp slice.
    if tcpErr == 1 {
        nc.tcpClients = append(nc.tcpClients, tcpClient) 
        result := fmt.Sprint("Added connection to TCP slice: ", tcpClient.GetTCPConn().LocalAddr().String())
        nc.al.Send_To_Log(nc.Identifier, logger.INFO, result)
    } else {
        nc.al.Send_To_Log(nc.Identifier, logger.ERROR, "Error connecting to TCP.")
    }


}

func (nc *NetController) connectUDP(udpAddr string) {

    // Check if socket is already connected to udpAddr
    // TODO: need to have a flag for connected or not.
    for _, udpConnection := range nc.udpClients {
        if udpConnection.GetUDPConn() != nil { //TODO Verify that this works
            if strings.EqualFold(udpConnection.GetUDPConn().LocalAddr().String(), udpAddr) {
                result := fmt.Sprint("Already connected to that address: ", udpAddr, " --> ", udpConnection.GetUDPConn().LocalAddr().String())
                nc.al.Send_To_Log(nc.Identifier, logger.ERROR, result)
                return
            }
        }
    }

    udpClient := SocketClient.SocketClient{Identifier: "UDP_SOCKETCLIENT"}
    udpClient.Create(nc.al, nc.sendOrderChannel) // TODO: one struct for UDP, one for TCP
    udpErr := udpClient.ConnectUDP(udpAddr)

    // Add udp connection to udp slice.
    if udpErr == 1 {
        nc.udpClients = append(nc.udpClients, udpClient)
        result := fmt.Sprint("Added connection to UDP slice: ", udpClient.GetUDPConn().LocalAddr().String())
        nc.al.Send_To_Log(nc.Identifier, logger.INFO, result)
        udpClient.SendHeartbeat()
    } else {
        nc.al.Send_To_Log(nc.Identifier, logger.ERROR, "Error connecting to UDP.")
    }
}

func (nc *NetController) Run() {

    UDP_BroadcastServer.Run(nc.broadcastChan, nc.BroadcastPort)
    UDP_BroadcastClient.Run(nc.bcChan, nc.BroadcastPort)
    SocketServer.Run(nc.orderChan, nc.heartbeatChan, nc.TCPPort)

    go func() {
        for {
            select {
            case bClient := <-nc.bcChan :
                // TODO: when to stop broadcasting?
                nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Sent ", bClient, " bytes."))

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
                }()

            case heartbeat := <-nc.heartbeatChan :
                //fmt.Println(ssUDP)
                // Update heartbeat, or not... Client should do that!?

                go func() {
                    for _, client := range nc.clientList {
                        if strings.EqualFold(client.IP, heartbeat.IP) { //TODO: fix!
//                     if bytes.Equal(client.IP, heartbeat.IP) {
                            nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Already part of UDP client list: ", heartbeat.IP))
                            return
                        }
                    }
                    nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Appending to UDP client list: ", heartbeat.IP))
                    nc.clientList = append(nc.clientList, DataStore.Client{IP : heartbeat.IP, Ticks : 0})

                }()

            case orderMsg := <-nc.orderChan :
                //fmt.Println(ssTCP)

                // We dont need a client list? Or?
                go func() {
/*
                    for _, client := range nc.tcpClientList {
                        if strings.EqualFold(client, ssTCP) {
                            nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Already part of TCP client list: ", ssTCP))
                            return
                        }
                    }
                    nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Appending to TCP client list: ", ssTCP))
                    nc.tcpClientList = append(nc.tcpClientList, ssTCP)
*/
                    convData, errInt := nc.unmarshal(orderMsg)
                    if errInt == -1 {
                        nc.al.Send_To_Log(nc.Identifier, logger.ERROR, fmt.Sprint("Cannot read message, somrthing went wrong unmarshaling."))
                        return
                    }

                    nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Message received from a client: ", convData.Message))

                }()
            }
        }
    }()
}

// TODO: Verify
// TODO: Need to sync the data! 
func (nc *NetController) validateConnections() {
    for _, client := range nc.clientList {
//        if client != nil { // TODO Verify that this works
            client.Ticks = client.Ticks + 1
//        }
    }
}

// Parameter is not to be a string, but serialized data.
// TODO Waiting for structure.
func (nc *NetController) SendData(data DataStore.Order_Message) {
    convData := nc.marshal(data)
    if convData != nil {
        nc.sendOrderChannel <- convData
        return
    }
    nc.al.Send_To_Log(nc.Identifier, logger.ERROR, fmt.Sprint("Error while sending data: *NetController.SendData()."))
}

func (nc *NetController) marshal(data DataStore.Order_Message) []byte {
    convData, err := json.Marshal(data)
    if err != nil {
        nc.al.Send_To_Log(nc.Identifier, logger.ERROR, fmt.Sprint("Error while marshalling: ", err.Error()))
        return nil
    }
    return convData
}

func (nc *NetController) unmarshal(data []byte) (DataStore.Order_Message, int) {
    convData := DataStore.Order_Message{}
    err := json.Unmarshal(data, &convData)
    if err != nil {
        nc.al.Send_To_Log(nc.Identifier, logger.ERROR, fmt.Sprint("Error while unmarshalling: ", err.Error()))
        return convData, -1
    }
    return convData, 1
}
