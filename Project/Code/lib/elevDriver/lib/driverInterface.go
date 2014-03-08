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
//Buttons
const STOP = C.STOP
const INT_BTN_1 = C.FLOOR_COMMAND1
const INT_BTN_2 = C.FLOOR_COMMAND2
const INT_BTN_3 = C.FLOOR_COMMAND3
const INT_BTN_4 = C.FLOOR_COMMAND4
// Test function for elevator
func Test_Run() {
    fmt.Println("Attempting to run elevator")
    err := C.run()
    if err != 0 {
        fmt.Println("Test program terminated abnormally")
    }
    fmt.Println("Test program terminated normally")
}

// start: real driver functions 
// INPUT to controller
func Init(){
  C.elev_init()
}

func SetSpeed(speed C.int){
  //fmt.Println("Attmepting to set speed to ", speed)
  C.elev_set_speed(speed)
}

func StopElevator(){
   C.elev_set_speed(0)
}

//TODO Consider using const UP = 1 instead of strings
//func SetButtonLamp(btn C.elev_button_type_t,floor C.int, value C.int){
func SetButtonLamp(direction string,floor C.int, value C.int){
  if direction == "UP"{
      fmt.Println("goin' up")
      C.elev_set_button_lamp(C.BUTTON_CALL_UP, floor, value)
  }else if direction == "DOWN"{
      fmt.Println("goin' down")
      C.elev_set_button_lamp(C.BUTTON_CALL_DOWN, floor, value)
  }else{
      //Trying to kick the shit out of software in case of wrong input
      fmt.Println("Panicking")
      panic(fmt.Sprintf("%v", direction)) 
  } 
}

func SetFloorLamp(floor C.int){
   C.elev_set_floor_indicator(floor)
}

func SetStopLamp(value C.int){
  C.elev_set_stop_lamp(value)
}

func SetDoorLamp(value C.int){
  C.elev_set_door_open_lamp(value)  
}


//OUTPUT from controller
func GetFloorSignal() int{
   return int(C.elev_get_floor_sensor_signal())
}

func GetStopSignal() int{
   return int(C.elev_get_stop_signal())
}

func GetObstructionSignal() int{
   return int(C.elev_get_obstruction_signal())
}

func GetButtonSignal() int{
   //C.elev_get_button_signal(elev_button_type_t button, int floor)
   return int(C.io_read_bit(STOP))
}
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
		ch <- GetButtonSignal() // this need to be changed!!
	}
}

func createStopListener(ch chan C.int){
   
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

