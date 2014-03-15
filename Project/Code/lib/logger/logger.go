package logger

import (
    "fmt"
    "log"
    "time"
    "os"
    "strings"
)

const (
    INFO = 1
    ERROR = 2
    FAILURE = 3
)

type AppLogger struct {
    maps map[string]*log.Logger
    files []*os.File
}

// Close files
// TODO: catch Ctrl + C or that the app is killed somewhere
func (lg *AppLogger) Destroy() {
    for _, f := range lg.files {
        if f != nil { //TODO verify that this works
            defer f.Close()
        }
    }
}

func (lg *AppLogger) Create() {

    lg.maps = make(map[string]*log.Logger)
    lg.files = make([]*os.File, 10)

    fileName := fmt.Sprint("log/Error/Error_", time.Now().Format(time.RFC3339), ".log")
    symLink := "log/Error.log"
    file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
    if err != nil {
        fmt.Println("Error creating log file: ", err.Error())
        return
    }

    lg.files = append(lg.files, file)

    identifier := "SYSTEM"
    lg.maps[identifier] = log.New(file, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)

    lg.Send_To_Log(identifier, INFO, "========== New log ==========")

    os.Remove(symLink)
    err = os.Symlink(strings.TrimLeft(fileName, "log/"), symLink)
    if err != nil {
        lg.Send_To_Log(identifier, ERROR, fmt.Sprint("Error creating symlink: ", err.Error()))
    }

}

func (lg *AppLogger) SetPackageLog(identifier string, fileName string, symlink string) {

    file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
    if err != nil {
        lg.Send_To_Log("SYSTEM", ERROR, fmt.Sprint("Error creating log file: ", err.Error()))
        return
    }

    lg.maps[identifier] = log.New(file, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)

    lg.Send_To_Log(identifier, INFO, "========== New log ==========")

    os.Remove(symlink)
    err = os.Symlink(strings.TrimLeft(fileName, "log/"), symlink)
    if err != nil {
        lg.Send_To_Log("SYSTEM", ERROR, fmt.Sprint("Error creating symlink: ", err.Error()))
    }

}

func (lg *AppLogger) Send_To_Log(identifier string, logLevel int, logMessage string) {
    switch logLevel {

    //info, print to the package log
    case INFO :
        l, ok := lg.maps[identifier]
        if ok {
            l.Println(logMessage)
        }

    // error, print to errorlog
    case ERROR :
        l, ok := lg.maps["SYSTEM"]
        if ok {
            l.Println("ERROR: ", identifier, " : ", logMessage)
        }

    // fatal, close application
    case FAILURE :
        l, ok := lg.maps["SYSTEM"]
        if ok {
            l.Println("FAILURE: ", identifier, " : ", logMessage)
            os.Exit(1) // TODO not this dirty
        }
    default :
        l, ok := lg.maps["SYSTEM"]
        if ok {
            l.Println("INFO: ", identifier, ": Incorrect loglevel. Log Message: ", logMessage)
        } else {
            fmt.Println("Incorrect loglevel. Log Message: ", logMessage)
        }
    }
}
