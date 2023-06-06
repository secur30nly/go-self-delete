/*
    License: MIT Licence

    References:
        - https://github.com/LloydLabs/delete-self-poc
        - https://twitter.com/jonasLyk/status/1350401461985955840
*/


package selfdelete

import (
	"fmt"
	"unsafe"
	"golang.org/x/sys/windows"
)


type FILE_RENAME_INFO struct {
	Union struct {
		ReplaceIfExists bool
		Flags uint32
	}
	RootDirectory windows.Handle
	FileNameLength uint32
	FileName [1]uint16
}


type FILE_DISPOSITION_INFO struct {
	DeleteFile bool
}


func dsOpenHandle(pwPath *uint16) (windows.Handle, error) {
	handle, err := windows.CreateFile(
		pwPath, 
		windows.DELETE, 
		0, 
		nil,
		windows.OPEN_EXISTING, 
		windows.FILE_ATTRIBUTE_NORMAL, 
		0,
	)
	
	if err != nil {
		return 0, err
	}

	return handle, nil
}


func dsRenameHandle(hHandle windows.Handle) error {
	var fRename FILE_RENAME_INFO
	DS_STREAM_RENAME, err := windows.UTF16FromString(":deadbeef")
	
	if err != nil {
		return err
	}

	lpwStream := &DS_STREAM_RENAME[0]
	fRename.FileNameLength = uint32(unsafe.Sizeof(lpwStream))
	
	windows.NewLazyDLL("kernel32.dll").NewProc("RtlCopyMemory").Call(
		uintptr(unsafe.Pointer(&fRename.FileName[0])), 
		uintptr(unsafe.Pointer(lpwStream)), 
		unsafe.Sizeof(lpwStream),
	)

	err = windows.SetFileInformationByHandle(
		hHandle, 
		windows.FileRenameInfo, 
		(*byte)(unsafe.Pointer(&fRename)), 
		uint32(unsafe.Sizeof(fRename) + unsafe.Sizeof(lpwStream)),
	)
	
	if err != nil {
		return err
	}

	return nil
}


func dsDepositeHandle(hHandle windows.Handle) error {
	var fDelete FILE_DISPOSITION_INFO
	fDelete.DeleteFile = true

	err := windows.SetFileInformationByHandle(
		hHandle, 
		windows.FileDispositionInfo, 
		(*byte)(unsafe.Pointer(&fDelete)), 
		uint32(unsafe.Sizeof(fDelete)),
	)
	
	if err != nil {
		return err
	}

	return nil
}


func SelfDeleteExe() error {
	var wcPath [windows.MAX_PATH + 1]uint16
	var hCurrent windows.Handle

	_, err := windows.GetModuleFileName(0, &wcPath[0], windows.MAX_PATH)
	if err != nil {
		return err
	}

	hCurrent, err = dsOpenHandle(&wcPath[0])
	if err != nil || hCurrent == windows.InvalidHandle {
		return err
	}
	
	err = dsRenameHandle(hCurrent)
	if err != nil {
		windows.CloseHandle(hCurrent)
		return err
	}

	windows.CloseHandle(hCurrent)

	hCurrent, err = dsOpenHandle(&wcPath[0])
	if err != nil || hCurrent == windows.InvalidHandle {
		return err
	}

	err = dsDepositeHandle(hCurrent)
	if err != nil {
		windows.CloseHandle(hCurrent)
		return err
	}

	windows.CloseHandle(hCurrent)

	fmt.Println("Self deletion is successful. Program still running, hit key to continue")
	str := ""
	fmt.Scanln(&str)

	return nil
}
