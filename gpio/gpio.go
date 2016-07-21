package gpio

import (
	// "errors"
    // "fmt"
    "unsafe"
    "os"
    "syscall"
)

const M1_GPIO_BASE = 0x01C20800

const INPUT    = 0
const OUTPUT   = 1

const LOW      = 0
const HIGH     = 1

const PUD_OFF  = 0
const PUD_UP   = 1
const PUD_DOWN = 2

type Port struct {
    Pn_CFG	[4]uint32
    Pn_DAT	uint32
    Pn_DRV	[2]uint32
    Pn_PUL	[2]uint32
}

type GPIO struct {
    base []byte
    file *os.File
}

var gPinToPort[64]int = [64]int{
//   A       C   D   E   F   G
//   0   1   2   3   4   5   6   7   8   9
    -1, -1, -1, -1, -1, -1, -1,  6,  6, -1,     // 0
     6,  0,  0,  0, -1,  0,  6, -1,  6,  2,     // 1
    -1,  2,  0,  2,  2, -1,  0,  0,  0,  0,     // 2
    -1,  0,  0,  0, -1,  0,  0,  0,  0, -1,     // 3
     0, -1, -1, -1, -1, -1, -1, -1, -1, -1,     // 4
    -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,     // 5
    -1, -1, -1, -1}

var gPinToShift[64]int = [64]int {
//   0   1   2   3   4   5   6   7   8   9
    -1, -1, -1, -1, -1, -1, -1, 11,  6, -1,     // 0
     7,  0,  6,  2, -1,  3,  8, -1,  9,  0,     // 1
    -1,  1,  1, 29,  3, -1, 17, 19, 18, 20,     // 2
    -1, 21,  7,  8, -1, 16, 13,  9, 15, -1,     // 3
    14, -1, -1, -1, -1, -1, -1, -1, -1, -1,     // 4
    -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,     // 5
    -1, -1, -1, -1}

func PinToPort(pin int) int  {
    if pin < 0 || pin >= 64 {
        return -1
    } else {
        return gPinToPort[pin]
    }
}

func PinToShift(pin int) int {
    if pin < 0 || pin >= 64 {
        return -1
    } else {
        return gPinToShift[pin]
    }
}

func (port *Port) PinMode(pin int, mode uint32) {
    portn := PinToPort(pin)
    shift := PinToShift(pin)
    if portn < 0 || shift < 0 {
        return
    }

    n := uint32(shift / 8)
    s := uint32(shift % 8 * 4)
    mask := uint32(^(0xF << s))
    value := port.Pn_CFG[n] & mask
    value = value | (mode << s)
    port.Pn_CFG[n] = value
}

func (port *Port) PullUpDnControl(pin int, pud uint32) {
    portn := PinToPort(pin)
    shift := PinToShift(pin)
    if portn < 0 || shift < 0 {
        return
    }

    n := uint32(shift / 16)
    s := uint32(shift % 16 * 2)
    mask := uint32(^(0x3 << s))
    value := port.Pn_PUL[n] & mask
    port.Pn_PUL[n] = value | (pud << s)
}

func (port *Port) DigitalWrite(pin int, value int) {
    portn := PinToPort(pin)
    shift := PinToShift(pin)
    if portn < 0 || shift < 0 {
        return
    }

    if value == 1 {
        port.Pn_DAT = port.Pn_DAT | (1 << uint32(shift))
    } else {
        port.Pn_DAT = port.Pn_DAT & (^(1 << uint32(shift)))
    }
}

func (port *Port) DigitalRead(pin int, value int) int {
    portn := PinToPort(pin)
    shift := PinToShift(pin)
    if portn < 0 || shift < 0 {
        return -1
    }

    data := int(port.Pn_DAT)
    return (data & (^(1 << uint32(shift)))) >> uint32(shift)
}

func (gpio *GPIO) Setup() error {
    f, err := os.OpenFile("/dev/mem", os.O_RDWR, 0)
    if err != nil {
        return err
    }

    gpio.file = f

    mem, err := syscall.Mmap(int(f.Fd()), int64(M1_GPIO_BASE & 0xFFFFF000),
		os.Getpagesize(), syscall.PROT_READ | syscall.PROT_WRITE, syscall.MAP_SHARED)

    if err != nil {
        return err
    }

    gpio.base = mem
    return nil
}

func (gpio *GPIO) Cleanup() {
    if gpio.file != nil {
        gpio.file.Close()
    }
    if gpio.base != nil {
        syscall.Munmap(gpio.base)
    }
}

func (gpio *GPIO) portForPin(pin int) *Port {
    portn := PinToPort(pin)
    if portn < 0 {
        return nil
    }
    offset := M1_GPIO_BASE & 0x0000FFFF
    return (*Port)(unsafe.Pointer(&gpio.base[offset + portn * 0x24]))
}

func (gpio *GPIO) PinMode(pin int, mode uint32) {
    port := gpio.portForPin(pin)
    port.PinMode(pin, mode)
}

func (gpio *GPIO) PullUpDnControl(pin int, pud uint32) {
    port := gpio.portForPin(pin)
    port.PullUpDnControl(pin, pud)
}

func (gpio *GPIO) DigitalWrite(pin int, value int) {
    port := gpio.portForPin(pin)
    port.DigitalWrite(pin, value)
}

func (gpio *GPIO) DigitalRead(pin int, value int) int {
    port := gpio.portForPin(pin)
    return port.DigitalRead(pin, value)
}

// func main() {
//     p := PinToPort(7)
//     fmt.Println("port = ", p)
//
//     s := PinToShift(77)
//     fmt.Println("shift = ", s)
//
//     buffer := [1024]byte{}
//     buffer[0x24] = 0x77
//     buffer[0x25] = 0x77
//     buffer[0x26] = 0x77
//     buffer[0x27] = 0x77
//     // pp := Port{[4]uint32{0x77777777, 0x77777777, 0x77777777, 0x77777777}, 0, [2]uint32{0x55555555, 0x00000555}, [2]uint32{0, 0}}
//     pp := (*Port)(unsafe.Pointer(&buffer[0x24]))
//     fmt.Println(unsafe.Sizeof(pp))
//     fmt.Println(unsafe.Alignof(pp))
//     fmt.Println(unsafe.Offsetof(pp.Pn_DAT))
//
//     fmt.Printf("0x%08X\n", pp.Pn_CFG[0])
//     pp.PinMode(11, 0)
//     fmt.Printf("0x%08X\n", pp.Pn_CFG[0])
//     fmt.Printf("0x%08X\n", pp.Pn_CFG[1])
// }
