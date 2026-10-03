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


type FILE_DISPOSITION_INFO_EX struct {
	Flags uint32
}

// INTERNAL: Opens handle to target file
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

// INTERNAL: Changes target file's DataStream to ":deadbeef"
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

// INTERNAL: Removes file by applying "POSIX" file treating on windows FS
func dsDepositeHandle(hHandle windows.Handle) error {
	var fdiEx FILE_DISPOSITION_INFO_EX
	fdiEx.Flags = windows.FILE_DISPOSITION_DELETE | windows.FILE_DISPOSITION_POSIX_SEMANTICS
	var ioStatusBlock2 windows.IO_STATUS_BLOCK

	if err := windows.NtSetInformationFile(hHandle,&ioStatusBlock2,(*byte)(unsafe.Pointer(&fdiEx)),
	uint32(unsafe.Sizeof(fdiEx)),windows.FileDispositionInformationEx); err != nil{
		return err
	}

	return nil
}

// PUBLIC: Full windows "active" EXE self-removal
func SelfDeleteExe() error{
	// Get full executable path via GetModuleFileName windows API
	var wcPath [windows.MAX_PATH + 1]uint16
	var hCurrent windows.Handle

	_, err := windows.GetModuleFileName(0, &wcPath[0], windows.MAX_PATH)
	if err != nil {
		return fmt.Errorf("SelfDeleteExe error: Get executable fullPath error: %v",err)
	}

	// Open handle to exe file
	hCurrent, err = dsOpenHandle(&wcPath[0])
	if err != nil || hCurrent == windows.InvalidHandle {
		return fmt.Errorf("SelfDeleteExe error: Open exe file handle error: %v",err)
	}
	
	// Move all exe contents to a windows alternate DATA STREAM
	err = dsRenameHandle(hCurrent)
	if err != nil {
		windows.CloseHandle(hCurrent)
		return fmt.Errorf("SelfDeleteExe error: DataStream change error: %v",err)
	}

	// Close handle to exe file
	windows.CloseHandle(hCurrent)

	// Reopen exe file
	hCurrent, err = dsOpenHandle(&wcPath[0])
	if err != nil || hCurrent == windows.InvalidHandle {
		return fmt.Errorf("SelfDeleteExe error: Reopen exe file handle error: %v",err)
	}

	// Apply POSIX file treatment and DELETE flags to handle to allow self-removal when closing the handle
	err = dsDepositeHandle(hCurrent)
	if err != nil {
		windows.CloseHandle(hCurrent)
		return fmt.Errorf("SelfDeleteExe error: Apply DELETE and POSIX_SEMANTICS to handle error: %v",err)
	}

	// Close handle, effectively "tricking" windows to finally remove our exe file on disk (while still being loaded on mem)
	windows.CloseHandle(hCurrent)

	return nil
}
