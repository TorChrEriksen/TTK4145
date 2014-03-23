package main

import (
    "fmt"
    "strings"
)

func main() {
    s1 := "100.200.300.400"
    s11 := "100.200.30.40"
    fmt.Println(len(s1))

    fmt.Println(strings.LastIndex(s1, "."))
    fmt.Println(strings.LastIndex(s11, "."))

    s2 := s1[strings.LastIndex(s1, ".") + 1:]
    s3 := s11[strings.LastIndex(s11, ".") + 1:]
    fmt.Println(s2)
    fmt.Println(s3)

}
