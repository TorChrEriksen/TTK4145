package main

import (
    "./lib/netCtrl"
    "./lib/logger"
    "os"
    "os/signal"
    "fmt"
)

func main() {

    // Declaring and setting up application logger
    appLogger := logger.AppLogger{}
    appLogger.Create()

    // Stopping Ctrl + C kill signal
    go func() {
        killChan := make(chan os.Signal, 1)
        signal.Notify(killChan, os.Interrupt)

        for signal := range killChan {
            appLogger.Send_To_Log("", logger.ERROR, fmt.Sprint("Catched a killsignal:, ", signal))
        }
    }()

    // Declaring and setting up net controller
    netCtrl := netCtrl.NetController{Identifier: "NETCONTROLLER"}
    netCtrl.Create(&appLogger)
    netCtrl.Run()
}
