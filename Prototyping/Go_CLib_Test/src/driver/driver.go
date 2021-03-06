package driver

// #cgo CFLAGS: -std=c99 -g -Wall -O2 -I .
// #cgo LDFLAGS: -lcomedi -g -lm
// #include "../../lib/C/io.c"
// #include "../../lib/C/elev.c"
// #include "../../lib/C/runner.c"
import "C"

import (
    "fmt"
)

// Test function for elevator
func Test_Run() {
    err := C.run()
    if err != 0 {
        fmt.Println("Test program terminated abnormally")
    }
    fmt.Println("Test program terminated normally")
}

// start: real driver functions

// end:  real driver functions.












/* start: testcode

    // Initialize hardware
    err := C.elev_init()
    if err == 0 {
        fmt.Println("Unable to initialize elevator hardware")
        os.Exit(0)
    }
    
    fmt.Println("Press STOP button to stop elevator and exit program.")
    C.elev_set_speed(50)
    
    for {
        
        if C.elev_get_stop_signal() == 1 {
            stopElevator()
            break
        }
    }


func stopElevator() {
    fmt.Println("The elevator has come to a conclusion: stopping")
    C.elev_set_speed(0)
}

end: testcode */

