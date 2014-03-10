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
)

type NetController struct {
    // What if this is uninitialized when calling SendData
    // eg. calling SendData before Create(), or something....
    Identifier string
    al *logger.AppLogger
    localIP string
    hostList []string
//    tcpClientList []string
    clientList []DataStore.Client
    broadcastChan chan DataStore.Broadcast_Message
    heartbeatChan chan DataStore.Heartbeat_Message
    orderChan chan DataStore.Order_Message

    sc SocketClient.SocketClient
    bcChan chan int
}

func (nc *NetController) Create(a *logger.AppLogger) {
    fileName := fmt.Sprint("log/NetController/NetController_", time.Now().Format(time.RFC3339), ".log")
    logSymLink := "log/NetController.log"

    nc.al = a
    nc.al.SetPackageLog(nc.Identifier, fileName, logSymLink)

    var intErr int
    nc.localIP, intErr = NetServices.FindLocalIP() 
    if intErr != 1 {
        nc.al.Send_To_Log(nc.Identifier, logger.ERROR, "Error finding local IP")
        // TODO: We will ahve to disable the net ctrl when we have no valid local IP
        // Enough for detecting that we have no connection?
        // Or do we use the heartbeat for that as well, and just ignore the local IP?
    } else {
        nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Local IP found: ", nc.localIP))
    }

    nc.hostList = make([]string, 10)
//    nc.tcpClientList = make([]string, 10)
    nc.clientList = make([]DataStore.Client, 10)

    nc.broadcastChan = make(chan DataStore.Broadcast_Message)
    nc.heartbeatChan = make(chan DataStore.Heartbeat_Message)
    nc.bcChan = make(chan int)

    nc.orderChan = make(chan DataStore.Order_Message)


    nc.sc = SocketClient.SocketClient{Identifier: "SOCKETCLIENT"}
}

func (nc *NetController) Run() {

    UDP_BroadcastServer.Create(nc.broadcastChan)
    UDP_BroadcastClient.Create(nc.bcChan)
    SocketServer.Create(nc.orderChan, nc.heartbeatChan)
    nc.sc.Create(nc.al)

//    go func() {
        for {
            select {
            case bClient := <-nc.bcChan :
                // TODO: when to stop broadcasting?
                nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Sent ", bClient, " bytes."))

            // Received a broadcast, check if its a new elevator or old
            case broadcastMessage := <-nc.broadcastChan :
                go func() {
                    if strings.EqualFold(nc.localIP, broadcastMessage.IP) {
                        nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Ignoring broadcast from local IP: ", broadcastMessage.IP))
                        return
                    }

                    for _, host := range nc.hostList {
                        if strings.EqualFold(host, broadcastMessage.IP) {
                            nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Already part of host list: ", broadcastMessage))
                            return
                        }
                    }
                    nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Appending to host list: ", broadcastMessage))
                    nc.hostList = append(nc.hostList, broadcastMessage.IP)
                    nc.sc.ConnectTCP(broadcastMessage.IP + ":12345")
                    nc.sc.ConnectUDP(broadcastMessage.IP + ":12346")
                    nc.sc.SendHeartbeat()
                }()

                // Check if we are connected to this computer
                // Connect if not

            case heartbeat := <-nc.heartbeatChan :
                //fmt.Println(ssUDP)
                // Update heartbeat, or not... Client should do that!?

                go func() {
                    for _, client := range nc.clientList {
                        if strings.EqualFold(client.IP, heartbeat.IP) {
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
                    nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Message received from a client: ", orderMsg.Message))

                }()
            }
        }
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
func (nc *NetController) SendData(a string) {
    nc.sc.Send(a)
}
