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

type Client struct {
    IP string
    Ticks int
}

// Need a type or way to get FT data
