package main

import (
    "fmt"
)

func sumUp(ch chan int) {
    for i := 0; i < 50; i++{
        ch <- i
    }
    close(ch)
}

func main() {
    ch := make(chan int)
    go sumUp(ch)

    for i := range ch {
        fmt.Println(i)
    }
}
