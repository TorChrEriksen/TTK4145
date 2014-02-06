// SocketServer project main.go
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func acceptConn(conn net.Conn, l log.Logger) {
	l.Println("Success: Connection accepted from ", conn.RemoteAddr())
	listenForData(conn, l)
}

func listenForData(conn net.Conn, l log.Logger) {
	l.Println("listenForData")
	for {
		//io.Copy(conn, conn)
		dataIn, err := bufio.NewReader(conn).ReadString('\x00')
		if err != nil {
			l.Println(err.Error())
		} else {
			l.Println("Data from client: ", dataIn)
		}
	}

	// Handle timeout?!
	/*
		err := conn.Close()
		if err != nil {
			l.Println(err.Error())
		} else {
			l.Println("Connection closed.")
			fmt.Println("Connection closed.")
		}*/
}

func handleData() {

}

func main() {

	f, err := os.OpenFile("logfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("error opening file: ", err.Error())
	}
	defer f.Close()

	l := log.New(f, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	listener, err := net.Listen("tcp", ":12345")
	if err != nil {
		l.Println(err.Error())
	}
	defer listener.Close()

	l.Println("Listening for new connections...")
	for {
		fmt.Println("Listening for new connections...")
		conn, err := listener.Accept()
		if err != nil {
			l.Println(err.Error())
		} else {
			go acceptConn(conn, *l)
		}
	}

}
