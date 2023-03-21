package sys

import (
	"golang.org/x/sys/windows/registry"
	"syscall"
)

var k32 = syscall.NewLazyDLL("kernel32.dll")

func GetOs() string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return "Unknown"
	}
	defer k.Close()
	pn, _, err := k.GetStringValue("ProductName")
	if err != nil {
		return "Unknown"
	}
	return pn
}
