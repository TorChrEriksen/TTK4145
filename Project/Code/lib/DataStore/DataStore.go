package DataStore

type Heartbeat_Message struct{
    IP string
    Message string
}

type Broadcast_Message struct {
    IP string
    Message string
}

// Order Message for the elevator logic
type Order_Message struct {
	Floor		int
	Dir 		string
	RecipientIP	string
	OriginIP	string
	Cost		float64
	What		string
}

// Data received by socket server
//type Received_OrderData struct {
//    originIP string
//    data []byte
//}

// Message describes the state of the external buttons
type ExtButtons_Message struct {
    UpButton [3]bool
    DownButton [3]bool
}

/*
type Client struct {
    IP string
    ID int
}
*/

// Need a type or way to get FT data
