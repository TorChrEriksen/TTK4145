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

func TerminateConn(conn net.TCPConn) int {
	err := conn.Close()
	if err != nil {
		fmt.Println("Error closing connection: ", err.Error())
		return -1
	} else {
		return 1
	}
}

func SendData(conn net.TCPConn, a string) int {
	fmt.Println("Here")
	n, err := fmt.Fprintf(&conn, a)
	if err != nil {
		fmt.Println(err.Error())
		return -1
		/*
				status, err := bufio.NewReader(&conn).ReadString('\n')
				if err != nil {
					fmt.Println(status)
					return n
				} else {
					fmt.Println(err.Error())
					return -1
				}
			} else {
				fmt.Println(err.Error())
				return -1*/

	}
	return n
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
