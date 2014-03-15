package NetServices

import (
    "net"
    "fmt"
    "strings"
)

func FindTCPCandidate() (string, int) {
    ip, err := net.InterfaceAddrs()
    if err != nil {
        fmt.Println("Error Lookup: ", err.Error())
        return "", -1
    }

    for _, ipAddr := range ip {
        if strings.Contains(ipAddr.String(), "/24") {
            candidate := strings.TrimRight(ipAddr.String(), "/24")
            candidate = candidate + ":12345"
            return candidate, 1
        }
    }

    return "", -1
}

func FindUDPCandidate() (string, int) {
    ip, err := net.InterfaceAddrs()
    if err != nil {
        fmt.Println("Error Lookup: ", err.Error())
        return "", -1
    }

    for _, ipAddr := range ip {
        if strings.Contains(ipAddr.String(), "/24") {
            candidate := strings.TrimRight(ipAddr.String(), "/24")
            candidate = candidate + ":12346"
            return candidate, 1
        }
    }

    return "", -1
}

func FindLocalIP() (string, int) {
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        fmt.Println("Error Lookup: ", err.Error())
        return "", -1
    }

    //TODO : what if the address we need is not the first! O.o
    for _, ipAddr := range addrs {
        if ipnet, ok := ipAddr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            return ipnet.IP.String(), 1
        }
/*
        if strings.Contains(ipAddr.String(), "/23") { // TODO: fix how to get the actual local IP, dont want to do the /23 or /24....
            candidate := strings.TrimRight(ipAddr.String(), "/23")
            return candidate, 1 //TODO: fix this nasty conversion
        }
*/
    }

    return "", -1
}
