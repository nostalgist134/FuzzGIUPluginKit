package gen

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIUPluginKit/cmd/common"
	"github.com/nostalgist134/FuzzGIUPluginKit/convention"
	"github.com/nostalgist134/FuzzGIUPluginKit/env"
	"github.com/nostalgist134/FuzzGIUPluginKit/tmpl"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var Cmd = &cobra.Command{
	Use:   "gen",
	Short: "generate plugin development go project",
	Run:   runCmdGen,
}

func init() {
	Cmd.Flags().StringP("type", "t", "", "plugin type to generate project. currently "+
		fmt.Sprintf("support: \n\t%s", convention.PluginTypes))
	Cmd.Flags().StringP("dir", "d", "", "directory to generate project(auto mkdir)")
	Cmd.Flags().BoolP("no-net", "n", false, "does not get fuzzTypes.go from net")
}

func getContentHttp(url string) ([]byte, error) {
	// 发起 HTTP GET 请求
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP stat error: %s", resp.Status)
	}

	// 读取整个响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	// 转换为字符串返回
	return bodyBytes, nil
}

// splitExistPath 拆分路径为存在部分和不存在部分
func splitExistPath(fullPath string) (string, string, error) {
	fullPath = filepath.Clean(fullPath)
	parts := strings.Split(fullPath, string(os.PathSeparator))

	if filepath.IsAbs(fullPath) {
		// 保留根路径前缀（Unix: "/", Windows: "C:\")
		parts[0] += string(os.PathSeparator)
	}

	var existParts []string
	for i := 1; i <= len(parts); i++ {
		current := filepath.Join(parts[:i]...)
		if _, err := os.Stat(current); err == nil {
			existParts = parts[:i]
		} else {
			break
		}
	}

	exist := filepath.Join(existParts...)
	notExist := strings.TrimPrefix(fullPath, exist)
	notExist = strings.TrimPrefix(notExist, string(os.PathSeparator))
	exist, _ = filepath.Abs(exist)
	return exist, notExist, nil
}

var filesCreated = make([]*os.File, 0)

func addHelpers(baseDir string, moduleName string) error {
	helperDir := filepath.Join(baseDir, "helper")
	err := os.Mkdir(helperDir, 0755)
	if err != nil {
		return err
	}
	files := tmpl.GetTemplatesDir("helper")
	// 创建并写入helper文件
	for _, f := range files {
		helperName := strings.TrimSuffix(filepath.Join(helperDir, f.Name), "tmp")
		fPtr, err := os.Create(helperName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create '%s': %v\n", helperName, err)
		}
		contentStr := string(f.Content)
		contentStr = tmpl.Replace(contentStr, tmpl.PHModuleName, moduleName)
		_, err = fPtr.WriteString(contentStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to write to '%s': %v\n", helperName, err)
		}
	}
	return nil
}

func addFuzzType(baseDir string, noNet bool, moduleName string) error {
	var err error

	// 创建fuzzTypes目录
	fuzzTypesDir := filepath.Join(baseDir, "fuzzTypes")
	err = os.MkdirAll(fuzzTypesDir, 0755)
	common.FailExit(err)

	templateFiles := tmpl.GetTemplatesDir("fuzzTypes")
	fileNames := make([]string, len(templateFiles))
	for i, f := range templateFiles {
		fileNames[i] = strings.TrimSuffix(filepath.Join(fuzzTypesDir, f.Name), "tmp")
	}

	contentUrls := []string{
		"https://raw.githubusercontent.com/nostalgist134/FuzzGIU/main/components/fuzzTypes/fuzzTypes.go",
		"https://raw.githubusercontent.com/nostalgist134/FuzzGIU/main/components/fuzzTypes/receivers.go",
	}

	for i, fileName := range fileNames {
		fmt.Print(fileName + ": ")
		var f *os.File
		f, err = os.Create(fileName)
		if err != nil {
			return err
		}
		filesCreated = append(filesCreated, f)
		var content []byte
		if !noNet { // 尝试从github上拉取
			fmt.Print("from net...")
			content, err = getContentHttp(contentUrls[i])
			if err != nil {
				fmt.Println("failed - ", err)
			} else {
				fmt.Println("success")
			}
		}
		if err != nil || noNet { // github拉取失败或指定了从本地获取
			fmt.Println("from embedded")
			content = templateFiles[i].Content
		}
		content = tmpl.ReplaceBytes(content, "github.com/nostalgist134/FuzzGIU/components/common",
			"/* MODULE_NAME *//components/helper")
		content = tmpl.ReplaceBytes(content, "common.RegexMatch", "helper.RegexMatch")
		content = tmpl.ReplaceBytes(content, tmpl.PHModuleName, moduleName)
		_, err = f.Write(content)
		if err != nil {
			return err
		}
	}
	return nil
}

