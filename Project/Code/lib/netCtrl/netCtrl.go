package netCtrl

import (
    "./UDP_BroadcastServer"
    "./UDP_BroadcastClient"
    "./SocketServer"
    "./SocketClient"
    "./NetServices"
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
    tcpClientList []string
    udpClientList []string
    bsChan chan string
    bcChan chan int
    sUDPServerChan chan string
    sTCPServerChan chan string
    sc SocketClient.SocketClient
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
    nc.tcpClientList = make([]string, 10)
    nc.udpClientList = make([]string, 10)

    nc.bsChan = make(chan string)
    nc.bcChan = make(chan int)

    nc.sUDPServerChan = make(chan string)
    nc.sTCPServerChan = make(chan string)


    nc.sc = SocketClient.SocketClient{Identifier: "SOCKETCLIENT"}
}

func (nc *NetController) Run() {

    UDP_BroadcastServer.Create(nc.bsChan)
    UDP_BroadcastClient.Create(nc.bcChan)
    SocketServer.Create(nc.sTCPServerChan, nc.sUDPServerChan)
    nc.sc.Create(nc.al)

//    go func() {
        for {
            select {
            case bClient := <-nc.bcChan :
                // TODO: when to stop broadcasting?
                nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Sent ", bClient, " bytes."))

            case bServer := <-nc.bsChan :
                go func() {
                    if strings.EqualFold(nc.localIP, bServer) {
                        nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Ignoring broadcast from local IP: ", bServer))
                        return
                    }

                    for _, host := range nc.hostList {
                        if strings.EqualFold(host, bServer) {
                            nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Already part of host list: ", bServer))
                            return
                        }
                    }
                    nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Appending to host list: ", bServer))
                    nc.hostList = append(nc.hostList, bServer)
                    nc.sc.ConnectTCP(bServer + ":12345")
                    nc.sc.ConnectUDP(bServer + ":12346")
                    nc.sc.SendHeartbeat()
                }()

                // Check if we are connected to this computer
                // Connect if not

            case ssUDP := <-nc.sUDPServerChan :
                //fmt.Println(ssUDP)
                // Update heartbeat, or not... Client should do that!?

                go func() {
                    for _, client := range nc.udpClientList {
                        if strings.EqualFold(client, ssUDP) {
                            nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Already part of UDP client list: ", ssUDP))
                            return
                        }
                    }
                    nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Appending to UDP client list: ", ssUDP))
                    nc.udpClientList = append(nc.udpClientList, ssUDP)

                }()

            case ssTCP := <-nc.sTCPServerChan :
                //fmt.Println(ssTCP)

                go func() {
                    for _, client := range nc.tcpClientList {
                        if strings.EqualFold(client, ssTCP) {
                            nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Already part of TCP client list: ", ssTCP))
                            return
                        }
                    }
                    nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Appending to TCP client list: ", ssTCP))
                    nc.tcpClientList = append(nc.tcpClientList, ssTCP)

                }()

                // Check if this is a first time connect.
                // Connect back if yes
            }
        }
//    }()
/*
    go func() {
        for {
            select {
            case :
            case :
            }
        }
    } 
    */

}

// Parameter is not to be a string, but serialized data.
// TODO Waiting for structure.
func(nc *NetController) SendData(a string) {
    nc.sc.Send(a)
}

