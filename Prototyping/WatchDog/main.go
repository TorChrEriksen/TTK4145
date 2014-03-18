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

var TIMEOUT = time.Duration(time.Second * 5)

func restartPrimaryAsSecondary(wdPID int) (*os.Process, error) {
    argv := []string{"./main", strconv.Itoa(1), strconv.Itoa(wdPID)} // 1 = START_SECONDARY
    attr := new(os.ProcAttr)
    attr.Files = []*os.File{nil, os.Stdout, os.Stderr}
    proc, err := os.StartProcess("main", argv, attr) // need struct to keep track of the PIDs

    return proc, err
}

func restartSecondary(wdPID int) (*os.Process, error) {
    argv := []string{"./main", strconv.Itoa(1), strconv.Itoa(wdPID)} // 1 = START_SECONDARY
    attr := new(os.ProcAttr)
    attr.Files = []*os.File{nil, os.Stdout, os.Stderr}
    proc, err := os.StartProcess("main", argv, attr) // need struct to keep track of the PIDs

    return proc, err
}

func upgradeSecondary(secPid int) {
    proc, err := os.FindProcess(secPid)
    if err != nil {
        fmt.Println("Upgrade: there was an error finding the Secondary process: ", err.Error())
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

    timer := time.NewTimer(TIMEOUT)
    go func() {
        <-timer.C

        // Primary timed out.
        obsChan <- 1
    }()

    for sig := range signalChan {

        _ = sig

//        fmt.Println("WD: signal received from Primary: ", sig)
        timer.Reset(TIMEOUT)
    }

    fmt.Println("WD: Exited waitForAliveFromPrimary, SWAG!!!")
}

func waitForAliveFromSecondary(signalChan chan os.Signal, obsChan chan int) {

    timer := time.NewTimer(TIMEOUT)
    go func() {
        <-timer.C

        // Secondary timed out.
        obsChan <- -1
    }()

    for sig := range signalChan {

        _ = sig

//        fmt.Println("WD: signal received from Secondary: ", sig)
        timer.Reset(TIMEOUT)

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

    fmt.Println("WD: Exited waitForAliveFromSecondary, YOLO!!!")
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

        fmt.Println("Received Primary PID: ", procIDs.Primary)

        procIDs.Self = os.Getpid()
        fmt.Println("WatchDog PID: ", procIDs.Self)

        procIDs.Primary = primary

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
                        fmt.Println("Setting secondary PID: ", secondary)
                        procIDs.Secondary = secondary
                    }
                }
            }
        }()

        /*
    case 3:
        primary, priErr := strconv.Atoi(os.Args[1])
        if priErr != nil {
            fmt.Println("Error getting primary PID: ", priErr.Error())
            os.Exit(1)
        }

        fmt.Println("Received Primary PID: ", procIDs.Primary) 

        secondary, secErr := strconv.Atoi(os.Args[2])
        if secErr != nil {
            fmt.Println("Error getting primary PID: ", priErr.Error())
            os.Exit(1)
        }

        fmt.Println("Received Secondary PID: ", procIDs.Secondary)

        procIDs.Self = os.Getpid()
        fmt.Println("WatchDog PID: ", procIDs.Self)

        procIDs.Primary = primary
        procIDs.Secondary = secondary
        */

    default:
        procIDs.Self = os.Getpid()
        fmt.Println("WatchDog PID: ", procIDs.Self)
    }

    // Receive heartbeat from Primary

    // Receive heartbeat from Secondary

		// Check if application was terminated correctly

		// Debug code
		for {
			time.Sleep(time.Second * 2)
		}
}

/*
func continueOperation() {

    // Check if file is available
    file, err := os.Open("testfile")
    if err != nil {
        fmt.Println("Error os.Open(): ", err.Error())
    }

    reader := bufio.NewReader(file)
    lastValue, _ := reader.ReadString('\n')
    lastValue = strings.Trim(lastValue, "\n")

    pwd, err := os.Getwd()
    if err != nil {
        fmt.Println("Error os.Getwd(): ", err.Error())
        return
    }

    app := "sh"
    arg0 := pwd + "/test.sh"

    for {
        i++
        arg1 := strconv.Itoa(i)
        cmd := exec.Command(app, arg0, arg1)
        _, err := cmd.Output()
        if err != nil {
            fmt.Println("Error cmd.Output(): ", err.Error())
        }
        fmt.Println(i)
        time.Sleep(time.Second)
    }

}

func operate() {
    var i int = 0
    pwd, err := os.Getwd()
    if err != nil {
        fmt.Println("Error Getwd(): ", err.Error())
        return
    }

    app := "sh"
    arg0 := pwd + "/test.sh"

    for {
        i++
        arg1 := strconv.Itoa(i)
        cmd := exec.Command(app, arg0, arg1)
        _, err := cmd.Output()
        if err != nil {
            fmt.Println("Error Command(): ", err.Error())
        }
        fmt.Println(i)
        time.Sleep(time.Second)
    }
}
*/
