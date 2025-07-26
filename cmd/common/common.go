package common

import (
	"fmt"
	"os"
)

var subCmd = ""
var exitDefer = func() {}

func SetCurrentCmd(sub string) {
	subCmd = sub
}

func SetExitDefer(f func()) {
	exitDefer = f
}

func ClearExitDefer() {
	exitDefer = func() {}
}

// FailExit 接收错误信息或错误类型，如果接收错误信息则退出，如果接收到nil则直接返回，不退出（这么改之后就能少写几个panic了）
func FailExit(reason any, code ...int) {
	if reason == nil {
		return
	}
	exitCode := 1
	if len(code) > 0 {
		exitCode = code[0]
	}
	fmt.Printf("%s execution failed, reason: %s\nnow exitting...\n", subCmd, reason)
	exitDefer()
	os.Exit(exitCode)
}
