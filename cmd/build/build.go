package build

import (
	"FuzzGIUPluginKit/cmd/common"
	"FuzzGIUPluginKit/convention"
	"FuzzGIUPluginKit/env"
	"FuzzGIUPluginKit/goParser"
	"FuzzGIUPluginKit/tmpl"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var Cmd = &cobra.Command{
	Use:   "build",
	Short: "build plugin",
	Run:   runCmdBuild,
}

func init() {
	Cmd.Flags().StringP("path", "p", "", "path/file to build plugin")
	Cmd.Flags().StringP("out", "o", "", "out file")
	Cmd.Flags().StringP("go-path", "g", "", "go binary path")
	Cmd.Flags().StringP("usage-file", "u", "", "usage file to be used for PluginInfo")
	Cmd.Flags().BoolP("no-clean", "k", false, "keep intermediate files")
	Cmd.Flags().BoolP("info", "i", false, "generate PluginInfo function for plugin")
}

func exclusiveImports(imp []string, imp1 []string) []string {
	exclusive := make([]string, 0)
	for _, i := range imp {
		overlap := false
		for _, j := range imp1 {
			if i == j {
				overlap = true
				break
			}
		}
		if !overlap {
			exclusive = append(exclusive, i)
		}
	}
	return exclusive
}

func getImpStr(imports []string) string {
	if len(imports) == 0 {
		return ""
	}
	sb := strings.Builder{}
	sb.WriteString("import (\n")
	for _, i := range imports {
		sb.WriteByte('\t')
		sb.WriteString(i)
		sb.WriteByte('\n')
	}
	sb.WriteByte(')')
	return sb.String()
}

func buildSharedLib(goPath string, src string, out string, env1 env.Env, funInfo *convention.FuncDecl) {
	// go mod tidy
	fmt.Printf("> %s mod tidy\n", goPath)
	c := exec.Command(goPath, "mod", "tidy")
	output, err := c.CombinedOutput()
	if len(output) > 0 {
		fmt.Println(string(output))
	}
	if err != nil {
		buf := make([]byte, 1)
		fmt.Print("go mod tidy failed, continue building anyway?(Y/N)")
		os.Stdin.Read(buf)
		if buf[0] != 'Y' && buf[0] != 'y' {
			common.FailExit(err)
		}
	}
	// go fmt
	fmt.Printf("> %s fmt wrapped.go\n", goPath)
	c = exec.Command(goPath, "fmt", "wrapped.go")
	output, _ = c.CombinedOutput()
	if len(output) > 0 {
		fmt.Print(string(output))
	}
	// go build ...
	buildArgs := env.GetBuildArgs(env1, out, src)
	fmt.Printf("> %s ", goPath)
	for _, a := range buildArgs {
		if strings.Index(a, "-ldflags=") == 0 {
			fmt.Printf("%s\"%s\" ", a[0:9], a[9:])
			continue
		}
		fmt.Printf("%s ", a)
	}
	os.Stdout.Write([]byte{'\n'})
	c = exec.Command(goPath, buildArgs...)
	output, err = c.CombinedOutput()
	if len(output) > 0 {
		fmt.Print(string(output))
	}
	common.FailExit(err)
	out, _ = filepath.Abs(out)
	fmt.Printf("successfully built %s, parameters: %v\n", out, funInfo.Params)
}

func runCmdBuild(cmd *cobra.Command, _ []string) {
	common.SetCurrentCmd(cmd.Use)
	cwd := env.GetCwd()
	common.SetExitDefer(func() {
		os.Chdir(cwd)
	})
	defer os.Chdir(cwd)
	goPath, _ := cmd.Flags().GetString("go-path")
	if goPath == "" {
		goPath = "go"
	}
	// 检查构建环境
	env1 := env.Check(goPath)
	if env1.OkToBuild == false {
		common.FailExit("environment check failed")
	}
	fmt.Printf("currently build under %s, using go version %s\n", env1.OS, env1.GoVersion)
	// 检查路径
	path, err := cmd.Flags().GetString("path")
	common.FailExit(err)
	if path == "" {
		common.FailExit("missing build path/file")
	}
	pluginFile := path
	stat, err := os.Stat(path)
	common.FailExit(err)
	if stat.IsDir() {
		pluginFile = filepath.Join(pluginFile, "main.go")
	}
	// 寻找插件函数
	var fd *convention.FuncDecl
	var paraMeta []convention.ParaMeta
	pType := ""
	for _, pFun := range convention.PluginFunNames {
		fd, paraMeta, err = goParser.FindFunction(pluginFile, pFun)
		if os.IsNotExist(err) {
			continue
		}
		common.FailExit(err)
		pType = convention.GetPluginType(pFun)
		break
	}
	if os.IsNotExist(err) || fd == nil {
		common.FailExit("cannot find supported plugin function")
	}
	// 检查插件函数是否符合约定
	fmt.Printf("plugin type - %s\n", pType)
	ok, msg := convention.CheckPluginFun(pType, *fd)
	if !ok {
		common.FailExit(fmt.Sprintf("plugin function check failed: %s", msg))
	}

	wrapped, err := tmpl.GetTemplate(env1.OS, pType)
	common.FailExit(err)
	// 替换模板中去重的import语句
	tempImports, _ := goParser.GetImports(wrapped, true)
	srcImports, _ := goParser.GetImports(pluginFile)
	eImports := exclusiveImports(srcImports, tempImports)
	wrapped = tmpl.Replace(wrapped, tmpl.PHCustomImports, getImpStr(eImports), -1)
	// 替换形参与实参列表
	formal, actual := convention.GetParamStrings(env1.OS, pType, fd.Params)
	wrapped = tmpl.Replace(wrapped, tmpl.PHFormalPara, formal, -1)
	wrapped = tmpl.Replace(wrapped, tmpl.PHActualPara, actual, -1)
	// 将code占位符替换为源码
	code, err := goParser.GetCode(pluginFile)
	common.FailExit(err)
	wrapped = tmpl.Replace(wrapped, tmpl.PHCode, code, -1)
	// 输出文件名
	out, _ := cmd.Flags().GetString("out")
	if out == "" {
		out = "FuzzGIU" + convention.GetPluginFunName(pType) + env1.BinSuffix
	}
	fmt.Printf("output file: %s\n", out)
	// 根据需要生成PluginInfo函数
	if genPi, _ := cmd.Flags().GetBool("info"); genPi {
		usageFile, _ := cmd.Flags().GetString("usage-file")
		wrapped += "\n" + convention.GenPlugInfoFun(out, pType, env1.GoVersion, usageFile, paraMeta)
	}
	// 进入项目目录，创建并写入文件
	err = os.Chdir(filepath.Dir(pluginFile))
	common.FailExit(err)
	f, err := os.Create("wrapped.go")
	common.FailExit(err)
	defer f.Close()
	_, err = f.WriteString(wrapped)
	common.FailExit(err)
	// 编译文件
	buildSharedLib(goPath, f.Name(), out, env1, fd)
	// 决定是否保留中间文件
	if noClean, _ := cmd.Flags().GetBool("no-clean"); !noClean {
		f.Close()
		defer env.RemoveIntermediateFiles(env1, out)
	}
}
