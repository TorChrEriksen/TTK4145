package UDP_BroadcastServer

import (
    "net"
//    "fmt"
    "./../../DataStore"
)

func runServer(udpAddr net.UDPAddr, ch chan DataStore.Broadcast_Message) {
    socket, err := net.ListenUDP("udp4", &udpAddr)

    if err != nil {
        //fmt.Println("Error: ", err.Error())
        return
    }

    for {
        data := make([]byte, 4096)
        _, remoteAddr, err := socket.ReadFromUDP(data)

        if err != nil {
            //fmt.Println("Error: ", err.Error())
            continue
        }
_ = remoteAddr
//        ch <- fmt.Sprint(remoteAddr.IP)
    }
    return
}

func Create(ch chan DataStore.Broadcast_Message) {
    ipv4_broadcast := net.IPv4(255, 255, 255, 255)
    udpAddr := net.UDPAddr{IP : ipv4_broadcast, Port: 12345}

    go runServer(udpAddr, ch)
}
