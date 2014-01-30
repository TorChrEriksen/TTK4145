package TCPConn

import (
	"bufio"
	"fmt"
	"net"
)

func TestComm(addr string) int {
	tcpAddress, err := net.ResolveTCPAddr("tcp", addr)

	if err != nil {
		fmt.Println("Error resolving TCP Address: ", err.Error())
		return -1
	} 

	conn, err := net.DialTCP("tcp", nil, tcpAddress)

	if err != nil {
		fmt.Println("Error: ", err.Error())
		return -1
	}

	fmt.Fprintf(conn, "Ni hao!\r\n\r\n")
	status, err := bufio.NewReader(conn).ReadString('\n')

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
