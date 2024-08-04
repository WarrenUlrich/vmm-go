package vmm

/*
#cgo CFLAGS: -I${SRCDIR}/include
#cgo LDFLAGS: -L${SRCDIR} -ldl -lpthread ${SRCDIR}/lib/vmm.so ${SRCDIR}/lib/leechcore.so ${SRCDIR}/lib/leechcore_device_qemu.so

#define LINUX 1

#include <leechcore.h>
#include <vmmdll.h>
*/
import (
	"C"
)

import (
	"errors"
	"unsafe"
)

type Handle struct {
	cHandle C.VMM_HANDLE
}

type ProcessWinInfo struct {
	VaEPROCESS     uint64
	VaPEB          uint64
	Reserved1      uint64
	FWow64         bool
	VaPEB32        uint32
	DwSessionId    uint32
	QwLUID         uint64
	SzSID          [260]byte
	IntegrityLevel int32
}

type ProcessInformation struct {
	Magic         uint64
	WVersion      uint16
	WSize         uint16
	TpMemoryModel int32
	TpSystem      int32
	FUserOnly     bool
	DwPID         uint32
	DwPPID        uint32
	DwState       uint32
	SzName        [16]byte
	SzNameLong    [64]byte
	PaDTB         uint64
	PaDTB_UserOpt uint64
	Win           ProcessWinInfo
}

type ImageSectionHeader struct {
	Name                 [8]byte
	Misc                 [4]byte
	VirtualAddress       uint32
	SizeOfRawData        uint32
	PointerToRawData     uint32
	PointerToRelocations uint32
	PointerToLinenumbers uint32
	NumberOfRelocations  uint16
	NumberOfLinenumbers  uint16
	Characteristics      uint32
}

func Initialize(args []string) (*Handle, error) {
	cArgs := []*C.char{}
	for _, arg := range args {
		cArgs = append(cArgs, C.CString(arg))
	}

	defer func() {
		for _, arg := range cArgs {
			C.free(unsafe.Pointer(arg))
		}
	}()

	ret := C.VMMDLL_Initialize(C.uint(len(args)), (*C.LPCSTR)(unsafe.Pointer(&cArgs[0])))
	if ret == nil {
		return nil, errors.New("Failed to initialize VMMDLL")
	}

	return &Handle{cHandle: ret}, nil
}

func CloseAll() {
	C.VMMDLL_CloseAll()
}

func (handle *Handle) Close() {
	C.VMMDLL_Close(handle.cHandle)
}

func (handle *Handle) GetPidFromName(name string) (uint32, error) {
	var pid C.uint
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	ret := C.VMMDLL_PidGetFromName(handle.cHandle, cName, &pid)
	if ret == 0 {
		return 0, errors.New("Failed to get PID from name")
	}

	return uint32(pid), nil
}

func (handle *Handle) ReadMem(pid uint32, addr uint64, size uint) ([]byte, error) {
	buffer := make([]byte, size)
	res := C.VMMDLL_MemRead(handle.cHandle, C.uint32_t(pid), C.ulonglong(addr), (*C.BYTE)(unsafe.Pointer(&buffer[0])), C.uint(size))
	if res == 0 {
		return nil, errors.New("Failed to read memory")
	}

	return buffer, nil
}

func (handle *Handle) WriteMem(pid uint32, addr uint64, data []byte) error {
	res := C.VMMDLL_MemWrite(handle.cHandle, C.uint32_t(pid), C.ulonglong(addr), (*C.BYTE)(unsafe.Pointer(&data[0])), C.uint(len(data)))
	if res == 0 {
		return errors.New("Failed to write memory")
	}

	return nil
}

func (handle *Handle) GetProcessModuleBase(pid uint32, moduleName string) (uint64, error) {
	var cModuleName *C.char = nil
	if moduleName != "" {
		cModuleName = C.CString(moduleName)
	}

	defer func() {
		if cModuleName != nil {
			C.free(unsafe.Pointer(cModuleName))
		}
	}()

	res := C.VMMDLL_ProcessGetModuleBaseU(handle.cHandle, C.uint(pid), cModuleName)
	if res == 0 {
		return 0, errors.New("Failed to get module base")
	}

	return uint64(res), nil
}

func (handle *Handle) GetProcessInfo(pid uint32) (*ProcessInformation, error) {
	var processInfo C.VMMDLL_PROCESS_INFORMATION
	processInfo.magic = C.VMMDLL_PROCESS_INFORMATION_MAGIC
	processInfo.wVersion = C.VMMDLL_PROCESS_INFORMATION_VERSION

	var size C.SIZE_T = C.sizeof_VMMDLL_PROCESS_INFORMATION
	if C.VMMDLL_ProcessGetInformation(handle.cHandle, C.uint(pid), &processInfo, &size) == 0 {
		return nil, errors.New("Failed to get process information")
	}

	return (*ProcessInformation)(unsafe.Pointer(&processInfo)), nil
}

func (handle *Handle) GetProcessSections(pid uint32, moduleName string) ([]ImageSectionHeader, error) {
	var cSections C.DWORD
	cModuleName := C.CString(moduleName)
	if C.VMMDLL_ProcessGetSectionsU(
		handle.cHandle,
		C.uint(pid),
		cModuleName,
		nil,
		0,
		&cSections) == 0 || cSections == 0 {
		return nil, errors.New("Failed to get process sections count")
	}

	sections := make([]ImageSectionHeader, cSections)
	if C.VMMDLL_ProcessGetSectionsU(
		handle.cHandle,
		C.uint(pid),
		cModuleName,
		(*C.IMAGE_SECTION_HEADER)(unsafe.Pointer(&sections[0])),
		cSections,
		&cSections) == 0 || cSections == 0 {
		return nil, errors.New("Failed to get process sections")
	}

	return sections, nil
}
