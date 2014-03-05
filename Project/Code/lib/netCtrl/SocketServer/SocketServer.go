package SocketServer

import (
	"fmt"
	"log"
	"net"
	"os"
    "./../NetServices"
)

func convertData(data []byte, n int) {
	convertedData := make([]byte, n)
	for i := 0; i < n; i++ {
		convertedData[i] = data[i]
	}

	fmt.Println("Data from client: ", n, " --> ", (string)(convertedData))
}

func acceptConn(conn net.Conn, l log.Logger) {
	l.Println("Success: Connection accepted from ", conn.RemoteAddr())
	fmt.Println("Success: Connection accepted from ", conn.RemoteAddr())
	for {
		data := make([]byte, 4096)
		n, err := conn.Read(data)

        if err != nil {
            fmt.Println("Error while reading from connection: ", err.Error())
            return
        }

        fmt.Println("Number of bytes read: ", n)
        convertData(data, n)

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

func listenForData(conn net.Conn, l log.Logger) {
	l.Println("listenForData")

}

func handleData() {

}

func startTCPServ(ch chan int) {
	f, err := os.OpenFile("TCPServer.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
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
			fmt.Println("Firing goroutine for handling connection.")
			go acceptConn(conn, *l)
		}
	}

    ch <- 1
}

func startUDPServ(ch chan int) {
	f, err := os.OpenFile("UDP_Server.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("error opening file: ", err.Error())
	}
	defer f.Close()

	l := log.New(f, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)
    _ = l

    candidate, errIntUDP := NetServices.FindUDPCandidate()
    if errIntUDP == -1 {
        fmt.Println("Error: could not find any local IP address")
        return
    }

    addr, err := net.ResolveUDPAddr("udp", candidate)
    if err != nil {
        fmt.Println("Error: ", err.Error())
        return
    }

	listener, err := net.ListenUDP("udp", addr)
	if err != nil {
		l.Println(err.Error())
        fmt.Println("Error: ", err.Error())
	}
	defer listener.Close()

    buffer := make([]byte, 4096)
	for {
        n, _, err := listener.ReadFromUDP(buffer)
        if err != nil {
            fmt.Println("Error reading from UDP: ", err.Error())
        }

        convertData(buffer, n)
	}
    ch <- 1
}

func Create() {

    // "join" threads
    ch := make(chan int)

    go startTCPServ(ch)
    go startUDPServ(ch)

    <-ch
}
