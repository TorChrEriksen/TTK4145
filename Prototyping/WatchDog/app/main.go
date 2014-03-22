package main

import(
    "fmt"
    "os"
    "os/signal"
//    "os/exec"
    "time"
    "syscall"
    "strconv"
    "bufio"
    "strings"
)

const (
    START_PRIMARY = 0
    START_SECONDARY = 1
	WD_ALIVE = 2
	WD_OFFLINE = 3
)

const (
    NO_INFO = 0
    INFO = 1
)

var TIMEOUT = time.Duration(time.Second * 5)

// Only for debugging/testing
func killProc() {
    proc, err := os.FindProcess(os.Getpid())
    if err != nil {
        fmt.Println(err.Error())
        return
    }

    err = proc.Kill()
    if err != nil {
        fmt.Println(err.Error())
        return
    }
}

func waitForAliveFromWD(signalChan chan os.Signal, obsChan chan int) {

    fmt.Println("Primary: waiting for signal from WD")

    timer := time.NewTimer(TIMEOUT)
    go func() {
        <-timer.C

        // Restart WD if it times out
        obsChan <- -1
        fmt.Println("WD timed out!")
    }()

    for sig := range signalChan {

        _ = sig

//        fmt.Println("Primary: signal received from WD: ", sig)
        timer.Reset(TIMEOUT)
    }
}

func waitForWdCommand(upgChan chan bool) {
    fmt.Println("Secondary: waiting for command from wd")

    ch := make(chan os.Signal, 1)
    signal.Notify(ch, syscall.SIGUSR1)

    for sig := range ch {

        fmt.Println("Secondary: RECEIVED COMMAND from WD: ", sig)

        switch sig {
        case syscall.SIGUSR1:
            close(ch)
            upgChan <- true
        default:
            fmt.Println("Unknown command received from WD")
        }
    }

    fmt.Println("waitForWdCommand() finished")
}


/*
func waitForAliveFromPrimary() {

    fmt.Println("Secondary: waiting for signal from Primary")

    ch := make(chan os.Signal, 1)
    signal.Notify(ch, syscall.SIGHUP)

    timer := time.NewTimer(TIMEOUT)
    go func() {
        <-timer.C

        close(ch)
//        takeOver()
    }()

    for sig := range ch {

        fmt.Println("Secondary: signal received from Primary: ", sig)
        timer.Reset(TIMEOUT)
    }
}
*/

func spawnCopy(wdPID int) (*os.Process, error) {

    // fmt.Println("spawning copy")

    argv := []string{os.Args[0], strconv.Itoa(START_SECONDARY), strconv.Itoa(wdPID)}
    attr := new(os.ProcAttr)
    attr.Files = []*os.File{nil, os.Stdout, os.Stderr}
    proc, err := os.StartProcess("main", argv, attr)
    return proc, err
}

func spawnWD(priPID int) (*os.Process, error) {

    // fmt.Println("spawning WD")

    argv := []string{"WD", strconv.Itoa(priPID)}
    attr := new(os.ProcAttr)
    attr.Files = []*os.File{nil, os.Stdout, os.Stderr}
    proc, err := os.StartProcess("WD", argv, attr)
    return proc, err

}

// Primary uses this to notify WD
func notifyPrimaryAlive(p *os.Process, ch chan bool) {
    halt := false

    go func() {
        <-ch
        halt = true
    }()
    for {
        if halt {
            break
        }
        time.Sleep(time.Second)
//        fmt.Println("Primary alive, sending signal")
        p.Signal(syscall.SIGHUP) // Signal from primary
//        p2.Signal(syscall.SIGHUP) // Signal from primary
    }
}

// Secondary notifies WD
func notifySecondaryAlive(p *os.Process, ch chan bool) {
    halt := false

    go func() {
        for term := range ch {
            halt = term
            break
        }
    }()
    for {
        if halt {
            break
        }
        time.Sleep(time.Second)
//        fmt.Println("Secondary alive, sending signal")
        p.Signal(syscall.SIGFPE)
    }
}

func operate() {
    for {
        fmt.Println("Operating....")
        time.Sleep(time.Second * 5)
    }
}

