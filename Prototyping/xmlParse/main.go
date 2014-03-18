package main

import (
    "encoding/xml"
    "fmt"
    "io"
    "os"
    "path/filepath"
)

/*
type XMLStrap struct {
    XMLName xml.Name `xml:"config"`
    CatchInterrupt bool `xml:"CatchInterrupt,attr"`
    Redundant bool `xml:"Redundant,attr"`
    PortTCP int `xml:"PortTCP,attr"`
    PortUDP int `xml:"PortUDP,attr"`
    PortBroadcast int `xml:"PortBroadcast,attr"`
    DebugMode bool `xml:"DebugMode,attr"`
}
*/

type XMLStrap struct {
    XMLName xml.Name `xml:"config"`
    Key string `xml:"key,attr"`
    Value string `xml:"value,attr"`
}

type XMLStraps struct {
    XMLName xml.Name `xml:"appcnf"`
    Conf []*XMLStrap `xml:"config"`
}

func ReadStraps(reader io.Reader) ([]*XMLStrap, error) {
    xmlStraps := &XMLStraps{}
    decoder := xml.NewDecoder(reader)

    if err := decoder.Decode(xmlStraps); err != nil {
        return nil, err
    }

    return xmlStraps.Conf, nil
}

func main() {
    var xmlStraps []*XMLStrap
    var file *os.File

    defer func() {
        if file != nil {
            file.Close()
        }
    }()

    // Build the location of the straps.xml file
    // filepath.Abs appends the file name to the default working directly
    strapsFilePath, err := filepath.Abs("strapdir/straps.xml")

    if err != nil {
        panic(err.Error())
    }

    // Open the straps.xml file
    file, err = os.Open(strapsFilePath)

    if err != nil {
        panic(err.Error())
    }

    // Read the straps file
    xmlStraps, err = ReadStraps(file)

    if err != nil {
        panic(err.Error())
    }

    // Display the first strap
    for _, element := range xmlStraps {
        fmt.Println(element.Key, ": ", element.Value)
    }
}
