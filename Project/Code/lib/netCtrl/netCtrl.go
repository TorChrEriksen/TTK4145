package netCtrl

import (
    "./UDP_BroadcastServer"
    "./UDP_BroadcastClient"
    "./SocketServer"
    "./SocketClient"
    "./NetServices"
    "./ClientCtrl"
    "./../DataStore"
    "./../logger"
    "fmt"
    "time"
    "strings"
    "encoding/json"
//    "net"
//    "bytes"
)

type NetController struct {
    Identifier string
    TCPPort int
    UDPPort int
    BroadcastPort int
    PacketSize int
    DisableComm bool
    Timeout time.Duration
    al *logger.AppLogger
    localIP string // TODO: change this to net.IP and do byte compare
    hostList []string //TODO : we are doing string compare, do it with bytes instead in some way
    tcpClients []SocketClient.SocketClient
    udpClients []SocketClient.SocketClient
    clientList []ClientCtrl.ClientInfo
    broadcastChan chan DataStore.Broadcast_Message
    heartbeatChan chan DataStore.Heartbeat_Message
    orderChan chan []byte
    bcChan chan int

//    sendOrderChannel chan []byte
}

func (nc *NetController) Create(a *logger.AppLogger) {
    fileName := fmt.Sprint("log/NetController/NetController_", time.Now().Format(time.RFC3339), ".log")
    logSymLink := "log/NetController.log"

    nc.al = a
    nc.al.SetPackageLog(nc.Identifier, fileName, logSymLink)

    var intErr int
    nc.localIP, intErr = NetServices.FindLocalIP()
    if intErr == 1 {
        nc.DisableComm = false
        nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Local IP found: ", nc.localIP))
    } else {
        nc.DisableComm = true
        nc.al.Send_To_Log(nc.Identifier, logger.ERROR, "Error finding local IP, disabling net communication")
    }

    nc.hostList = make([]string, 10)
    nc.tcpClients = make([]SocketClient.SocketClient, 10)
    nc.udpClients = make([]SocketClient.SocketClient, 10)
    nc.clientList = make([]ClientCtrl.ClientInfo, 10)

    nc.broadcastChan = make(chan DataStore.Broadcast_Message)
    nc.heartbeatChan = make(chan DataStore.Heartbeat_Message)
    nc.bcChan = make(chan int)

    nc.orderChan = make(chan []byte)

//    nc.sendOrderChannel = make(chan []byte)
}

func (nc *NetController) connectTCP(tcpAddr string) {

    for _, tcpConnection := range nc.tcpClients {
        if tcpConnection.GetTCPConn() != nil { //TODO Verify that this works
            if strings.EqualFold(tcpConnection.GetTCPConn().LocalAddr().String(), tcpAddr) {
                result := fmt.Sprint("Already connected to that address: ", tcpAddr, " --> ", tcpConnection.GetTCPConn().LocalAddr().String())
                nc.al.Send_To_Log(nc.Identifier, logger.ERROR, result)
                return
            }
        }
    }

    tcpClient := SocketClient.SocketClient{Identifier: "TCP_SOCKETCLIENT"}
    tcpClient.Create(nc.al)
    tcpErr := tcpClient.ConnectTCP(tcpAddr) //TODO: FIX

    // Add tcp connection to tcp slice.
    if tcpErr == 1 {
        nc.tcpClients = append(nc.tcpClients, tcpClient) 
        result := fmt.Sprint("Added connection to TCP slice: ", tcpClient.GetTCPConn().LocalAddr().String())
        nc.al.Send_To_Log(nc.Identifier, logger.INFO, result)
    } else {
        nc.al.Send_To_Log(nc.Identifier, logger.ERROR, "Error connecting to TCP.")
    }


}

