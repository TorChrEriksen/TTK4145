package main

import (
    "fmt"
	"time"
    "strconv"
    "os"
    "os/signal"
    "syscall"
    "bufio"
    "strings"
)

var APP_TIMEOUT = time.Duration(time.Second * 5)

func restartPrimaryAsSecondary(wdPID int) (*os.Process, error) {
    argv := []string{"./elevApp", strconv.Itoa(1), strconv.Itoa(wdPID)} // 1 = START_SECONDARY
    attr := new(os.ProcAttr)
    attr.Files = []*os.File{nil, os.Stdout, os.Stderr}
    proc, err := os.StartProcess("elevApp", argv, attr) // need struct to keep track of the PIDs

    return proc, err
}

func restartSecondary(wdPID int) (*os.Process, error) {
    argv := []string{"./elevApp", strconv.Itoa(1), strconv.Itoa(wdPID)} // 1 = START_SECONDARY
    attr := new(os.ProcAttr)
    attr.Files = []*os.File{nil, os.Stdout, os.Stderr}
    proc, err := os.StartProcess("elevApp", argv, attr) // need struct to keep track of the PIDs

    return proc, err
}

func upgradeSecondary(secPid int) {
    proc, err := os.FindProcess(secPid)
    if err != nil {
        fmt.Println("Watch Dog: there was an error finding the Secondary process to upgrade it: ", err.Error())
        return
    }

    proc.Signal(syscall.SIGUSR1)
}

type ProcessIDs struct {
    Primary int
	Secondary int
	Self int
}

func waitForAliveFromPrimary(signalChan chan os.Signal, obsChan chan int) {

    timer := time.NewTimer(APP_TIMEOUT)
    go func() {
        <-timer.C

        // Primary timed out.
        obsChan <- 1
    }()

    for sig := range signalChan {

        _ = sig

//        fmt.Println("WD: signal received from Primary: ", sig)
        timer.Reset(APP_TIMEOUT)
    }

    //fmt.Println("WD: Exited waitForAliveFromPrimary, SWAG!!!")
}

func waitForAliveFromSecondary(signalChan chan os.Signal, obsChan chan int) {

    timer := time.NewTimer(APP_TIMEOUT)
    go func() {
        <-timer.C

        // Secondary timed out.
        obsChan <- -1
    }()

    for sig := range signalChan {

        _ = sig

//        fmt.Println("WD: signal received from Secondary: ", sig)
        timer.Reset(APP_TIMEOUT)

        // Open file and read PID so that we can 
        file, err := os.Open("secondaryPID")
        if err != nil {
            fmt.Println("There was an error opening the SECONDARY PID file")
        } else {
            reader := bufio.NewReader(file)
            val, _ := reader.ReadString('\n')
            val = strings.Trim(val, "\n")
            pid, err := strconv.Atoi(val)

            if err != nil {
                fmt.Println("There was an error converting the data to an int")
            } else {
                obsChan <- pid
            }
        }
        defer file.Close()
    }

    //fmt.Println("WD: Exited waitForAliveFromSecondary, YOLO!!!")
}

func notifyAlive(priPID int, ch chan bool) {
    halt := false
    priProcess, err := os.FindProcess(priPID)
    if err != nil {
        fmt.Println("There was an error finding the Primary process: ", err.Error())
        return
    }

    go func() {
        <-ch
        halt = true
//        halt <-ch
        /*
        for term := range ch {
            halt = term
            break
        }
        */
    }()

    for {
        if halt {
            break
        }
        time.Sleep(time.Second)
        priProcess.Signal(syscall.SIGILL)
    }
}

func main() {
    procIDs := ProcessIDs{}

    switch len(os.Args) {
    case 2:
        primary, err := strconv.Atoi(os.Args[1])
        if err != nil {
            fmt.Println("Error getting primary PID: ", err.Error())
            os.Exit(1)
        }

        procIDs.Self = os.Getpid()
        fmt.Println("WatchDog PID: ", procIDs.Self)

        procIDs.Primary = primary
        fmt.Println("Watch Dog: received Primary PID: ", procIDs.Primary)

        priChan := make(chan int)
        secChan := make(chan int)
        haltChan := make(chan bool)

        secSignalChan := make(chan os.Signal, 1)
        signal.Notify(secSignalChan, syscall.SIGFPE)

        priSignalChan := make(chan os.Signal, 1)
        signal.Notify(priSignalChan, syscall.SIGHUP)

        go notifyAlive(procIDs.Primary, haltChan) // Do we need a chan here? no?
        go waitForAliveFromPrimary(priSignalChan, priChan)
        go waitForAliveFromSecondary(secSignalChan, secChan)

        go func() {
            for {
                select {
                case primary := <-priChan:

                    _ = primary

                    // Primary timed out
                    // Make secondary new primary
                    // Restart primary as secondary
                    haltChan <- true
                    go waitForAliveFromPrimary(priSignalChan, priChan)

                    fmt.Println(procIDs.Secondary)
                    procIDs.Primary = procIDs.Secondary
                    upgradeSecondary(procIDs.Secondary)

                    proc, err := restartPrimaryAsSecondary(os.Getpid())
                    if err != nil {
                        fmt.Println("Error restaring primary: ", err.Error())
                    } else {
                        fmt.Println("Primary was restarted successfully as secondary")
                        procIDs.Secondary = proc.Pid
                        go notifyAlive(procIDs.Primary, haltChan)
                    }
                case secondary := <-secChan:

                    // Secondary terminated
                    if secondary == -1 {

                        // Secondary timed out
                        // Start listening for new secondary
                        // Restart secondary
                        go waitForAliveFromSecondary(secSignalChan, secChan)

                        proc, err := restartSecondary(os.Getpid())
                        if err != nil {
                            fmt.Println("Error restarting secondary: ", err.Error())
                        } else {
                            fmt.Println("Secondary was restarted successfully")
                            procIDs.Secondary = proc.Pid
                        }
                    } else { // Set PID
                        //fmt.Println("Setting secondary PID: ", secondary)
                        procIDs.Secondary = secondary
                    }
                }
            }
        }()

    default:
        procIDs.Self = os.Getpid()
        fmt.Println("WatchDog PID: ", procIDs.Self)
    }

		// Debug code
		for {
			time.Sleep(time.Second * 2)
		}
}
