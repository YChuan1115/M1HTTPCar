/**********************************************
 * NanoPi M1 HTTP Car
 * author: bluebanboom
 * email:  bluebanboom@gmail.com
 * 2016.07.21
 * https://github.com/bluebanboom/M1HttpCar
 **********************************************/
package main

import (
    "fmt"
    "log"
    "net/http"
    "encoding/json"
    "html/template"
    "net"
    "time"
    "./gpio"
)

const LeftIn1  = 11
const LeftIn2  = 13
const RightIn1 = 15
const RightIn2 = 16

var gCar *Car = nil

var Actions = map[string] [4]int {
    "forward" : {1, 0, 1, 0},
    "backward": {0, 1, 0, 1},
    "left"    : {0, 0, 1, 0},
    "right"   : {1, 0, 0, 0},
    "stop"    : {0, 0, 0, 0}}

type Car struct {
    Pins [4]int
    Gpio *gpio.GPIO
}

func NewCar() *Car {
    car := new(Car)
    car.Pins = [4]int{LeftIn1, LeftIn2, RightIn1, RightIn2}
    return car
}

func (car *Car)On()  {
    car.Gpio = new(gpio.GPIO)
    err := car.Gpio.Setup()
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println("-> Car on.")
    for _, pin := range car.Pins {
        car.Gpio.PinMode(pin, gpio.OUTPUT)
        car.Gpio.PullUpDnControl(pin, gpio.PUD_DOWN)
    }
}

func (car *Car)Off() {
    fmt.Println("-> Car off.")
    car.Gpio.Cleanup()
    car.Gpio = nil
}

func (car *Car)DoAction(action string) int {
    output, ok := Actions[action]
    if ok {
        for i, val := range output {
            pin := car.Pins[i]
            fmt.Printf("pin[%d] = %d\n", pin, val)
            car.Gpio.DigitalWrite(pin, val)
            time.Sleep(10 * time.Millisecond)
        }
        fmt.Println("-------------------------------")
        return 1
    }
    return 0
}

func GetLocalIP() string {
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        return ""
    }
    for _, address := range addrs {
        // check the address type and if it is not a loopback the display it
        if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() != nil {
                return ipnet.IP.String()
            }
        }
    }
    return ""
}

func main() {
    ip := GetLocalIP()
    if len(ip) == 0 {
        ip = "localhost"
    }
    fmt.Println("-------------------------------------------")
    fmt.Printf("Car is ready on http://%s:8000\n", ip)
    fmt.Println("-------------------------------------------")
    gCar = NewCar()

    // static
    http.Handle("/css/", http.FileServer(http.Dir("template")))
    http.Handle("/js/", http.FileServer(http.Dir("template")))
    http.Handle("/fonts/", http.FileServer(http.Dir("template")))

    http.HandleFunc("/", handler)
    http.HandleFunc("/action", action)
    log.Fatal(http.ListenAndServe(ip + ":8000", nil))
}

func handler(w http.ResponseWriter, r *http.Request)  {
    t, err := template.ParseFiles("template/html/index.html")
    if (err != nil) {
        log.Println(err)
    }
    t.Execute(w, nil)
}

func action(w http.ResponseWriter, r *http.Request)  {
    w.Header().Set("content-type", "application/json")
    action := r.URL.RawQuery
    fmt.Println("action: ", action)
    actionSuccess := 1
    if action == "on" {
        gCar.On()
    } else if action == "off" {
        gCar.DoAction("stop")
        gCar.Off()
    } else {
        if action != "stop" {
            gCar.DoAction("stop")
        }
        actionSuccess = gCar.DoAction(action)
    }
    data := map[string] int {"result" : actionSuccess}
    result, _ := json.Marshal(data)
    w.Write(result)
}
