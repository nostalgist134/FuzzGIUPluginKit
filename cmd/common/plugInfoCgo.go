//go:build windows

package common

import (
	"FuzzGIUPluginKit/convention"
	"encoding/json"
	"errors"
	"syscall"
	"unsafe"
)

// PluginInfo 调用插件的PluginInfo函数并返回
func PluginInfo(pluginFile string) (*convention.PluginInfo, error) {
	dll, err := syscall.LoadDLL(pluginFile)
	if err != nil {
		return nil, err
	}
	defer syscall.FreeLibrary(dll.Handle)
	pi, err := dll.FindProc("PluginInfo")
	if err != nil {
		return nil, err
	}
	ret, _, err := pi.Call()
	var errno syscall.Errno
	if err != nil && (!errors.As(err, &errno) || errno != 0) {
		return nil, err
	}
	s := (*string)(unsafe.Pointer(ret))
	jsonBytes := []byte(*s)
	pInfo := new(convention.PluginInfo)
	err = json.Unmarshal(jsonBytes, pInfo)
	return pInfo, err
}
