package main

import (
    "fmt"
    "time"
)

func receiveAndPrint(id string, ch chan int) {
    for s := range ch {
        fmt.Println(id, ": ", s)
    }
}

func main() {
    ch := make(chan int)
    go receiveAndPrint("Goroutine 1", ch)
    go receiveAndPrint("Goroutine 2", ch)

    printThis := 1

    for {
        ch <- printThis
        ch <- printThis
        printThis += 1
        time.Sleep(time.Second)
    }

}