func (nc *NetController) connectUDP(udpAddr string) {

    for _, udpConnection := range nc.udpClients {
        if udpConnection.GetUDPConn() != nil { //TODO Verify that this works
            if strings.EqualFold(udpConnection.GetUDPConn().LocalAddr().String(), udpAddr) {
                result := fmt.Sprint("Already connected to that address: ", udpAddr, " --> ", udpConnection.GetUDPConn().LocalAddr().String())
                nc.al.Send_To_Log(nc.Identifier, logger.ERROR, result)
                return
            }
        }
    }

    udpClient := SocketClient.SocketClient{Identifier: "UDP_SOCKETCLIENT"}
    udpClient.Create(nc.al) // TODO: one struct for UDP, one for TCP
    udpErr := udpClient.ConnectUDP(udpAddr)

    // Add udp connection to udp slice.
    if udpErr == 1 {
        nc.udpClients = append(nc.udpClients, udpClient)
        result := fmt.Sprint("Added connection to UDP slice: ", udpClient.GetUDPConn().LocalAddr().String())
        nc.al.Send_To_Log(nc.Identifier, logger.INFO, result)
        udpClient.SendHeartbeat()
    } else {
        nc.al.Send_To_Log(nc.Identifier, logger.ERROR, "Error connecting to UDP.")
    }
}

func (nc *NetController) Run() {

    UDP_BroadcastServer.Run(nc.broadcastChan, nc.BroadcastPort, nc.PacketSize)
    UDP_BroadcastClient.Run(nc.bcChan, nc.BroadcastPort)
    SocketServer.Run(nc.orderChan, nc.heartbeatChan, nc.TCPPort, nc.PacketSize)
    go nc.validateConnections()

    go func() {
        for {
            select {

            case bClient := <-nc.bcChan :
                // TODO: when to stop broadcasting, never? :)
                nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Sent a heartbeat with ", bClient, " bytes."))

            // Received a broadcast, check if its a new elevator or old
            case broadcastMessage := <-nc.broadcastChan :
                //fmt.Println([]byte(nc.localIP))
                //fmt.Println([]byte(broadcastMessage.IP))

                go func() {
                    if strings.EqualFold(string(nc.localIP), string(broadcastMessage.IP)) {
//                    if bytes.Equal(nc.localIP, broadcastMessage.IP) {
                        nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Ignoring broadcast from local IP: ", broadcastMessage.IP))
                        return
                    }

                // Check if we are connected to this computer
                // Connect if not
                // Need more logic here
                // Or maybe its enough? What happens if the heartbeat is still running but we loose the TCP connection?

                    for _, host := range nc.hostList {
                        if strings.EqualFold(string(host), string(broadcastMessage.IP)) {
//                    if bytes.Equal(host, broadcastMessage.IP) {
                            nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Already part of host list: ", broadcastMessage.IP))
                            return
                        }
                    }
                    nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Appending to host list: ", broadcastMessage.IP))
                    nc.hostList = append(nc.hostList, broadcastMessage.IP)


                    nc.connectTCP(fmt.Sprint(broadcastMessage.IP, ":", nc.TCPPort)) //TODO fix?
                    nc.connectUDP(fmt.Sprint(broadcastMessage.IP, ":", nc.UDPPort)) //TODO fix?
                }()

            // Received a heartbeat
            case heartbeat := <-nc.heartbeatChan :
                go func() {
                    for _, client := range nc.clientList {
                        if strings.EqualFold(client.GetIP(), heartbeat.IP) { //TODO: fix!
//                     if bytes.Equal(client.IP, heartbeat.IP) {
                            nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Already part of UDP client list: ", heartbeat.IP))

                            // Reset timer for this connection
                            client.SetAlive()
                            return
                        }
                    }
                    nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Appending to client list: ", heartbeat.IP))
                    newClient := ClientCtrl.ClientInfo{}
                    newClient.Create(heartbeat.IP, nc.Timeout)
                    nc.clientList = append(nc.clientList, newClient)

                }()

            // Received an order
            case orderMsg := <-nc.orderChan :
                go func() {
                    convData, errInt := nc.unmarshal(orderMsg)
                    if errInt == -1 {
                        //nc.al.Send_To_Log(nc.Identifier, logger.ERROR, fmt.Sprint("Cannot read message, somrthing went wrong unmarshaling."))
                        return
                    }

                    fmt.Println("Message on orderChan: ", convData.Message)
                    nc.al.Send_To_Log(nc.Identifier, logger.INFO, fmt.Sprint("Message received from a client: ", convData.Message))

                }()
            }
        }
    }()
}

