// Elevator control logic

package main

import(
    "./lib"
    "fmt"
	"time"
    //"os"
)

func main(){
//    driver.Test_Run()

	// Create driver with all the channels we need
	buttonChannel := make(chan int)
	driverInterface.Create(buttonChannel)
	
	for {
		select {
			case buttonSignal := <-buttonChannel :
				// release thread (use a channel or fire a go routine)
				go func() {
					fmt.Println(buttonSignal)
					time.Sleep(time.Second * 2) // Just for demonstration why we want to release the thread!
				}()

			// more cases
//			case floorSignal := <-floorChannel :

		}
	}
}