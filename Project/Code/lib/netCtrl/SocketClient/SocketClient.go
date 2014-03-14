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
    udpConn *net.UDPConn // mod
    tcpConn *net.TCPConn // mod
    heartbeatChan chan bool
    orderChan chan []byte
}

// Always called before any other function in this module
func (sc *SocketClient) Create(a *logger.AppLogger, ch chan []byte) {
    fileName := fmt.Sprint("log/SocketClient/SocketClient_", time.Now().Format(time.RFC3339), ".log")
    logSymLink := "log/SocketClient.log"

    sc.al = a
    sc.al.SetPackageLog(sc.Identifier, fileName, logSymLink)

    sc.heartbeatChan = make(chan bool)
    sc.orderChan = ch

    go sc.waitForInput()
}

// Connect to host
// Returns -1 if the connection was not successfull, in that case retry to connect
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

// Connect to host
// Returns -1 if the connection was not successfull, in that case retry to connect
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


func (sc *SocketClient) SendHeartbeat() {
    // TODO: need to stop the heartbeat?
    // sc.heartbeatChan <- true
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

// TODO: Need to make the package SocketClient only one connection, and let the netCtrl control each one of them.
func (sc *SocketClient) waitForInput() {
    for order := range sc.orderChan {
//        for _, host := range sc.tcpConn {
//            if host != nil {
        if sc.tcpConn != nil {
            fmt.Println("ni hao")
                n := TCPConn.SendData(sc.tcpConn, order) // TODO: use return value for something?
                _ = n
            }
//        }
    }
}

func (sc *SocketClient) GetTCPConn() *net.TCPConn {
    return sc.tcpConn
}

func (sc *SocketClient) GetUDPConn() *net.UDPConn {
    return sc.udpConn
}
