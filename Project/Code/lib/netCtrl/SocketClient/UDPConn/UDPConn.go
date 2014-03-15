package UDPConn

import (
	"fmt"
	"net"
    "time"
)

func InitComm(addr string) (int, *net.UDPAddr) {

	udpAddress, err := net.ResolveUDPAddr("udp", addr)

	if err != nil {
		fmt.Println("Error resolving UDP Address: ", err.Error())
		return -1, udpAddress
	}

	return 1, udpAddress
}

func OpenComm(addr net.UDPAddr) (int, *net.UDPConn) {

	conn, err := net.DialUDP("udp", nil, &addr)

	if err != nil {
		fmt.Println("Error(UDP): ", err.Error())
		return -1, conn
	}

	return 1, conn
}

func TerminateConn(conn net.UDPConn) int {
	err := conn.Close()
	if err != nil {
		fmt.Println("Error closing connection: (UDP)", err.Error())
		return -1
	} else {
		return 1
	}
}

func SendData(conn net.UDPConn, a string) int {
	//fmt.Println("SendData() (UDP)")
	data := make([]byte, 4096)
    data = []byte(a)

	n, err := conn.Write(data)
	if err != nil {
		fmt.Println("Error writing to connection: (UDP)", err.Error())
		return -1
	}
	return n
}

func SendHeartbeat(conn *net.UDPConn, a string, quit chan bool, ch chan string) {
    data := make([]byte, 4096)
    data = []byte(a)

    for {
        select {
        case <- quit :
            ch <- "I was told to quit"
            return
        default :
//            for _, c := range conn {
//                if c != nil {
            if conn != nil {
                _, err := conn.Write(data)
                if err != nil {
                    ch <- fmt.Sprint("Error writing to connection: (UDP)", err.Error())
                }
                ch <- "Sent a heartbeat"
            }
//            }
            time.Sleep(time.Second)
        }
    }
}


