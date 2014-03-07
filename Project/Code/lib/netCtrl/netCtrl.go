package netCtrl

import (
    "./UDP_BroadcastServer"
    "./UDP_BroadcastClient"
    "./SocketServer"
    "./SocketClient"
    "./NetServices"
    "fmt"
    "time"
    "strings"
    "log"
    "os"
)

type NetController struct {
    // What if this is uninitialized when calling SendData
    // eg. calling SendData before Create(), or something....
    file *os.File
    log *log.Logger
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

func (nc *NetController) Create() {
    fileName := fmt.Sprint("log/NetController/NetController", time.Now().Format(time.RFC3339), ".log")
    logSymLink := "log/NetController.log"

	nc.file, _ = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
    nc.log = log.New(nc.file, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)

    os.Remove(logSymLink)
    err := os.Symlink(strings.TrimLeft(fileName, "log/"), logSymLink)
    if err != nil {
        nc.log.Println("Error creating symlink: ", err.Error())
    }

    nc.log.Println("========== New log ==========")

    var intErr int
    nc.localIP, intErr = NetServices.FindLocalIP() 
    if intErr != 1 {
        if nc.log != nil {
            nc.log.Println("Error finding local IP")
            // TODO: We will ahve to disable the net ctrl when we have no valid local IP
            // Enough for detecting that we have no connection?
            // Or do we use the heartbeat for that as well, and just ignore the local IP?
        }
    } else {
        if nc.log != nil {
            nc.log.Println("Local IP found: ", nc.localIP)
        }
    }

    nc.hostList = make([]string, 10)
    nc.tcpClientList = make([]string, 10)
    nc.udpClientList = make([]string, 10)

    nc.bsChan = make(chan string)
    nc.bcChan = make(chan int)

    nc.sUDPServerChan = make(chan string)
    nc.sTCPServerChan = make(chan string)


    nc.sc = SocketClient.SocketClient{}
}

func (nc *NetController) Run() {

    UDP_BroadcastServer.Create(nc.bsChan)
    UDP_BroadcastClient.Create(nc.bcChan)
    SocketServer.Create(nc.sTCPServerChan, nc.sUDPServerChan)
    nc.sc.Create()

//    go func() {
        for {
            select {
            case bClient := <-nc.bcChan :
                // TODO: when to stop broadcasting?
                if nc.log != nil {
                    nc.log.Println( "Sent ", bClient, " bytes.")
                }

            case bServer := <-nc.bsChan :
                go func() {
                    if strings.EqualFold(nc.localIP, bServer) {
                        if nc.log != nil {
                            nc.log.Println("Ignoring broadcast from local IP: ", bServer)
                        }
                        return
                    }

                    for _, host := range nc.hostList {
                        if strings.EqualFold(host, bServer) {
                            if nc.log != nil {
                                nc.log.Println("Already part of host list: ", bServer)
                            }
                            return
                        }
                    }
                    if nc.log != nil {
                        nc.log.Println("Appending to host list: ", bServer)
                    }
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
                            if nc.log != nil {
                                nc.log.Println("Already part of UDP client list: ", ssUDP)
                            }
                            return
                        }
                    }
                    if nc.log != nil {
                        nc.log.Println("Appending to UDP client list: ", ssUDP)
                    }
                    nc.udpClientList = append(nc.udpClientList, ssUDP)

                }()

            case ssTCP := <-nc.sTCPServerChan :
                //fmt.Println(ssTCP)

                go func() {
                    for _, client := range nc.tcpClientList {
                        if strings.EqualFold(client, ssTCP) {
                            if nc.log != nil {
                                nc.log.Println("Already part of TCP client list: ", ssTCP)
                            }
                            return
                        }
                    }
                    if nc.log != nil {
                        nc.log.Println("Appending to TCP client list: ", ssTCP)
                    }
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

