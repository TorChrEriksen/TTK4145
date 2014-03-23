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
    "math"
)
//Number of floors
const FLOORS int = 4

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
func SetButtonLamp(direction string,floor int, value int){
  if direction == "UP"{
      //fmt.Println("goin' up") //Debug
      C.elev_set_button_lamp(C.BUTTON_CALL_UP, C.int(floor), C.int(value))
  }else if direction == "DOWN"{
      //fmt.Println("goin' down") //Debug
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

func Create(intBtChan chan int, floorChan chan int, stopChan chan int, extBtChan chan int) {
go createIntButtonListener(intBtChan)
    go createFloorListener(floorChan)
    go createStopListener(stopChan)
    go createExtButtonListener(extBtChan)
}

const INT_BTN_BASE int = 11
func createIntButtonListener(ch chan int) {
var FloorCommands = [FLOORS]C.int{C.FLOOR_COMMAND1, C.FLOOR_COMMAND2, C.FLOOR_COMMAND3, C.FLOOR_COMMAND4}
    //TODO think about fault tolerance here
    var prevStateIntBtn [FLOORS]int
    var StateIntBtn [FLOORS]int
    for i := range StateIntBtn{
     prevStateIntBtn[i] = 0
     StateIntBtn[i] = 0
    }

    for{
     time.Sleep(time.Millisecond * 10)	
     for i := range StateIntBtn{
     StateIntBtn[i] = int(C.io_read_bit(FloorCommands[i]))
     if StateIntBtn[i] != prevStateIntBtn[i]{
     prevStateIntBtn[i] = StateIntBtn[i]
     if StateIntBtn[i] == 1{
     ch <- INT_BTN_BASE + i
     }
     }
     }
    }
}

const FLOOR_NO_BASE int = 31
func createFloorListener(ch chan int){
    var prevStateFloor int = -2
    var StateFloor int = 0
    var FloorOutput [FLOORS]int
    for i := range FloorOutput{
     FloorOutput[i] = i
    }
    
    for{
    time.Sleep(time.Millisecond * 10)
    StateFloor = int(C.elev_get_floor_sensor_signal())
if StateFloor != prevStateFloor{
prevStateFloor = StateFloor
for i:= 0; i<FLOORS; i++{
if i==0 && StateFloor == -1{
ch <- FLOOR_NO_BASE - 1
}
if StateFloor == FloorOutput[i]{
ch <- FLOOR_NO_BASE +i
}
}
}
}
}

const STOP_BASE int = 10
func createStopListener(ch chan int){
   var prevStateStop int = 0
   var StateStop int = 0
   
   for{
        time.Sleep(time.Millisecond * 10)
        StateStop = int(C.elev_get_stop_signal())
        if StateStop != prevStateStop{
            prevStateStop = StateStop
            if StateStop == 1{
                ch <- STOP_BASE
            }
        }
    }
}

const EXT_BTN_BASE int = 40
func createExtButtonListener(ch chan int){
var FloorChans [FLOORS*2]int
    var FloorCommands = [FLOORS*2]C.int{0, C.FLOOR_UP1, C.FLOOR_DOWN2, C.FLOOR_UP2, C.FLOOR_DOWN3, C.FLOOR_UP3, C.FLOOR_DOWN4, 0}
    var prevStateExtBtn [FLOORS*2]int
    var StateExtBtn [FLOORS*2]int
    step:= -10
    for i := range StateExtBtn{
     prevStateExtBtn[i] = 0
     StateExtBtn[i] = 0
     if math.Mod(float64(i),2)==0{
     step+=10
     FloorChans[i] = EXT_BTN_BASE+step
     }else{
     FloorChans[i] = EXT_BTN_BASE+step+1
     }
    }
    
    for{
     time.Sleep(time.Millisecond * 10)
for i := range StateExtBtn{
     StateExtBtn[i] = int(C.io_read_bit(FloorCommands[i]))
     if StateExtBtn[i] != prevStateExtBtn[i]{
     prevStateExtBtn[i] = StateExtBtn[i]
     if StateExtBtn[i] == 1{
     ch <- FloorChans[i]
     }
     }
     }
    }
}
