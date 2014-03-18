package main

import (
    "./lib/netCtrl"
    "./lib/logger"
    "./lib/DataStore"
    "os"
    "os/signal"
    "fmt"
    "time"
    "encoding/xml"
    "io"
    "path/filepath"
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

    // Import config
    config := importConfig("config/appConfig.xml")
    fmt.Println("Loaded application config: ", *config)

    // Fire up interrupt catcher|
    if config.CatchInterrupt {
        go catchKill(appLogger)
    }

    // Declaring and setting up net controller
    if !config.DebugMode {
        netCtrl := netCtrl.NetController{Identifier: "NETCONTROLLER",
                                         TCPPort: config.PortTCP,
                                         UDPPort: config.PortUDP,
                                         BroadcastPort: config.PortBroadcast,
                                         PacketSize: config.PacketSize}
        netCtrl.Create(&appLogger)
        netCtrl.Run()

        // Sending some test data
        sendEggData(netCtrl)
    }
}

func sendEggData(nc netCtrl.NetController) {
    dataForTheEgg := DataStore.Order_Message{Message : "(╯°□°）╯︵ ┻━┻)"}
    time.Sleep(time.Second * 10)
    for {
        nc.SendData(dataForTheEgg)
        time.Sleep(time.Second * 1)
    }
}

// Config declaration and import part

type ImportedConfig struct {
    CatchInterrupt bool
    Redundant bool
    PortTCP int
    PortUDP int
    PortBroadcast int
    DebugMode bool
    Floors int
    ButtonBaseInternal int
    ButtonBaseExternal int
    FloorNumberBase int
    StopButtonBase int
    PacketSize int
    ElevID int
}

type ConfigLine struct {
    XMLName xml.Name `xml:"config"`
    Key string `xml:"key,attr"`
    Value int `xml:"value,attr"`
}

type AppConfig struct {
    XMLName xml.Name `xml:"appcnf"`
    Conf []*ConfigLine `xml:"config"`
}

func readConf(reader io.Reader) ([]*ConfigLine, error){
    config := &AppConfig{}
    decoder := xml.NewDecoder(reader)

    err := decoder.Decode(config)
    if err != nil {
        return nil, err
    }

    return config.Conf, nil
}

func importConfig(filePath string) *ImportedConfig {
    var appConfig []*ConfigLine
    var file *os.File

    defer func() {
        if file != nil {
            file.Close()
        }
    }()

    // Build the location of the straps.xml file
    // filepath.Abs appends the file name to the default working directly
    configFilePath, err := filepath.Abs(filePath)

    if err != nil {
        panic(err.Error())
    }

    // Open the config xml file
    file, err = os.Open(configFilePath)

    if err != nil {
        panic(err.Error())
    }

    // Read the config file
    appConfig, err = readConf(file)

    if err != nil {
        panic(err.Error())
    }

    // TODO: Default config?

    impCnf := &ImportedConfig{}

    // Nasty conversion, check out xml.unmarshall and that stuff....
    for n, element := range appConfig {
        switch n {
        case 0:
            if element.Value == 0 {
                impCnf.CatchInterrupt = false
            } else {
                impCnf.CatchInterrupt = true
            }
        case 1:
            if element.Value == 0 {
                impCnf.Redundant = false
            } else {
                impCnf.Redundant = true
            }
        case 2:
            impCnf.PortTCP = element.Value
        case 3:
            impCnf.PortUDP = element.Value
        case 4:
            impCnf.PortBroadcast = element.Value
        case 5:
            if element.Value == 0 {
                impCnf.DebugMode = false
            } else {
                impCnf.DebugMode = true
            }
        case 6:
            impCnf.Floors = element.Value
        case 7:
            impCnf.ButtonBaseInternal = element.Value
        case 8:
            impCnf.ButtonBaseExternal = element.Value
        case 9:
            impCnf.FloorNumberBase = element.Value
        case 10:
            impCnf.StopButtonBase = element.Value
        case 11:
            impCnf.PacketSize = element.Value
        case 12:
            impCnf.ElevID = element.Value
        }
    }

    return impCnf
}

// end Config part
