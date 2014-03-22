package DataStore

type Heartbeat_Message struct{
    IP string
    Message string
}

type Broadcast_Message struct {
    IP string
    Message string
}

// TODO Not finished
type Order_Message struct {
    Message string
}

// Message describes the state of the external buttons
type ExtButtons_Message struct {
    UpButton [3]bool
    DownButton [3]bool
}

type Client struct {
    IP string
    ID int
}

// Need a type or way to get FT data