// Validate our connections, remove those that has timed out
// TODO: Can we use this to detect if we are without network comm?
func (nc *NetController) validateConnections() {

    // Check if we have net comm.
    // Like if the slice is empty or so?

    for n, client := range nc.clientList {

        // Client timed out, remove it from our list
        if client.GetStatus() {
        nc.al.Send_To_Log(nc.Identifier, logger.INFO, "Connection timed out")

            // Remove the connection from TCP client list
            for k, tcpClient := range nc.tcpClients {
                if tcpClient.GetTCPConn() != nil { //TODO Verify that this works
                    if strings.EqualFold(client.GetIP(), tcpClient.GetTCPConn().LocalAddr().String()) {

                        nc.al.Send_To_Log(nc.Identifier, logger.INFO, "Attempting to kill tcp connection")
                        tcpClient.KillTCPConnection()
                        nc.tcpClients = append(nc.tcpClients[:k], nc.tcpClients[k + 1])
                    }
                }
            }

            // Remove the connection from UDP client list
            for j, udpClient := range nc.udpClients {
                if udpClient.GetUDPConn() != nil { //TODO Verify that this works
                    if strings.EqualFold(client.GetIP(), udpClient.GetUDPConn().LocalAddr().String()) {

                        nc.al.Send_To_Log(nc.Identifier, logger.INFO, "Attempting to kill udp connection")
                        udpClient.KillUDPConnection()
                        nc.udpClients = append(nc.udpClients[:j], nc.udpClients[j + 1])
                    }
                }
            }

            nc.clientList = append(nc.clientList[:n], nc.clientList[n + 1])
        }
    }
}

// Parameter is not to be a string, but serialized data.
// TODO Waiting for structure.
// TODO Do we want to check the tcpClient list vs the clientList?
func (nc *NetController) SendData(data DataStore.Order_Message) {

    fmt.Println(nc.tcpClients)

    // Send to all hosts
    convData := nc.marshal(data)
    if convData != nil {
        for _, client := range nc.tcpClients {
            if client.GetTCPConn() != nil {
                fmt.Println("Sending data")
                client.SendData(convData)
            }
        }
        return
    }
    nc.al.Send_To_Log(nc.Identifier, logger.ERROR, fmt.Sprint("Error while sending data: *NetController.SendData()."))

    // Check if we are connected to the client
        // print error message
        // reconnect and resend message?

    // Send message
}

func (nc *NetController) SendData_SingleRecepient(data DataStore.Order_Message, elevID int) {

    // TODO: Send to single recepient

    // Use elevID to find IP of that client.

    // Check if we are connected to the client
        // print error message
        // reconnect and resend message?

    // Send message
}

// Serialize data to send
func (nc *NetController) marshal(data DataStore.Order_Message) []byte {
    convData, err := json.Marshal(data)
    if err != nil {
        nc.al.Send_To_Log(nc.Identifier, logger.ERROR, fmt.Sprint("Error while marshalling: ", err.Error()))
        return nil
    }
    return convData
}

// Unserialize data received
func (nc *NetController) unmarshal(data []byte) (DataStore.Order_Message, int) {
    convData := DataStore.Order_Message{}
    err := json.Unmarshal(data, &convData)
    if err != nil {
        nc.al.Send_To_Log(nc.Identifier, logger.ERROR, fmt.Sprint("Error while unmarshalling: ", err.Error()))
        return convData, -1
    }
    return convData, 1
}
