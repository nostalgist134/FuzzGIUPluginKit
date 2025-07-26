package env

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type Env struct {
	BinSuffix string
	OS        string
	BuildMode string
	GoPath    string
	GoVersion string
	OkToBuild bool
}

var binSuffixes = map[string]string{
	"windows": ".dll",
	"linux":   ".so",
	"darwin":  ".dylib",
}

var GlobEnv = Env{}

func Check(golangPath ...string) Env {
	goPath := "go"
	if len(golangPath) > 0 {
		goPath = golangPath[0]
	}
	environ := Env{}
	environ.GoPath = goPath
	// 全局环境
	defer func() { GlobEnv = environ }()
	environ.OS = runtime.GOOS
	// 文件格式
	environ.BinSuffix = binSuffixes[environ.OS]
	// 根据系统检查相应工具链是否完整
	environ.OkToBuild = true
	if environ.OS == "windows" {
		// windows采用cgo库编译
		environ.BuildMode = "-buildmode=c-shared"
		// 尝试获取gcc版本
		gcc := exec.Command("gcc", "--version")
		err := gcc.Run()
		if err != nil {
			cc := exec.Command("cc", "--version")
			err = cc.Run()
			if err != nil {
				fmt.Printf("get gcc failed - %v\n", err)
				environ.OkToBuild = false
			}
		}
	} else { // linux或者macos使用go的plugin包编译
		environ.BuildMode = "-buildmode=plugin"
	}
	// 检查golang环境
	ver, err := exec.Command(goPath, "version").Output()
	if err != nil {
		fmt.Printf("get go binary failed - %v\n", err)
		environ.OkToBuild = false
		return environ
	}
	// 获取golang版本
	environ.GoVersion = strings.Split(string(ver), " ")[2][2:]
	return environ
}

// GetBuildArgs 生成go命令使用的命令行参数
func GetBuildArgs(e Env, out string, goFile string) []string {
	bf := []string{"build", e.BuildMode}
	if e.OS == "windows" {
		bf = append(bf, "-ldflags=-s -w")
	}
	bf = append(bf, "-o", out, goFile)
	return bf
}

func RemoveIntermediateFiles(e Env, out string) {
	// 删除wrapped.go文件
	err := os.Remove("wrapped.go")
	if err != nil {
		fmt.Printf("remove wrapped.go failed: %v\n", err)
	}
	if e.OS != "windows" {
		return
	}
	// 删除.h文件（windows）
	hFile := ""
	if suffix := strings.LastIndex(out, ".dll"); suffix != -1 {
		hFile = out[:suffix] + ".h"
	} else {
		hFile = out + ".h"
	}
	err = os.Remove(hFile)
	if err != nil {
		fmt.Printf("remove %s failed: %v\n", hFile, err)
	}
}

func GetCwd() string {
	cwd, _ := os.Getwd()
	if cwd == "" {
		cwd = "."
	}
	return cwd
}
