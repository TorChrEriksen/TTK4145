package main

import (
    "net"
    "fmt"
)

func runClient(udpAddr net.UDPAddr) {
    socket, err := net.DialUDP("udp4", nil, &udpAddr)
    if err != nil {
        fmt.Println("Error: ", err.Error())
        return
    }

    data := []byte("I cannot allow you to do that Dave.")
    n, err := socket.WriteToUDP(data, &udpAddr)

    if err != nil {
        fmt.Println("Error: ", err.Error())
    }
    fmt.Println("Sent ", string(n), " bytes")
    return
}

func main() {
    ipv4_broadcast := net.IPv4(255, 255, 255, 255)
    udpAddr := net.UDPAddr{IP : ipv4_broadcast, Port: 12345}

    runClient(udpAddr)
}
