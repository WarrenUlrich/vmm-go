package main

import (
	"github.com/warrenulrich/vmm-go/pkg/vmm"
)

type Player uintptr

func (p Player) Test() {

}

func main() {
	var err error
	handle, err := vmm.Initialize([]string{
		"",
		"-disable-python",
		"-device",
		"qemu://shm=qemu-win7.mem,qmp=/tmp/qmp-win7.sock",
	})

	if err != nil {
		panic(err)
	}

	vmm.ReadMemSlice[int](handle, 0x00, 0x00, 69, vmm.FLAG_NOCACHE)
}
