package SocketServer

import (
	"fmt"
	"log"
	"net"
	"os"
    "./../NetServices"
)

func convertData(data []byte, n int) string {
	convertedData := make([]byte, n)
	for i := 0; i < n; i++ {
		convertedData[i] = data[i]
	}

	return fmt.Sprint("Data from client: ", n, " --> ", (string)(convertedData))
}

func acceptConn(conn net.Conn, l log.Logger, ch chan string) {
	l.Println("Success: Connection accepted from ", conn.RemoteAddr())
	ch <- fmt.Sprint("Success: Connection accepted from ", conn.RemoteAddr())
	for {
		data := make([]byte, 4096)
		n, err := conn.Read(data)

        if err != nil {
            l.Println("Error while reading from connection: ", err.Error(), " I read ", n, " bytes.")
            l.Println("ALERT: Connection probably terminated???")
            return
        }

        convData := convertData(data, n)
        ch <- fmt.Sprint("Number of bytes read: ", n, " | Data: ", convData)
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

/* Durp remove?
func listenForData(conn net.Conn, l log.Logger) {
	l.Println("listenForData")

}

func handleData() {

}
*/

func startTCPServ(ch chan string) {
	f, err := os.OpenFile("TCP_Server.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		ch <- fmt.Sprint("Error opening file: ", err.Error())
	}
	defer f.Close()

	l := log.New(f, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	listener, err := net.Listen("tcp", ":12345")
	if err != nil {
        ch <- fmt.Sprint("Error: ", err.Error())
		l.Println(err.Error())
	}
	defer listener.Close()

	l.Println("Listening for new connections...")
	for {
		ch <- fmt.Sprint("Listening for new connections...")
		conn, err := listener.Accept()
		if err != nil {
			l.Println(err.Error())
		} else {
			ch <- fmt.Sprint("Firing goroutine for handling connection.")
			go acceptConn(conn, *l, ch)
		}
	}

    ch <- "-1"
}

func startUDPServ(ch chan string) {
	f, err := os.OpenFile("UDP_Server.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		ch <- fmt.Sprint("error opening file: ", err.Error())
	}
	defer f.Close()

	l := log.New(f, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)

    candidate, errIntUDP := NetServices.FindUDPCandidate()
    if errIntUDP == -1 {
        ch <- fmt.Sprint("Error: could not find any local IP address")
        l.Println("Error: could not find any local IP address")
        return
    }

    addr, err := net.ResolveUDPAddr("udp", candidate)
    if err != nil {
        ch <- fmt.Sprint("Error: ", err.Error())
        l.Println("Error: ", err.Error())
        return
    }

	listener, err := net.ListenUDP("udp", addr)
	if err != nil {
        ch <- fmt.Sprint("Error: ", err.Error())
        l.Println("Error: ", err.Error())
	}
	defer listener.Close()

    buffer := make([]byte, 4096)
	for {
        n, _, err := listener.ReadFromUDP(buffer)
        if err != nil {
            ch <- fmt.Sprint("Error reading from UDP: ", err.Error())
            l.Println("Error reading from UDP: ", err.Error())
        }

        l.Println("Received UDP data, converting.")

        convData := convertData(buffer, n)
        ch <- fmt.Sprint("Number of bytes read: ", n, " | Data: ", convData)

        l.Println("Converting seems successfull.")
	}

    ch <- "-1"

}

func Create(tcpChan chan string, udpChan chan string) {

    go startTCPServ(tcpChan)
    go startUDPServ(udpChan)
}
