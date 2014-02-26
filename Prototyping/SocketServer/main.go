// SocketServer project main.go
package main

import (
	"fmt"
	"log"
	"net"
	"os"
    "time"
    "strings"
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

//func startTCPServ(ch chan int) {
func startTCPServ() {
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

//    ch <- 1
}

// TODO
// Duplicate of SocketClient/src/NetServices/NetServices.go::FindCandidate()
func findCandidate() (string, int) {
    ip, err := net.InterfaceAddrs()
    if err != nil {
        fmt.Println("Error Lookup: ", err.Error())
        return "", -1
    }

    for _, ipAddr := range ip {
        if strings.Contains(ipAddr.String(), "/24") {
            candidate := strings.TrimRight(ipAddr.String(), "/24")
            candidate = candidate + ":12346"
            return candidate, 1
        }
    }

    return "", -1
}

//func startUDPServ(ch chan int) {
func startUDPServ() {
	f, err := os.OpenFile("UDP_Server.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("error opening file: ", err.Error())
	}
	defer f.Close()

	l := log.New(f, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)
    _ = l

    candidate, errInt := findCandidate()
    if errInt == -1 {
        fmt.Println("Error: could not find any local IP address")
        return
    }

//    candidate := "localhost:12346"

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
//    ch <- 1
}

func main() {

    // "join" threads
//    ch := make(chan int)

//    go startTCPServ(ch)
//    go startUDPServ(ch)

    go startTCPServ()
    go startUDPServ()

    time.Sleep(30 * time.Second)

//    <-ch
}
