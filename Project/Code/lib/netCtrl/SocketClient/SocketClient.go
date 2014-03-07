package SocketClient

import (
	"./TCPConn"
    "./UDPConn"
	"net"
	"os"
    "log"
    "fmt"
    "time"
    "strings"
)

type SocketClient struct {
    file *os.File
    log *log.Logger
    udpConn []*net.UDPConn
    tcpConn []*net.TCPConn
    heartbeatChan chan bool
}

// Call before closing down Socket Client and or program
func (sc *SocketClient) Destory() {
	defer sc.file.Close()
}

// Always called before any other function in this module
func (sc *SocketClient) Create() {
    sc.file = nil
    sc.log = nil
    sc.udpConn = make([]*net.UDPConn, 10)
    sc.tcpConn = make([]*net.TCPConn, 10)
    sc.heartbeatChan = make(chan bool)

    fileName := fmt.Sprint("log/SocketClient/SocketClient_", time.Now().Format(time.RFC3339), ".log")
    logSymLink := "log/SocketClient.log"

	sc.file, _ = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
    sc.log = log.New(sc.file, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)

    os.Remove(logSymLink)
    err := os.Symlink(strings.TrimLeft(fileName, "log/"), logSymLink)
    if err != nil {
        sc.log.Println("Error creating symlink: ", err.Error())
    }

    if sc.log != nil {
        sc.log.Println("========== New log ==========")
    }
}

// Connect to host
// Returns -1 if the connection was not successfull, in that case retry to connect
func (sc *SocketClient) ConnectUDP(udpAddr string) int {

    // Check if socket is already connected to udpAddr
    for n, udpConnection := range sc.udpConn {
        if n < 1 { break }

        if strings.EqualFold(udpConnection.LocalAddr().String(), udpAddr) {
            if sc.log != nil {
                sc.log.Println("Already connected to that address: ", udpAddr, " --> ", udpConnection.LocalAddr().String())
            }
            return 1
        }
    }

    _, udpAddress := UDPConn.InitComm(udpAddr)
    udpErr, udpConn := UDPConn.OpenComm(*udpAddress)

    // Add udp connection to udp slice.
    if udpErr == 1 {
        sc.udpConn = append(sc.udpConn, udpConn)
        if sc.log != nil {
            sc.log.Println("Added UDP connection to udpConn slice: ", udpConn.LocalAddr().String())
        }
    }

    if udpErr != 1 {
        if sc.log != nil {
		    sc.log.Println("Error connecting (UDP)")
        }
        return -1
    } else {
        return 1 // Everything ok.
    }
}

func (sc *SocketClient) ConnectTCP(tcpAddr string) int {

    // Check if socket is already connected to tcpAddr
    for n, tcpConnection := range sc.tcpConn {
        if n < 1 { break }

        if strings.EqualFold(tcpConnection.LocalAddr().String(), tcpAddr) {
            if sc.log != nil {
                sc.log.Println("Already connected to that address: ", tcpAddr, " --> ", tcpConnection.LocalAddr().String())
            }
            return 1
        }
    }

	_, tcpAddress := TCPConn.InitComm(tcpAddr)
	tcpErr, tcpConn  := TCPConn.OpenComm(*tcpAddress)

    // Add tcp connection to tcp slice.
    if tcpErr == 1 {
        sc.tcpConn = append(sc.tcpConn, tcpConn)
        if sc.log != nil {
            sc.log.Println("Added TCP connection to tcpConn slice: ", tcpConn.LocalAddr().String())
        }
    }

	if tcpErr != 1 {
        if sc.log != nil {
		    sc.log.Println("Error connecting (TCP)")
        }
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
