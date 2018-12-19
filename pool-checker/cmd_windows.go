// +build windows

package main

import "syscall"

const (
	// CREATE_NEW_PROCESS_GROUP is like Setpgid on UNIX
	// https://msdn.microsoft.com/en-us/library/windows/desktop/ms684863(v=vs.85).aspx
	CREATE_NEW_PROCESS_GROUP = syscall.CREATE_NEW_PROCESS_GROUP
	// DETACHED_PROCESS does not inherit the parent console
	DETACHED_PROCESS = 0x00000008
)

func GetSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: CREATE_NEW_PROCESS_GROUP | DETACHED_PROCESS,
	}
}
