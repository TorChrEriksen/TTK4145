package driverInterface

// #cgo CFLAGS: -std=c99 -g -Wall -O2 -I .
// #cgo LDFLAGS: -lcomedi -g -lm
// #include "C/io.c"
// #include "C/elev.c"
// #include "C/runner.c"
import "C"

import (
    "fmt"
    "time"
)

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

func SetSpeed(speed int){
  //fmt.Println("Attmepting to set speed to ", speed)
  C.elev_set_speed(C.int(speed))
}

func StopElevator(){
   C.elev_set_speed(0)
}

//TODO Consider using const UP = 1 instead of strings
//func SetButtonLamp(btn C.elev_button_type_t,floor C.int, value C.int){
func SetButtonLamp(direction string, floor int, value int){
  if direction == "UP"{
      fmt.Println("goin' up")
      C.elev_set_button_lamp(C.BUTTON_CALL_UP, C.int(floor), C.int(value))
  }else if direction == "DOWN"{
      fmt.Println("goin' down")
      C.elev_set_button_lamp(C.BUTTON_CALL_DOWN, C.int(floor), C.int(value))
  }else if direction == "INT"{
      C.elev_set_button_lamp(C.BUTTON_COMMAND, C.int(floor), C.int(value))
  }else{
      //Trying to kick the shit out of software in case of wrong input
      fmt.Println("Panicking")
      panic(fmt.Sprintf("%v", direction)) 
  } 
}

func SetFloorLamp(floor int){
   C.elev_set_floor_indicator(C.int(floor))
}

func SetStopLamp(value int){
  C.elev_set_stop_lamp(C.int(value))
}

