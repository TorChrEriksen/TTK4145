package main

import (
    "./lib/netCtrl"
    "./lib/logger"
)

func main() {

    // Declaring and setting up application logger
    appLogger := logger.AppLogger{}
    appLogger.Create()

    // Declaring and setting up net controller
    netCtrl := netCtrl.NetController{Identifier: "NETCONTROLLER"}
    netCtrl.Create(&appLogger)
    netCtrl.Run()
}
