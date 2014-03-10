package main

import (
    "./lib/netCtrl"
    "./lib/logger"
    "os"
    "os/signal"
    "fmt"
)


// Stopping Ctrl + C kill signal
func catchKill(appLog logger.AppLogger) {
    killChan := make(chan os.Signal, 1)
    signal.Notify(killChan, os.Interrupt)

    for signal := range killChan {
        appLog.Send_To_Log("", logger.ERROR, fmt.Sprint("Catched a killsignal:, ", signal))
    }
}

func main() {

    // Declaring and setting up application logger
    appLogger := logger.AppLogger{}
    appLogger.Create()

    //go catchKill(appLogger)

    // Declaring and setting up net controller
    netCtrl := netCtrl.NetController{Identifier: "NETCONTROLLER"}
    netCtrl.Create(&appLogger)
    netCtrl.Run()
}
