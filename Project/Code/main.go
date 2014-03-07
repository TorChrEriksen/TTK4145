package main

import (
    "./lib/netCtrl"
)

func main() {

    // Declaring and setting up net controller
    netCtrl := netCtrl.NetController{}
    netCtrl.Create()
    netCtrl.Run()
}
