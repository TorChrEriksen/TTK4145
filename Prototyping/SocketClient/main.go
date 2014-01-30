// SocketClient project main.go
package main

import (
	"fmt"
	"time"
	"os"
	"./src/TCPConn"
)

func tryConnect(addr string) {

	for {
		time.Sleep(1000 * time.Millisecond)

		result := TCPConn.TestComm(addr)

		if result == -1 {
			fmt.Println("Error connecting to host")
		} else {
			fmt.Println("Connection terminated correctly")
		}
	}

}

func main() {

	go tryConnect("129.241.187.153:12345") // Faulty connection
	go tryConnect("129.241.187.156:12345") // Correct connection

	fmt.Println("press 1 to quit:")

	for {
		var input int
		fmt.Scanf("%x", &input)

		if input == 1 {
			os.Exit(1)
		}
	}

}
