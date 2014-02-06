// SocketClient project main.go
package main

import (
	"./src/TCPConn"
	"fmt"
	"net"
	"os"
)

func tryConnect(addr string, identifier string, c *net.TCPConn) {

	fmt.Println("Running: ", identifier)

	result, tcpAddress := TCPConn.InitComm(addr)

	if result != -1 {
		result, conn := TCPConn.OpenComm(*tcpAddress)
		if result != -1 {
			c = conn
		} else {
			c = nil
		}
	}
}

func main() {

	//go tryConnect("129.241.187.153:12345", "Connection_1") // Faulty connection
	//go tryConnect("129.241.187.156:12345", "Connection_2") // Correct connection
	//	go tryConnect("129.241.187.161:33546") // Correct connection
	var conn_1 net.TCPConn
	//var conn_2 net.TCPConn
	go tryConnect("localhost:12345", "Connection_1", &conn_1) // Correct connection
	//go tryConnect("localhost:12346", "Connection_2", &conn_2) // Correct connection

	fmt.Println("press 1 to quit:")

	for {
		var input int
		fmt.Scanf("%d", &input)

		switch input {
		case 0:
			{
				continue
			}
		case 1:
			{
				os.Exit(1)
			}
		case 2:
			{
				TCPConn.TerminateConn(conn_1)
			}
		case 3:
			{
				fmt.Println(input)
			}
		case 4:
			{
				var stringInput string
				fmt.Print("String to send to server: ")
				fmt.Scanf("%s", &stringInput)
				TCPConn.SendData(conn_1, "This is data from conn_1\x00")
				//TCPConn.SendData(conn_2, "This is data from conn_2\r\n\r\n")
			}
		default:
			{
				continue
			}
		}
	}
}
