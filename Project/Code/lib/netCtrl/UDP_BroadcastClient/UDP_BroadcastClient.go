package UDP_BroadcastClient

import (
    "net"
    "fmt"
    "time"
)

func runClient(udpAddr net.UDPAddr, ch chan int) {
    socket, err := net.DialUDP("udp4", nil, &udpAddr)
    if err != nil {
        //fmt.Println("Error_1: ", err.Error())
        return
    }

    data := []byte("Im new here, I cannot allow you to do that Dave.")

    for {
        n, err := socket.Write(data)
        if err != nil {
            fmt.Println("Error_2: ", err.Error())
        }
        //fmt.Println("Sent ", n, " bytes")
        ch <- n
        time.Sleep(time.Second * 5)
    }
    ch <- -1
}

func Run(ch chan int) {

    ipv4_broadcast := net.IPv4(255, 255, 255, 255)
    udpAddr := net.UDPAddr{IP : ipv4_broadcast, Port: 12345}

    go runClient(udpAddr, ch)
}
