// SocketServer project main.go
package main

import (
	"fmt"
	"io"
	"net"
	"log"
	"os"
)

func acceptConn(conn net.Conn, l log.Logger) {
	l.Println("Success: Connection accepted from ", conn.RemoteAddr())
	io.Copy(conn, conn)
	conn.Close()
}

func main() {

	f, err := os.OpenFile("logfile", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("error opening file: ", err.Error())
	}
	defer f.Close()

	l := log.New(f, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	listener, err := net.Listen("tcp", ":12345")
	if err != nil {
		l.Println("Error creating listener")
	}
	defer listener.Close()

	l.Println("Listening for new connections...")
	for {
		fmt.Println("Listening for new connections...")
		conn, err := listener.Accept()
		if err != nil {
			l.Println("Error accepting connection from")
		} else {
			go acceptConn(conn, *l)
		}
	}

}
