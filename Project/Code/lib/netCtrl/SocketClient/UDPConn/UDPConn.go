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

func TerminateConn(conn net.UDPConn) error {
	err := conn.Close()
    return err
}

func SendData(conn net.UDPConn, a string) int {
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
            if conn != nil {
                _, err := conn.Write(data)
                if err != nil {
                    ch <- fmt.Sprint("Error writing to connection: (UDP)", err.Error())
                }
                ch <- "Sent a heartbeat"
            }
            time.Sleep(time.Second)
        }
    }
}
