package main

import (
    "os/exec"
    "fmt"
    "strconv"
    "os"

)

func main() {
    pwd, err := os.Getwd()
    if err != nil {
        fmt.Println("Error: ", err.Error())
        return
    }

    app := "sh"
    arg0 := pwd + "/test.sh"

    for i := 0; i < 10; i++ {
        arg1 := strconv.Itoa(i)
        cmd := exec.Command(app, arg0, arg1)
        _, err := cmd.Output()
        if err != nil {
            fmt.Println("Error_2: ", err.Error())
        }
        fmt.Println(i)
        //fmt.Println(string(out))
    }
}
