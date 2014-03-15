package main

import (
    "fmt"
    "os"
    "os/signal"
)

func main(){
    ch := make(chan os.Signal, 1)
    signal.Notify(ch, os.Interrupt)

    for sig := range ch {
        fmt.Println("Signal received: ", sig)
    }
    return
}
