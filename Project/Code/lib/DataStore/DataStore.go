package DataStore

type Heartbeat_Message struct {
	IP      string
	Message string
}

type Broadcast_Message struct {
	IP      string
	Message string
}

// Order Message for the elevator logic
type Order_Message struct {
	Floor       int
	Dir         string
	RecipientIP string
	OriginIP    string
	Cost        float64
	What        string
}

// Global Order Data
type Global_OrderData struct {
    Floor int
    Dir string
    HandlingIP string
    Clear bool
}

// Global Order List
type Received_OrderData struct {
    OriginIP string
    OrderList []Order_Message
}

// Message describes the state of the external buttons
type ExtButtons_Message struct {
	Floor int
	Dir   string
	Value int
}
