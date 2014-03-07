package SocketClient

import (
	"./TCPConn"
    "./UDPConn"
    "./../../logger"
	"net"
    "fmt"
    "time"
    "strings"
)

type SocketClient struct {
    Identifier string
    al *logger.AppLogger
    udpConn []*net.UDPConn
    tcpConn []*net.TCPConn
    heartbeatChan chan bool
}

// Always called before any other function in this module
func (sc *SocketClient) Create(a *logger.AppLogger) {
    fileName := fmt.Sprint("log/SocketClient/SocketClient_", time.Now().Format(time.RFC3339), ".log")
    logSymLink := "log/SocketClient.log"

    sc.al = a
    sc.al.SetPackageLog(sc.Identifier, fileName, logSymLink)

    sc.udpConn = make([]*net.UDPConn, 10)
    sc.tcpConn = make([]*net.TCPConn, 10)
    sc.heartbeatChan = make(chan bool)
}

// Connect to host
// Returns -1 if the connection was not successfull, in that case retry to connect
func (sc *SocketClient) ConnectUDP(udpAddr string) int {

    // Check if socket is already connected to udpAddr
    for _, udpConnection := range sc.udpConn {
        if udpConnection != nil { //TODO Verify that this works
            if strings.EqualFold(udpConnection.LocalAddr().String(), udpAddr) {
                sc.al.Send_To_Log(sc.Identifier, logger.INFO,
                    fmt.Sprint("Already connected to that address: ", udpAddr, " --> ", udpConnection.LocalAddr().String()))
                return 1
            }
        }
    }

    _, udpAddress := UDPConn.InitComm(udpAddr)
    udpErr, udpConn := UDPConn.OpenComm(*udpAddress)

    // Add udp connection to udp slice.
    if udpErr == 1 {
        sc.udpConn = append(sc.udpConn, udpConn)
        sc.al.Send_To_Log(sc.Identifier, logger.INFO,
            fmt.Sprint("Added UDP connection to udpConn slice: ", udpConn.LocalAddr().String()))
    }

    if udpErr != 1 {
        sc.al.Send_To_Log(sc.Identifier, logger.ERROR, fmt.Sprint("Error connecting (UDP)"))
        return -1
    } else {
        return 1 // Everything ok.
    }
}

func (sc *SocketClient) ConnectTCP(tcpAddr string) int {

    // Check if socket is already connected to tcpAddr
    for _, tcpConnection := range sc.tcpConn {
        if tcpConnection != nil { //TODO Verify that this works
            if strings.EqualFold(tcpConnection.LocalAddr().String(), tcpAddr) {
                sc.al.Send_To_Log(sc.Identifier, logger.INFO,
                    fmt.Sprint("Already connected to that address: ", tcpAddr, " --> ", tcpConnection.LocalAddr().String()))
                return 1
            }
        }
    }

	_, tcpAddress := TCPConn.InitComm(tcpAddr)
	tcpErr, tcpConn  := TCPConn.OpenComm(*tcpAddress)

    // Add tcp connection to tcp slice.
    if tcpErr == 1 {
        sc.tcpConn = append(sc.tcpConn, tcpConn)
        sc.al.Send_To_Log(sc.Identifier, logger.INFO,
            fmt.Sprintln("Added TCP connection to tcpConn slice: ", tcpConn.LocalAddr().String()))
    }

	if tcpErr != 1 {
        sc.al.Send_To_Log(sc.Identifier, logger.INFO, fmt.Sprint("Error connecting (TCP)"))
        return -1
	} else {
        return 1
    }


    /*
		case 2:
			{
				TCPConn.TerminateConn(*tcpConn)
			}

		case 4:
			{
				//TCPConn.SendData(conn_1, "This is data from conn_1\x00")
                TCPConn.SendData(*tcpConn, "Here is something mongo!£@11!: ¤¤¤ %%% Ni Hao!! END-not-here-but-here")
				//TCPConn.SendData(conn_2, "This is data from conn_2\r\n\r\n")
			}


		}
	}
    */
}

func (sc *SocketClient) Send(a string) {
}

func (sc *SocketClient) SendHeartbeat() {
    sc.heartbeatChan <- true
    UDPConn.SendHeartbeat(sc.udpConn, "Im aliiiiiive!", sc.heartbeatChan)
}
