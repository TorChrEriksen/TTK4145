package netCtrl

import (
    "./UDP_BroadcastServer"
    "./UDP_BroadcastClient"
//    "./SocketServer"
//    "./SocketClient"
    "fmt"
)

type NetController struct {}

func (nc NetController) Create() {

    bsChan := make(chan string)
    UDP_BroadcastServer.Create(bsChan)

    bcChan := make(chan int)
    UDP_BroadcastClient.Create(bcChan)

    SocketServer.Create()
    SocketClient.Create()

    go func() {
        for {
            select {
            case bClient := <-bcChan :
                fmt.Println( "Sent ", bClient, " bytes.")
            case bServer := <-bsChan :
                fmt.Println(bServer)
            }
        }
    }()

    go func() {
        for {
            select {
            case :
            case :
            }
        }
    } 
}

func(nc NetController) SendData(a string) {}

