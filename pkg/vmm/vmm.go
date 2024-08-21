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

type Handle C.VMM_HANDLE

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
	Name [8]byte
	Misc struct {
		PhysicalAddress uint32
		VirtualSize     uint32
	}
	VirtualAddress       uint32
	SizeOfRawData        uint32
	PointerToRawData     uint32
	PointerToRelocations uint32
	PointerToLinenumbers uint32
	NumberOfRelocations  uint16
	NumberOfLinenumbers  uint16
	Characteristics      uint32
}

type Flag uint64

const (
	FLAG_NOCACHE Flag = C.VMMDLL_FLAG_NOCACHE
)

func Initialize(args []string) (Handle, error) {
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

	return (Handle)(ret), nil
}

func Close(handle Handle) {
	C.VMMDLL_Close(handle)
}

func CloseAll() {
	C.VMMDLL_CloseAll()
}

func GetPidFromName(handle Handle, name string) (uint32, error) {
	var pid C.uint
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	ret := C.VMMDLL_PidGetFromName(handle, cName, &pid)
	if ret == 0 {
		return 0, errors.New("Failed to get PID from name")
	}

	return (uint32)(pid), nil
}

func ReadMem[T any](handle Handle, pid uint32, addr uintptr, flags Flag) (T, error) {
	var result T
	var sizeRead C.DWORD
	var expectedSize C.DWORD = C.DWORD(unsafe.Sizeof(result))

	res := C.VMMDLL_MemReadEx(
		handle,
		C.uint32_t(pid),
		C.ulonglong(addr),
		(*C.BYTE)(unsafe.Pointer(&result)),
		C.uint(unsafe.Sizeof(result)),
		&sizeRead,
		(C.ulonglong)(flags),
	)

	if res == 0 || sizeRead != expectedSize {
		return result, errors.New("failed to read memory")
	}

	return result, nil
}

func ReadMemSlice[T any](handle Handle, pid uint32, addr uintptr, count uint, flags Flag) ([]T, error) {
	var result []T = make([]T, count)
	var sizeRead C.DWORD
	var expectedSize C.DWORD = C.DWORD(unsafe.Sizeof(result[0])) * C.DWORD(count)

	res := C.VMMDLL_MemReadEx(
		handle,
		C.uint32_t(pid),
		C.ulonglong(addr),
		(*C.BYTE)(unsafe.Pointer(&result[0])),
		expectedSize,
		&sizeRead,
		(C.ulonglong)(flags),
	)

	if res == 0 || sizeRead != expectedSize {
		return result, errors.New("failed to read memory")
	}

	return result, nil
}

func GetProcessModuleBase(handle Handle, pid uint32, moduleName string) (uint64, error) {
	var cModuleName *C.char = nil
	if moduleName != "" {
		cModuleName = C.CString(moduleName)
	}

	defer func() {
		if cModuleName != nil {
			C.free(unsafe.Pointer(cModuleName))
		}
	}()

	res := C.VMMDLL_ProcessGetModuleBaseU(handle, C.uint(pid), cModuleName)
	if res == 0 {
		return 0, errors.New("Failed to get module base")
	}

	return uint64(res), nil
}

func GetProcessInfo(handle Handle, pid uint32) (*ProcessInformation, error) {
	var processInfo C.VMMDLL_PROCESS_INFORMATION
	processInfo.magic = C.VMMDLL_PROCESS_INFORMATION_MAGIC
	processInfo.wVersion = C.VMMDLL_PROCESS_INFORMATION_VERSION

	var size C.SIZE_T = C.sizeof_VMMDLL_PROCESS_INFORMATION
	if C.VMMDLL_ProcessGetInformation(handle, C.uint(pid), &processInfo, &size) == 0 {
		return nil, errors.New("Failed to get process information")
	}

	return (*ProcessInformation)(unsafe.Pointer(&processInfo)), nil
}

func GetProcessSections(handle Handle, pid uint32, moduleName string) ([]ImageSectionHeader, error) {
	var cSections C.DWORD
	cModuleName := C.CString(moduleName)
	if C.VMMDLL_ProcessGetSectionsU(
		handle,
		C.uint(pid),
		cModuleName,
		nil,
		0,
		&cSections) == 0 || cSections == 0 {
		return nil, errors.New("Failed to get process sections count")
	}

	sections := make([]ImageSectionHeader, cSections)
	if C.VMMDLL_ProcessGetSectionsU(
		handle,
		C.uint(pid),
		cModuleName,
		(*C.IMAGE_SECTION_HEADER)(unsafe.Pointer(&sections[0])),
		cSections,
		&cSections) == 0 || cSections == 0 {
		return nil, errors.New("Failed to get process sections")
	}

	return sections, nil
}
