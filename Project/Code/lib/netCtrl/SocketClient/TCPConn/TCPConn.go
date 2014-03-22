package TCPConn

import (
	"fmt"
	"net"
)

func InitComm(addr string) (int, *net.TCPAddr) {

	tcpAddress, err := net.ResolveTCPAddr("tcp", addr)

	if err != nil {
		fmt.Println("Error resolving TCP Address: ", err.Error())
		return -1, tcpAddress
	}

	return 1, tcpAddress
}

func OpenComm(addr net.TCPAddr) (int, *net.TCPConn) {

	conn, err := net.DialTCP("tcp", nil, &addr)

	if err != nil {
        fmt.Println("Error (TCP): ", err.Error())
		return -1, conn
	}

	return 1, conn

}

func TerminateConn(conn net.TCPConn) error {
	err := conn.Close()
    return err
}

func SendData(conn *net.TCPConn, data []byte) int {
	//fmt.Println("SendData() (UDP)")
	//data := make([]byte, 4096)
    //data = []byte(a)

    fmt.Println("ni hao")

	n, err := conn.Write(data)
	if err != nil {
		fmt.Println("Error writing to connection: (TCP)", err.Error())
		return -1
	}
	return n
}
