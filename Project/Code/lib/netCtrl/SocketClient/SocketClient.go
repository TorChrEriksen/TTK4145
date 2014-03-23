package SocketClient

import (
	"./TCPConn"
    "./UDPConn"
    "./../../logger"
	"net"
    "fmt"
    "time"
)

type SocketClient struct {
    Identifier string
    al *logger.AppLogger
    udpConn *net.UDPConn
    tcpConn *net.TCPConn
    heartbeatChan chan bool
    orderChan chan []byte
}

// Always called before any other function in this module
func (sc *SocketClient) Create(a *logger.AppLogger) {
    fileName := fmt.Sprint("log/SocketClient/SocketClient_", time.Now().Format(time.RFC3339), ".log")
    logSymLink := "log/SocketClient.log"

    sc.al = a
    sc.al.SetPackageLog(sc.Identifier, fileName, logSymLink)

    sc.heartbeatChan = make(chan bool)
    sc.orderChan = make(chan []byte)

    go sc.waitForInput()
}

// Connect to host
func (sc *SocketClient) ConnectTCP(tcpAddr string) int {

	_, tcpAddress := TCPConn.InitComm(tcpAddr)
    var tcpErr int
    tcpErr, sc.tcpConn = TCPConn.OpenComm(*tcpAddress)

	if tcpErr != 1 {
        sc.al.Send_To_Log(sc.Identifier, logger.INFO, fmt.Sprint("Error connecting (TCP)"))
        return -1
	} else {
        return 1
    }
}

// Connect to host
func (sc *SocketClient) ConnectUDP(udpAddr string) int {

    _, udpAddress := UDPConn.InitComm(udpAddr)
    var udpErr int
    udpErr, sc.udpConn = UDPConn.OpenComm(*udpAddress)

    if udpErr != 1 {
        sc.al.Send_To_Log(sc.Identifier, logger.ERROR, fmt.Sprint("Error connecting (UDP)"))
        return -1
    } else {
        return 1
    }
}

func (sc *SocketClient) KillTCPConnection() {
    err := TCPConn.TerminateConn(*sc.tcpConn)
	if err != nil {
        sc.al.Send_To_Log(sc.Identifier, logger.ERROR, fmt.Sprint("Error closing connection (TCP): ", err.Error()))
	}
}

func (sc *SocketClient) KillUDPConnection() {
    err := UDPConn.TerminateConn(*sc.udpConn)
	if err != nil {
        sc.al.Send_To_Log(sc.Identifier, logger.ERROR, fmt.Sprint("Error closing connection (UDP): ", err.Error()))
	}
}


func (sc *SocketClient) SendHeartbeat() {
    callback := make(chan string)
    go UDPConn.SendHeartbeat(sc.udpConn, "Im aliiiiiive!", sc.heartbeatChan, callback)
    go func() {
        for data := range callback {
            switch data {
            case "quit" :
                return
            default:
                sc.al.Send_To_Log(sc.Identifier, logger.INFO, fmt.Sprint(data))

            }
        }
    }()
}

func (sc *SocketClient) SendData(data []byte) {
    sc.orderChan <- data
}

func (sc *SocketClient) waitForInput() {
    for order := range sc.orderChan {
        if sc.tcpConn != nil {
            n := TCPConn.SendData(sc.tcpConn, order)
            _ = n
        }
    }
}

func (sc *SocketClient) GetTCPConn() *net.TCPConn {
    return sc.tcpConn
}

func (sc *SocketClient) GetUDPConn() *net.UDPConn {
    return sc.udpConn
}
