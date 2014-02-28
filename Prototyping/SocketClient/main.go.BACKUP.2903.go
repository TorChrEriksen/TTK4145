// SocketClient project main.go
package main

import (
	"fmt"
	"time"
	"os"
	"./src/TCPConn"
)

<<<<<<< HEAD
func main() {
	conn, err := net.Dial("tcp", "78.91.51.215:12345")
=======
func tryConnect(addr string, identifier string) {
>>>>>>> master

    fmt.Println("Running: ", identifier)

	result, tcpAddress := TCPConn.InitComm(addr)

	if result == -1 {
		fmt.Println("Error connecting to host")
	} else {
		result, conn := TCPConn.OpenComm(*tcpAddress)
		if result == -1 {
			fmt.Println(identifier, ": Error connecting to host")
		} else {
			for {
				time.Sleep(1000 * time.Millisecond)
				result := TCPConn.TestComm(*conn)
				if result == -1 {
					fmt.Println("Error connecting to host")
				} else {
					fmt.Println("Connect to host correctly")
				}
			}
		}
	}
}

func main() {

	go tryConnect("129.241.187.153:12345", "Connection_1") // Faulty connection
	go tryConnect("129.241.187.156:12345", "Connection_2") // Correct connection
//	go tryConnect("129.241.187.161:33546") // Correct connection

	fmt.Println("press 1 to quit:")

	for {
		var input int
		fmt.Scanf("%x", &input)

		if input == 1 {
			os.Exit(1)
		}
	}

}
