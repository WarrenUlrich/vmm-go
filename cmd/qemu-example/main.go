package main

import (
	"fmt"

	"github.com/warrenulrich/vmm-go/pkg/vmm"
)

func main() {
	vm, err := vmm.Initialize([]string{
		"",
		"-disable-python",
		"-device",
		"qemu://shm=qemu-win7.mem,qmp=/tmp/qmp-win7.sock",
	})

	if err != nil {
		panic(err)
	}

	defer vm.Close()

	pid, err := vm.GetPidFromName("explorer.exe")
	if err != nil {
		panic(err)
	}

	fmt.Println("PID:", pid)

	base, err := vm.GetProcessModuleBase(pid, "")
	if err != nil {
		panic(err)
	}

	data, err := vm.ReadMem(pid, base, 2) // Should be 'MZ'
	if err != nil {
		panic(err)
	}

	fmt.Printf("Data: %s\n", data)

	sections, err := vm.GetProcessSections(pid, "")
	if err != nil {
		panic(err)
	}

	for _, section := range sections {
		fmt.Printf("%s\n", section.Name)
	}
}