func writePidToFile(filename string) {
    // Remove file if it already exists
    os.Remove(filename)

    file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
    if err != nil {
        fmt.Println("Error creating pid file: ", err.Error())
        os.Exit(0)
    }

    pid := strconv.Itoa(os.Getpid())
    n, err := file.Write([]byte(pid))

    if err != nil {
        fmt.Println("Error writing pid to file: ", err.Error())
        os.Exit(0)
    }

    fmt.Println("Wrote ", n, " bytes to file.")
    defer file.Close()

}

func main() {
//    fmt.Println("Welcome to the redundant Go app...")

//    fmt.Println(len(os.Args), os.Args)

    if len(os.Args) == 2 { // Should be primary
        arg1, err1 := strconv.Atoi(os.Args[1])
        //arg2, err2 := strconv.Atoi(os.Args[2])
        if err1 != nil {
            fmt.Println("Primary: Invalid argument (1)")
            os.Exit(0)
//        } else if err2 != nil {
//            fmt.Println("Invalid argument (2)")
//            os.Exit(0)
        } else {
            if arg1 == START_PRIMARY {

                /**
                 * Here we want to start the secondary application and
                 * start the watch dog. The watch dog will get the PID of both
                 */

                 // Start watch dog
                 wd, err := spawnWD(os.Getpid())
                 if err != nil {
                     fmt.Println("Error main() -> spawnWD(): ", err.Error())
                     os.Exit(0)
                 }

                 // Start secondary
                _, err = spawnCopy(wd.Pid)
                if err != nil {
                    fmt.Println("Error main() -> spawnCopy(): ", err.Error())
                    os.Exit(0)
                }

                /*
                if arg2 == NO_INFO {
                    PRINT_INFO = false
                } else if arg2 == INFO {
                    PRINT_INFO = true
                } else {
                    fmt.Println("Invalid argument (2)")
                    os.Exit(0)
                }

                if PRINT_INFO {
                    fmt.Println("Primary")
                }
                */

                // Nasty that program is still alive if main thread dies?
                ch := make(chan int)
                haltChan := make(chan bool)

                wdChan := make(chan int)
                wdSignalChan := make(chan os.Signal, 1)
                signal.Notify(wdSignalChan, syscall.SIGILL)

                go operate()
                go notifyPrimaryAlive(wd, haltChan)

                go func() {
                    for {
                        go waitForAliveFromWD(wdSignalChan, wdChan)
                        <-wdChan

                        fmt.Println("Primary: WD DIED!!")
                        haltChan <- true

                        // WD died, halt notify WD, kill secondary, restart WD, restard secondary.

                        // Open file and read PID so that we can kill secondary
                        file, err := os.Open("secondaryPID")
                        if err != nil {
                            fmt.Println("There was an error opening the SECONDARY PID file")
                            //break
                            os.Exit(0) // Remove all os.Exit's?
                        } else {
                            reader := bufio.NewReader(file)
                            val, _ := reader.ReadString('\n')
                            val = strings.Trim(val, "\n")
                            pid, err := strconv.Atoi(val)

                            if err != nil {
                                fmt.Println("There was an error converting the data to an int")
                            } else {

                                // We got the PID for secondary
                                proc, err := os.FindProcess(pid)
                                if err != nil {
                                    fmt.Println("Error finding the process for secondary PID: ", pid, ". Error: ", err.Error())
                                    os.Exit(0)
                                }

                                // Kill secondary
                                err = proc.Kill()
                                if err != nil {
                                    fmt.Println("Error killing secondary process: ",  err.Error())
                                    os.Exit(0)
                                }
                            }
                        }
                        defer file.Close()

                        // Restart WD
                        wd, err := spawnWD(os.Getpid())
                        if err != nil {
                            fmt.Println("Error restarting WD: ", err.Error())
                            os.Exit(0) // Remove all os.Exit's ?
                        }
                        fmt.Println("Primary: WD RESTARTED")

                        // Restart secondary
                        _, err = spawnCopy(wd.Pid)
                        if err != nil {
                            fmt.Println("Error main() -> spawnCopy(): ", err.Error())
                            os.Exit(0)
                        }
                        fmt.Println("Primary: SECONDARY RESTARTED")

                        // Restart notification
                        go notifyPrimaryAlive(wd, haltChan)
                    }
                }()

                for i := range ch {
                    fmt.Println(i)
                }

            }
        }
    } else if len(os.Args) == 3 { // Should be secondary
        arg1, err1 := strconv.Atoi(os.Args[1])
        wdPID, err2 := strconv.Atoi(os.Args[2])
        if err1 != nil {
            fmt.Println("Secondary: Invalid argument (1)")
            os.Exit(0)
        } else if err2 != nil {
            fmt.Println("Secondary: Invalid argument (2)")
            os.Exit(0)
        } else {
            if arg1 == START_SECONDARY {

                /*
                if arg2 == NO_INFO {
                    PRINT_INFO = false
                } else if arg2 == INFO {
                    PRINT_INFO = true
                } else {
                    fmt.Println("Invalid argument (2)")
                    os.Exit(0)
                }

                if PRINT_INFO {
                    fmt.Println("Secondary")
                }
                */

                writePidToFile("secondaryPID")

                wd, err := os.FindProcess(wdPID)
                if err != nil {
                    fmt.Println("Secondary: There was an error finding the Watch Dog process: ", err.Error())
                    return
                }

                stopNotifyChan := make(chan bool)
                upgChan := make(chan bool)

//                go waitForAliveFromPrimary()
                go notifySecondaryAlive(wd, stopNotifyChan)
                go waitForWdCommand(upgChan)

                for upgraded := range upgChan {

                    // Use upgraded?
                    _ = upgraded

                    stopNotifyChan <- true
                    fmt.Println("Secondary is now PRIMARY!")

                    ch := make(chan int)
                    haltChan := make(chan bool)

                    wdChan := make(chan int)
                    wdSignalChan := make(chan os.Signal, 1)
                    signal.Notify(wdSignalChan, syscall.SIGILL)

                    // go operate()
                    go notifyPrimaryAlive(wd, haltChan)

                    go func() {
                        for {
                            go waitForAliveFromWD(wdSignalChan, wdChan)
                            <-wdChan

                            fmt.Println("Primary: WD DIED!!")
                            haltChan <- true

                            // WD died, halt notify WD, kill secondary, restart WD, restard secondary.

                            // Open file and read PID so that we can kill secondary
                            file, err := os.Open("secondaryPID")
                            if err != nil {
                                fmt.Println("There was an error opening the SECONDARY PID file")
                                //break
                                os.Exit(0) // Remove all os.Exit's?
                            } else {
                                reader := bufio.NewReader(file)
                                val, _ := reader.ReadString('\n')
                                val = strings.Trim(val, "\n")
                                pid, err := strconv.Atoi(val)

                                if err != nil {
                                    fmt.Println("There was an error converting the data to an int")
                                } else {

                                    // We got the PID for secondary
                                    proc, err := os.FindProcess(pid)
                                    if err != nil {
                                        fmt.Println("Error finding the process for secondary PID: ", pid, ". Error: ", err.Error())
                                        os.Exit(0)
                                    }

                                    // Kill secondary
                                    err = proc.Kill()
                                    if err != nil {
                                        fmt.Println("Error killing secondary process: ",  err.Error())
                                        os.Exit(0)
                                    }
                                }
                            }
                            defer file.Close()

                            // Restart WD
                            wd, err := spawnWD(os.Getpid())
                            if err != nil {
                                fmt.Println("Error restarting WD: ", err.Error())
                                os.Exit(0) // Remove all os.Exit's ?
                            }
                            fmt.Println("Primary: WD RESTARTED")

                            // Restart secondary
                            _, err = spawnCopy(wd.Pid)
                            if err != nil {
                                fmt.Println("Error main() -> spawnCopy(): ", err.Error())
                                os.Exit(0)
                            }
                            fmt.Println("Primary: SECONDARY RESTARTED")

                            // Restart notification
                            go notifyPrimaryAlive(wd, haltChan)
                        }
                    }()

                    for i := range ch {
                        fmt.Println(i)
                    }

                }

            } else {

                fmt.Println("Invalid argument (1)")
                os.Exit(0)
            }
        }
    } else {
        fmt.Println("Invalid number of arguments")
        os.Exit(0)
    }
}
