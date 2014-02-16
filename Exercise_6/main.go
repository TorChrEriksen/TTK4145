package main

import(
    "fmt"
    "os"
    "os/signal"
    "os/exec"
    "time"
    "syscall"
    "strconv"
    "bufio"
    "strings"
)

const (
    SECONDARY = 0
    PRIMARY = 1
)

const (
    NO_INFO = 0
    INFO = 1
)

var PRINT_INFO bool = false
var TIMEOUT = time.Duration(time.Second * 5)

// killProc() is not used in main program, only used for debugging/testing.
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

func takeOver() {
    spawn, err := spawnCopy()
    if err != nil {
        fmt.Println(err.Error())
        os.Exit(0)
    }

    go continueOperation()

    ch := make(chan int)

    go notifyAlive(spawn, ch)

    if PRINT_INFO {
        fmt.Println("Secondary is now Primary")
    }

    for i := range ch {
        fmt.Println(i)
    }

}

func waitForAlive(waiter chan int) {
    if PRINT_INFO {
        fmt.Println("waitforAlive")
    }
    ch := make(chan os.Signal, 1)
    signal.Notify(ch, syscall.SIGHUP)

    timer := time.NewTimer(TIMEOUT)
    go func() {
        <-timer.C
        if PRINT_INFO {
            fmt.Println("Timer Expired, Secondary is taking over")
        }
        close(ch)
        takeOver()
    }()

    for sig := range ch {
        fmt.Println("Signal received: ", sig)
        timer.Reset(TIMEOUT)
    }

}

func spawnCopy() (*os.Process, error) {
    if PRINT_INFO {
        fmt.Println("Spawning copy of ourself")
    }
    argv := []string{os.Args[0], strconv.Itoa(SECONDARY), os.Args[2]}
    attr := new(os.ProcAttr)
    attr.Files = []*os.File{nil, os.Stdout, os.Stderr}
    proc, err := os.StartProcess("main", argv, attr)
    return proc, err
}

func notifyAlive(p *os.Process, waiter chan int) {
    for {
        if PRINT_INFO {
            fmt.Println("Notifying")
        }
        time.Sleep(time.Second)
        p.Signal(syscall.SIGHUP)
    }
}

func operate() {
    var i int = 0
    pwd, err := os.Getwd()
    if err != nil {
        fmt.Println("Error: ", err.Error())
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
            fmt.Println("Error: ", err.Error())
        }
        fmt.Println(i)
        time.Sleep(time.Second)
    }
}

func continueOperation() {

    // Read last operation from file
    file, err := os.Open("testfile")
    if err != nil {
        fmt.Println("Error: ", err.Error())
    }
    reader := bufio.NewReader(file)
    lastValue, _ := reader.ReadString('\n')
    lastValue = strings.Trim(lastValue, "\n")

    i, err := strconv.Atoi(lastValue)
    if err != nil {
        fmt.Println("Error: ", err.Error())
    }

    pwd, err := os.Getwd()
    if err != nil {
        fmt.Println("Error: ", err.Error())
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
            fmt.Println("Error: ", err.Error())
        }
        fmt.Println(i)
        time.Sleep(time.Second)
    }
}

func main() {
//    fmt.Println("Welcome to the redundant Go app...")

//    fmt.Println(len(os.Args), os.Args)

    if len(os.Args) == 3 {
        arg1, err1 := strconv.Atoi(os.Args[1])
        arg2, err2 := strconv.Atoi(os.Args[2])
        if err1 != nil {
            fmt.Println("Invalid argument (1)")
            os.Exit(0)
        } else if err2 != nil {
            fmt.Println("Invalid argument (2)")
            os.Exit(0)
        } else {
            if arg1 == PRIMARY {

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

                spawn, err := spawnCopy()
                if err != nil {
                    fmt.Println(err.Error())
                    os.Exit(0)
                }

                // Nasty that program is still alive if main thread dies?
                ch := make(chan int)
                go notifyAlive(spawn, ch)
                go operate()

                for i := range ch {
                    fmt.Println(i)
                }

            } else if arg1 == SECONDARY {

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

                ch := make(chan int)
                go waitForAlive(ch)

                for i := range ch {
                    fmt.Println(i)
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
