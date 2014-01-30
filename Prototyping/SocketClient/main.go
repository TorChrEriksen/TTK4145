// SocketClient project main.go
package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "78.91.51.215:12345")

	if err != nil {
		fmt.Println("Error: ", err)
	}

	fmt.Fprintf(conn, "Ni hao!\r\n\r\n")
	status, err := bufio.NewReader(conn).ReadString('\n')

	if err != nil {
		fmt.Println("Error: ", err)
	} else {
		fmt.Println("Reply from server: ", status)
	}

	err = conn.Close()
	if err != nil {
		fmt.Println("Error: ", err)
	} else {
		fmt.Println("Connection closed!")
	}
}
