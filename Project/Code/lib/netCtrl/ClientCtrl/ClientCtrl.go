package ClientCtrl

import (
    "time"
    "fmt"
)

type ClientInfo struct {
    ip string
    timedOut bool
    timeoutChan chan bool
    timeout time.Duration
}

func (ci *ClientInfo) Create(ip string, timeout time.Duration) {
    ci.ip = ip
    ci.timeout = timeout
    ci.timeoutChan = make(chan bool)
    go ci.runTimer()
}

func (ci *ClientInfo) GetIP() string {
    return ci.ip
}
func (ci *ClientInfo) GetStatus() bool {
    return ci.timedOut
}

func (ci *ClientInfo) SetAlive() {
    ci.timeoutChan <- false
}

func (ci *ClientInfo) runTimer() {

    timer := time.NewTimer(ci.timeout)
    go func() {
        <-timer.C

        // Client timed out, set the flag
        // TODO: make this dynamic set of channels in netctrl instead of using the flag?
        fmt.Println("Client timed out: ", ci.ip)
        ci.timedOut = true
        ci.timeoutChan <- true
    }()

    for sig := range ci.timeoutChan {

        // Received confirmation on timeout from timer.
        if sig {
            break
        }

        timer.Reset(ci.timeout)
    }
}
