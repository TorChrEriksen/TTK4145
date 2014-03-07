package driverInterface

// #cgo CFLAGS: -std=c99 -g -Wall -O2 -I .
// #cgo LDFLAGS: -lcomedi -g -lm
// #include "C/io.c"
// #include "C/elev.c"
// #include "C/runner.c"
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

// Parameters: all the different channels we need like:
// Create(buttonChannel, floorChannel, stopChannel, ....)
func Create(blChan chan int) {
	go createButtonListener(blChan)
}

func createButtonListener(ch chan int) {
	// Wait for IO

	// Register what kind of IO that was changed.

	// Notify on channel about change in IO
	for {
		ch <- 10 // this need to be changed!!
	}
}

//TODO Implement all input functions
func SetButtonLamp(btn C.elev_button_type_t,floor int, value int){
  C.elev_set_button_lamp(btn, floor, value){
}

func SetSpeed(speed int){
  C.elev_set_speed(speed)
}

func Init(){
  C.elev_init()
}

func SetStopLamp(value int){
  C.elev_stop_lamp(value)
}

func SetDoorLamp(value int){
  C.elev_set_door_open_lamp(value)  
}


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

