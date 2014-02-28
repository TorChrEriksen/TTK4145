package TCPConn

import (
	"bufio"
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
		fmt.Println("Error: ", err.Error())
		return -1, conn
	}

	return 1, conn

}

func TestComm(conn net.TCPConn) int {

	fmt.Fprintf(&conn, "Ni hao!\r\n\r\n")
	status, err := bufio.NewReader(&conn).ReadString('\n')

	if err != nil {
		fmt.Println("Error: ", err.Error())
		return -1
	}
	fmt.Println("Reply from server: ", status)

	err = conn.Close()
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return -1
	}

	fmt.Println("Connection closed!")
	return 1
}