func SetDoorLamp(value int){
  C.elev_set_door_open_lamp(C.int(value))
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

// Parameters: all the different channels we need like:
// Create(buttonChannel, floorChannel, stopChannel, ....)

func Create(intBtChan chan int, floorChan chan int, stopChan chan int, extBtChan chan int, timeoutChan chan int) {
	go createIntButtonListener(intBtChan)
    go createFloorListener(floorChan)
    go createStopListener(stopChan)
    go createExtButtonListener(extBtChan)
    go createTimeoutListener(timeoutChan)
}

const INT_BTN_1 int  = 11
const INT_BTN_2 int  = 12
const INT_BTN_3 int  = 13
const INT_BTN_4 int  = 14
func createIntButtonListener(ch chan int) {
    //TODO think about fault tolerance here    
    var prevStateIntBtn1 int = 0
    var StateIntBtn1 int = 0
    var prevStateIntBtn2 int = 0
    var StateIntBtn2 int = 0
    var prevStateIntBtn3 int = 0
    var StateIntBtn3 int = 0
    var prevStateIntBtn4 int = 0
    var StateIntBtn4 int = 0
    for{
        time.Sleep(time.Millisecond * 10)
        
        StateIntBtn1 = int(C.io_read_bit(C.FLOOR_COMMAND1))
        if StateIntBtn1 != prevStateIntBtn1{
            prevStateIntBtn1 = StateIntBtn1
            if StateIntBtn1 == 1{
                ch <- INT_BTN_1            
            }
        }

        StateIntBtn2 = int(C.io_read_bit(C.FLOOR_COMMAND2))
        if StateIntBtn2 != prevStateIntBtn2{
            prevStateIntBtn2 = StateIntBtn2
            if StateIntBtn2 == 1{
                ch <- INT_BTN_2            
            }
        }
        
        StateIntBtn3 = int(C.io_read_bit(C.FLOOR_COMMAND3))
        if StateIntBtn3 != prevStateIntBtn3{
            prevStateIntBtn3 = StateIntBtn3
            if StateIntBtn3 == 1{
                ch <- INT_BTN_3            
            }
        }

        StateIntBtn4 = int(C.io_read_bit(C.FLOOR_COMMAND4))
        if StateIntBtn4 != prevStateIntBtn4{
            prevStateIntBtn4 = StateIntBtn4
            if StateIntBtn4 == 1{
                ch <- INT_BTN_4            
            }
        }
    }	
}

const FLOOR_NO int  = 30
const FLOOR_1 int  = 31
const FLOOR_2 int  = 32
const FLOOR_3 int  = 33
const FLOOR_4 int  = 34
func createFloorListener(ch chan int){
    var prevStateFloor int = 0
    var StateFloor int = 0
    for{
        time.Sleep(time.Millisecond * 10)
        StateFloor = int(C.elev_get_floor_sensor_signal())
        if StateFloor != prevStateFloor{
            prevStateFloor = StateFloor
            if StateFloor == -1{
                ch <- FLOOR_NO           
            }
            if StateFloor == 0{
                ch <- FLOOR_1
                SetFloorLamp(0)           
            }
            if StateFloor == 1{
                ch <- FLOOR_2
                SetFloorLamp(1)           
            }
            if StateFloor == 2{
                ch <- FLOOR_3
                SetFloorLamp(2)          
            }
            if StateFloor == 3{
            	SetFloorLamp(3)
                ch <- FLOOR_4           
            }
        }
    }
}

const STOP int = 10
func createStopListener(ch chan int){
   var prevStateStop int = 0
   var StateStop int = 0
   for{
        time.Sleep(time.Millisecond * 10)
        StateStop = int(C.elev_get_stop_signal())
        if StateStop != prevStateStop{
            prevStateStop = StateStop
            if StateStop == 1{
                ch <- STOP            
            }
        }
    }
}

//Buttons not used added to highlight pattern
const EXT_BTN_1_UP int = 41
//const EXT_BTN_1_DOWN int = 40
const EXT_BTN_2_UP int = 51
const EXT_BTN_2_DOWN int = 50
const EXT_BTN_3_UP int = 61
const EXT_BTN_3_DOWN int = 60
//const EXT_BTN_4_UP int = 71
const EXT_BTN_4_DOWN int = 70

func createExtButtonListener(ch chan int){
    var prevStateExtBtn1Up int = 0
    var StateExtBtn1Up int = 0
    var prevStateExtBtn2Up int = 0
    var StateExtBtn2Up int = 0
    var prevStateExtBtn2Down int = 0
    var StateExtBtn2Down int = 0
    var prevStateExtBtn3Up int = 0
    var StateExtBtn3Up int = 0
    var prevStateExtBtn3Down int = 0
    var StateExtBtn3Down int = 0
    var prevStateExtBtn4Down int = 0
    var StateExtBtn4Down int = 0
    for{
        time.Sleep(time.Millisecond * 10)
        
        StateExtBtn1Up = int(C.io_read_bit(C.FLOOR_UP1))
        if StateExtBtn1Up != prevStateExtBtn1Up{
            prevStateExtBtn1Up = StateExtBtn1Up
            if StateExtBtn1Up == 1{
                ch <- EXT_BTN_1_UP           
            }
        }

        StateExtBtn2Up = int(C.io_read_bit(C.FLOOR_UP2))
        if StateExtBtn2Up != prevStateExtBtn2Up{
            prevStateExtBtn2Up = StateExtBtn2Up
            if StateExtBtn2Up == 1{
                ch <- EXT_BTN_2_UP           
            }
        }

        StateExtBtn2Down = int(C.io_read_bit(C.FLOOR_DOWN2))
        if StateExtBtn2Down != prevStateExtBtn2Down{
            prevStateExtBtn2Down = StateExtBtn2Down
            if StateExtBtn2Down == 1{
                ch <- EXT_BTN_2_DOWN          
            }
        }

        StateExtBtn3Up = int(C.io_read_bit(C.FLOOR_UP3))
        if StateExtBtn3Up != prevStateExtBtn3Up{
            prevStateExtBtn3Up = StateExtBtn3Up
            if StateExtBtn3Up == 1{
                ch <- EXT_BTN_3_UP           
            }
        }

        StateExtBtn3Down = int(C.io_read_bit(C.FLOOR_DOWN3))
        if StateExtBtn3Down != prevStateExtBtn3Down{
            prevStateExtBtn3Down = StateExtBtn3Down
            if StateExtBtn3Down == 1{
                ch <- EXT_BTN_3_DOWN          
            }
        }

        StateExtBtn4Down = int(C.io_read_bit(C.FLOOR_DOWN4))
        if StateExtBtn4Down != prevStateExtBtn4Down{
            prevStateExtBtn4Down = StateExtBtn4Down
            if StateExtBtn4Down == 1{
                ch <- EXT_BTN_4_DOWN          
            }
        }
    }
}

//TODO Ask how can we check if hardware was turned off
func createTimeoutListener(ch chan int) {
    for{    
        ch <- int(C.elev_get_floor_sensor_signal())
        time.Sleep(time.Second * 1)
    }
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

