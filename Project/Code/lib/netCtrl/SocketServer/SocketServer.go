package SocketServer

import (
	"fmt"
	"log"
	"net"
	"os"
    "strings"
    "time"
    "./../../DataStore"
    "./../NetServices"
)

func convertData(data []byte, n int) string {
	convertedData := make([]byte, n)
	for i := 0; i < n; i++ {
		convertedData[i] = data[i]
	}

	return fmt.Sprint("Data from client: ", n, " --> ", (string)(convertedData))
}

func acceptConn(conn net.Conn, l log.Logger, ch chan []byte, packetSize int) {
	l.Println("Success: Connection accepted from ", conn.RemoteAddr())
	for {
		data := make([]byte, packetSize)
		n, err := conn.Read(data)

        if err != nil {
            fmt.Println("HURP A DURP A LURP") // TODO remove
            l.Println("Error while reading from connection: ", err.Error(), " I read ", n, " bytes.")
            l.Println("ALERT: Connection probably terminated???")
            return
        }

//        convData := convertData(data, n)
//        convData = fmt.Sprint("Number of bytes read: ", n, " | Data: ", convData)
//        l.Println(convData)
        fmt.Println("Received some data, sending it on channel")
        ch <- data[:n]
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

func startTCPServ(ch chan []byte, port int, packetSize int) {
    fileName := fmt.Sprint("log/SocketServer/TCP_Server_", time.Now().Format(time.RFC3339), ".log")
    logSymLink := "log/TCP_Server.log"

	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Error opening file: ", err.Error())
	}
	defer f.Close()

	l := log.New(f, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)

    os.Remove(logSymLink)
    err = os.Symlink(strings.TrimLeft(fileName, "log/"), logSymLink)
    if err != nil {
        l.Println("Error creating symlink: ", err.Error())
    }

    l.Println("========== New log ==========")

    listenPort := fmt.Sprint(":", port)
	listener, err := net.Listen("tcp", listenPort)
	if err != nil {
        l.Println("Error: ", err.Error())
	}
	defer listener.Close()

	for {
	    l.Println("Listening for new connections...")
		conn, err := listener.Accept()
		if err != nil {
			l.Println(err.Error())
		} else {
			l.Println("Firing goroutine for handling connection.")
			go acceptConn(conn, *l, ch, packetSize)
		}
	}

    close(ch)
}

func startUDPServ(ch chan DataStore.Heartbeat_Message, packetSize int) {
    fileName := fmt.Sprint("log/SocketServer/UDP_Server_", time.Now().Format(time.RFC3339), ".log")
    logSymLink := "log/UDP_Server.log"

	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("error opening file: ", err.Error())
	}
	defer f.Close()

	l := log.New(f, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)

    os.Remove(logSymLink)
    err = os.Symlink(strings.TrimLeft(fileName, "log/"), logSymLink)
    if err != nil {
        l.Println("Error creating symlink: ", err.Error())
    }

    l.Println("========== New log ==========")

    candidate, errIntUDP := NetServices.FindUDPCandidate()
    if errIntUDP == -1 {
        l.Println("Error: could not find any local IP address")
        return
    }

    addr, err := net.ResolveUDPAddr("udp", candidate)
    if err != nil {
        l.Println("Error: ", err.Error())
        return
    }

	listener, err := net.ListenUDP("udp", addr)
	if err != nil {
        l.Println("Error: ", err.Error())
	}
	defer listener.Close()

    buffer := make([]byte, packetSize)
	for {
        n, netInfo, err := listener.ReadFromUDP(buffer)
        if err != nil {
            l.Println("Error reading from UDP: ", err.Error())
        }

        l.Println("Received UDP data, converting.")

        convData := convertData(buffer, n)
        convData = fmt.Sprint("Number of bytes read: ", n, " | Data: ", convData)
        l.Println(convData)
        ch <- DataStore.Heartbeat_Message{IP: fmt.Sprint(netInfo.IP), Message: convData}

        l.Println("Converting seems successfull.")
	}

    close(ch)
}

func Run(tcpChan chan []byte, udpChan chan DataStore.Heartbeat_Message, tcpPort int, packetSize int) {
    go startTCPServ(tcpChan, tcpPort, packetSize)
    go startUDPServ(udpChan, packetSize)
}
