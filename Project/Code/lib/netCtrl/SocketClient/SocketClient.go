package SocketClient

import (
	"./TCPConn"
    "./UDPConn"
    "./NetServices"
	"fmt"
	"net"
	"os"
)

func tryConnect(tcpAddr string, udpAddr string, identifier string) (*net.TCPConn, *net.UDPConn, int, int) {

	fmt.Println("Running: ", identifier)

	_, tcpAddress := TCPConn.InitComm(tcpAddr)
	tcpResult, tcpConn  := TCPConn.OpenComm(*tcpAddress)

    _, udpAddress := UDPConn.InitComm(udpAddr)
    udpResult, udpConn := UDPConn.OpenComm(*udpAddress)

	return tcpConn, udpConn, tcpResult, udpResult
}

func Create() {

	//go tryConnect("129.241.187.153:12345", "Connection_1") // Faulty connection
	//go tryConnect("129.241.187.156:12345", "Connection_2") // Correct connection
	//	go tryConnect("129.241.187.161:33546") // Correct connection
	//var conn_2 net.TCPConn
    localTCPAddress, errIntTCP := NetServices.FindTCPCandidate()
    localUDPAddress, errIntUDP := NetServices.FindUDPCandidate()
    if (errIntTCP == -1) ||  (errIntUDP == -1) {
        fmt.Println("Error finding candidates")
        os.Exit(1)
    }

    conn_1_TCP, conn_1_UDP, tcpErr, udpErr := tryConnect(localTCPAddress, localUDPAddress, "Connection_1") // Correct connection
	if tcpErr == -1 {
		fmt.Println("Error connecting (TCP)")
		os.Exit(1)
	}
    if udpErr == -1 {
		fmt.Println("Error connecting (UDP)")
		os.Exit(1)
    }
	//go tryConnect("localhost:12346", "Connection_2", &conn_2) // Correct connection

	fmt.Println("press 1 to quit:")

    //Debug: remove
    _ = conn_1_UDP

	for {
		var input int
		fmt.Scanf("%d", &input)

		switch input {
		case 0:
			{
				continue
			}
		case 1:
			{
				os.Exit(1)
			}
		case 2:
			{
				TCPConn.TerminateConn(*conn_1_TCP)
			}
		case 3:
			{
				fmt.Println(input)
			}
		case 4:
			{
				//TCPConn.SendData(conn_1, "This is data from conn_1\x00")
                TCPConn.SendData(*conn_1_TCP, "Here is something mongo!£@11!: ¤¤¤ %%% Ni Hao!! END-not-here-but-here")
				//TCPConn.SendData(conn_2, "This is data from conn_2\r\n\r\n")
			}
        case 5:
            {
                go UDPConn.SendHeartbeat(*conn_1_UDP, "Im aliiiiiiiiiiiive!!")
            }
		default:
			{
				continue
			}
		}
	}
}
