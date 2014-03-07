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
    ip, err := net.InterfaceAddrs()
    if err != nil {
        fmt.Println("Error Lookup: ", err.Error())
        return "", -1
    }

    for _, ipAddr := range ip {
        if strings.Contains(ipAddr.String(), "/24") {
            candidate := strings.TrimRight(ipAddr.String(), "/24")
            return candidate, 1
        }
    }

    return "", -1
}
