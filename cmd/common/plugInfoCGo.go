//go:build windows

package common

import (
	"FuzzGIUPluginKit/convention"
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

// ReadInputLine 从用户输入中读取一行
func ReadInputLine(prompt string, trim ...bool) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	if len(trim) > 0 && trim[0] || len(trim) == 0 {
		return strings.TrimSpace(line)
	}
	return line
}

func stringFromPtr(strBytes uintptr) string {
	sb := strings.Builder{}
	sb.WriteString(unsafe.String((*byte)(unsafe.Pointer(strBytes+4)), *(*int32)(unsafe.Pointer(strBytes))))
	return sb.String()
}

// GetPluginInfo 调用插件的PluginInfo函数并返回
func GetPluginInfo(pluginFile string) (*convention.PluginInfo, error) {
	dll, err := syscall.LoadDLL(pluginFile)
	if err != nil {
		return nil, err
	}
	pi, err := dll.FindProc("PluginInfo")
	if err != nil {
		return nil, err
	}
	ret, _, err := pi.Call()
	var errno syscall.Errno
	if err != nil && (!errors.As(err, &errno) || errno != 0) {
		return nil, err
	}
	s := stringFromPtr(ret)
	jsonBytes := []byte(s)
	pInfo := new(convention.PluginInfo)
	err = json.Unmarshal(jsonBytes, pInfo)
	return pInfo, err
}
