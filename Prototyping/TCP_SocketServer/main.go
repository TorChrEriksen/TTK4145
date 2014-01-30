// SocketServer project main.go
package main

import (
	"fmt"
	"io"
	"net"
)

func acceptConn(conn net.Conn) {
	fmt.Println("Success: Connection accepted!")
	io.Copy(conn, conn)
	conn.Close()
}

func main() {
	listener, err := net.Listen("tcp", ":12345")
	if err != nil {
		fmt.Println("Error creating listener")
	}
	defer listener.Close()

	for {
		fmt.Println("Listening for new connections...")
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting")
		} else {
			go acceptConn(conn)
		}
	}

}
