package ClientCtrl

import (
    "time"
    "fmt"
)

type ClientInfo struct {
    ip string
    setAliveChan chan bool
    timeout time.Duration
}

func (ci *ClientInfo) Create(ip string, timeout time.Duration) {
    ci.ip = ip
    ci.timeout = timeout
    ci.setAliveChan = make(chan bool)
}

func (ci *ClientInfo) RunCtrl(timeoutChan chan string) {

    timer := time.NewTimer(ci.timeout)
    go func() {
        <-timer.C

        // Client timed out, set the flag
        // TODO: make this dynamic set of channels in netctrl instead of using the flag?
        fmt.Println("Client timed out: ", ci.ip)
        ci.setAliveChan <- false
    }()

    for sig := range ci.setAliveChan {

        // Received confirmation on timeout from timer.
        if !sig {
            timeoutChan <- ci.ip
            break
        }
        timer.Reset(ci.timeout)
    }
}

func (ci *ClientInfo) GetIP() string {
    return ci.ip
}

func (ci *ClientInfo) SetAlive() {
    ci.setAliveChan <- true
}
