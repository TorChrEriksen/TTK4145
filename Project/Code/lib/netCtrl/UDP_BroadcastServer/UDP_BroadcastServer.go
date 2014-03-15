package UDP_BroadcastServer

import (
    "net"
    "fmt"
    "./../../DataStore"
)

// TODO: Implement logger, so we dont have to return and check for errors.....

func runServer(udpAddr net.UDPAddr, ch chan DataStore.Broadcast_Message, packetSize int) {
    socket, err := net.ListenUDP("udp4", &udpAddr)

    if err != nil {
        fmt.Println("Error: ", err.Error()) // TODO
        return
    }

    for {
        data := make([]byte, packetSize)
        _, remoteAddr, err := socket.ReadFromUDP(data)

        if err != nil {
            fmt.Println("Error: ", err.Error()) // TODO
            continue
        }
        if remoteAddr.IP != nil {
            ch <- DataStore.Broadcast_Message{IP: fmt.Sprint(remoteAddr.IP), Message: "Received a broadcast message"}
        }
    }
    return
}

func Run(ch chan DataStore.Broadcast_Message, broadcastPort int, packetSize int) {
    ipv4_broadcast := net.IPv4(255, 255, 255, 255)
    udpAddr := net.UDPAddr{IP : ipv4_broadcast, Port: broadcastPort}

    go runServer(udpAddr, ch, packetSize)
}
