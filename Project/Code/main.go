package main

import (
    "./lib/netCtrl"
    "./lib/logger"
    "./lib/DataStore"
    "os"
    "os/signal"
    "fmt"
    "time"
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

    sendEggData(netCtrl)
}

func sendEggData(nc netCtrl.NetController) {
    dataForTheEgg := DataStore.Order_Message{Message : "(╯°□°）╯︵ ┻━┻)"}
    for {
        nc.SendData(dataForTheEgg)
        time.Sleep(time.Second * 1)
    }

}