func createGoProj(path string, goVer string, code string, noNet bool) string {
	// 收尾函数
	pathExist, pathNonExist, _ := splitExistPath(path)
	cwd := env.GetCwd()
	defer os.Chdir(cwd)

	// 如果创建失败，清理残余文件
	common.SetExitDefer(func() {
		defer os.Chdir(cwd)
		for _, f := range filesCreated {
			if f != nil {
				f.Close()
			}
		}
		err := os.Chdir(pathExist)
		if err != nil {
			fmt.Printf("cleanup failed(chdir): %v\n", err)
			return
		}
		err = os.RemoveAll(pathNonExist)
		if err != nil {
			fmt.Printf("cleanup failed(remove): %v\n", err)
		}
	})
	defer common.ClearExitDefer()

	// 尝试创建并进入项目目录
	err := os.MkdirAll(path, 0755)
	common.FailExit(err)
	err = os.Chdir(path)
	common.FailExit(err)
	moduleName := filepath.Base(path)

	// 创建go.mod文件
	fmt.Printf("go.mod: module - %s, go version - %s\n", moduleName, goVer)
	goMod := fmt.Sprintf("module %s\ngo %s\n", moduleName, goVer)
	f, err := os.Create("go.mod")
	common.FailExit(err)
	defer f.Close()
	filesCreated = append(filesCreated, f)
	_, err = f.WriteString(goMod)
	common.FailExit(err)

	// 创建components目录
	err = os.Mkdir("./components", 0755)
	common.FailExit(err)

	// 创建fuzzTypes包
	err = addFuzzType("./components/", noNet, moduleName)
	common.FailExit(err)

	// 创建helper包
	err = addHelpers("./components/", moduleName)
	common.FailExit(err)

	// 创建main.go文件
	fmt.Printf("creating main.go")
	mainGo, err := os.Create("main.go")
	common.FailExit(err)
	defer mainGo.Close()
	filesCreated = append(filesCreated, mainGo)
	fmt.Println(" success")

	// 将模板中的模块名替换后，写入main.go
	code = tmpl.Replace(code, tmpl.PHModuleName, moduleName)
	code = strings.TrimPrefix(code, "\n")
	code = strings.TrimSuffix(code, "\n")
	_, err = mainGo.WriteString(code)
	common.FailExit(err)
	return filepath.Join(pathExist, pathNonExist)
}

func runCmdGen(cmd *cobra.Command, _ []string) {
	common.SetCurrentCmd(cmd.Use)
	// 检查插件类型是否支持
	pType, _ := cmd.Flags().GetString("type")
	if convention.GetPluginFunName(pType) == "" {
		common.FailExit(fmt.Sprintf("unsupported plugin type %s", pType))
	}
	// 创建项目
	env1 := env.Check()
	goVer := env1.GoVersion
	path, _ := cmd.Flags().GetString("dir")
	// 如果为空则使用当前目录
	if path == "" {
		path = "."
	}
	code := convention.GenCodePType(pType)
	noNet, _ := cmd.Flags().GetBool("no-net")
	projPath := createGoProj(path, goVer, code, noNet)
	fmt.Printf("successfully create go project at %s\n", projPath)
}
