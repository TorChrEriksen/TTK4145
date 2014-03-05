package netCtrl

import (
    "./UDP_BroadcastServer"
    "./UDP_BroadcastClient"
    "./SocketServer"
    "./SocketClient"
    "fmt"
)

type NetController struct {
    // What if this is uninitialized when calling SendData
    // eg. calling SendData before Create(), or something....
    tcpSocketClient SocketClient.SocketClient
}

func (nc NetController) Create() {

    bsChan := make(chan string)
    UDP_BroadcastServer.Create(bsChan)

    bcChan := make(chan int)
    UDP_BroadcastClient.Create(bcChan)

    sUDPServerChan := make(chan string)
    sTCPServerChan := make(chan string)

    SocketServer.Create(sUDPServerChan, sTCPServerChan)
//    SocketClient.Create()

//    go func() {
        for {
            select {
            case bClient := <-bcChan :
                fmt.Println( "Sent ", bClient, " bytes.")
            case bServer := <-bsChan :
                fmt.Println(bServer)
            case ssUDP := <-sUDPServerChan :
                fmt.Println(ssUDP)
            case ssTCP := <-sTCPServerChan :
                fmt.Println(ssTCP)
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
func(nc NetController) SendData(a string) {
    nc.tcpSocketClient.Send(a)
}

