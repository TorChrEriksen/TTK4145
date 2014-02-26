package main

import (
    "net"
    "fmt"
)

func runServer(udpAddr net.UDPAddr) {
    socket, err := net.ListenUDP("udp4", &udpAddr)

    if err != nil {
        fmt.Println("Error: ", err.Error())
        return
    }

    for {
        data := make([]byte, 4096)
        read, remoteAddr, err := socket.ReadFromUDP(data)

        if err != nil {
            fmt.Println("Error: ", err.Error())
            continue
        }

        fmt.Println("From: ", remoteAddr.IP, ":", remoteAddr.Port, " --> ",  read, string(data))
    }
    return
}

func main() {
    ipv4_broadcast := net.IPv4(255, 255, 255, 255)
    udpAddr := net.UDPAddr{IP : ipv4_broadcast, Port: 12345}

    runServer(udpAddr)
}
