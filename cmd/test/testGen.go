package test

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nostalgist134/FuzzGIUPluginKit/cmd/common"
	"github.com/nostalgist134/FuzzGIUPluginKit/convention"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"strings"
)

var subCmdGen = &cobra.Command{
	Use:   "gen",
	Short: "generate test data set for plugin",
	Long: `generate test data for plugin
	Use text files as data sources to generate test data. files are specified using -f flag, each
	seperated by comma, WHICH MEANS YOU CAN'T USE FILES WHOSE NAME CONTAINS COMMA.

	Each file will be automatically parsed according to its corresponding param's type: for the 
	parameter of basic types, the file content will be read line by line as each test param value; 
	for the struct parameters, the file will be parsed as json expression of one struct or array of
	structs(use any form of each as your will, program will parse automatically).

	Some plugin has parameter of struct type. If you don't want to write code to get these structs,
	you can use -s flag to marshal a struct(whose value is settled, though) to a file and modify
	the file and use it as data source.

	You must specify as many files as plugin's parameters or one more file(the one exceeds will be 
	treated as test expect source).

	Each file is allowed to contain different numbers of test parameters. program will cycle around 
	the shorter files(like fuzzGIU's pitchfork-cycle mode)`,
	Run: runCmdGen,
}

func init() {
	subCmdGen.Flags().StringP("path", "p", "", "path of plugin to generate test data")
	subCmdGen.Flags().StringP("out", "o", "test.json", "out file")
	subCmdGen.Flags().StringP("files", "f", "", "source files, each seperated with comma")
	subCmdGen.Flags().StringP("struct", "s", "", "marshal structs by type to a "+
		"file, which can be used as data source file in the future")
	subCmdGen.Flags().IntP("num", "n", 1, "number of struct to marshal")
}

func tryMarshal(test *Test) error {
	_, err := json.Marshal(test)
	return err
}

// getUserYN 获取用户输入Yes/No选项
func getUserYN(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s (Y/N): ", prompt)
	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to get user decision:", err)
		return false
	}
	line = strings.TrimSpace(line)
	if strings.ToLower(line) == "y" {
		return true
	}
	return false
}

// tryWrite 尝试写入文件
func tryWrite(b []byte, fname string) {
	for {
		err := os.WriteFile(fname, b, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "write to %s failed, reason: %v\n", fname, err)
			if getUserYN("specify a new file?") {
				fname = common.ReadInputLine("new file name: ")
				continue
			} else {
				fmt.Println("now exiting...")
				os.Exit(1)
			}
		}
		return
	}
}

// anyBasicVal 根据param类型和字符串input解析所有基本类型值
func anyBasicVal(param convention.Param, input string) any {
	var ret any
	// 根据参数类型解析输入
	switch param.Type {
	case "int":
		base := 10
		if len(input) > 1 && input[0:2] == "0x" {
			base = 16
			input = input[2:]
		}
		intVar, err := strconv.ParseInt(input, base, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse %s to integer failed: %v. please retry\n", input, err)
			return err
		}
		ret = intVar
	case "string":
		ret = input
	case "float64":
		floatVar, err := strconv.ParseFloat(input, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse %s to float failed: %v. please retry\n", input, err)
			return err
		}
		ret = floatVar
	case "bool":
		boolVar, err := strconv.ParseBool(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse %s to bool failed: %v. please retry\n", input, err)
			return err
		}
		ret = boolVar
	default:
		// 非基本类型，返回error
		ret = fmt.Errorf("type %s is not basic type", param.Type)
	}
	return ret
}

// readFileLines 函数接收文件名，按行读取文件，并返回字符串切片
func readFileLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func any2Str(a any) string {
	if s, ok := a.(string); ok {
		return "\"" + s + "\""
	}
	return fmt.Sprintf("%v", a)
}

// genTestsFiles 从文件中读取测试数据，并生成测试例文件
func genTestsFiles(fd convention.FuncDecl, fileNames string) ([]*Test, error) {
	tests := make([]*Test, 0)
	fnames := strings.Split(fileNames, ",")
	// 若文件数量与参数数量不同，直接退出
	if len(fnames) != len(fd.Params) && len(fnames) != len(fd.Params)+1 {
		return nil, errors.New("number of file sources not matched")
	}
	maxSrcLen := 0
	dataSources := make([][]any, 0)
	// 对文件个数循环
	for i, fname := range fnames[:len(fd.Params)] {
		// 参数类型为结构体
		if stru := convention.GetStruct(fd.Params[i].Type); stru != nil {
			j, err := readFileJson(fname, stru)
			common.FailExit(err)
			dataSources = append(dataSources, j)
			if len(j) > maxSrcLen {
				maxSrcLen = len(j)
			}
			continue
		}
		// 参数为基本类型
		lines, err := readFileLines(fname)
		common.FailExit(err)
		if len(lines) == 0 {
			common.FailExit(fmt.Sprintf("file %s is empty", fname))
		}
		dataSource := make([]any, 0)
		for j, line := range lines {
			v := anyBasicVal(fd.Params[i], line)
			// 解析失败，退出
			if e, ok := v.(error); ok {
				common.FailExit(fmt.Errorf("parsing %s: '%s'(line#%d) to type %s failed: %v", fname, line,
					j, fd.Params[i].Type, e))
			}
			dataSource = append(dataSource, v)
		}
		if len(dataSource) > maxSrcLen {
			maxSrcLen = len(dataSource)
		}
		dataSources = append(dataSources, dataSource)
	}
	// 获取expect
	var expects []any = nil
	if len(fnames) > len(fd.Params) {
		retType := fd.RetType
		expectFile := fnames[len(fnames)-1]
		// 返回值是结构
		if stru := convention.GetStruct(retType); stru != nil {
			j, err := readFileJson(expectFile, stru)
			if err != nil {
				fmt.Fprintf(os.Stderr, "parse file %s failed, skipping get expects\n", expectFile)
			}
			expects = j
		} else if ret := convention.GetRetPtr(retType); ret != nil { // 返回值是基本类型
			lines, err := readFileLines(expectFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "read file %s failed, skipping get expects\n", expectFile)
			}
			expects = make([]any, 0)
			for _, l := range lines {
				val := anyBasicVal(convention.Param{Type: retType}, l)
				if e, ok := val.(error); ok {
					fmt.Fprintf(os.Stderr, "error when parse line '%s' to %s: %v. skipping\n", l, retType, e)
					continue
				}
				expects = append(expects, val)
			}
		} else {
			fmt.Fprintf(os.Stderr, "unknown return type %s, skipping get expects\n", fd.RetType)
		}
	}
	if expects != nil {
		dataSources = append(dataSources, expects)
	}
	// 根据最大的数据源生成测试结构切片
	for i := 0; i < maxSrcLen; i++ {
		t := new(Test)
		t.Args = make([]any, 0)
		j := 0
		// 生成测试参数
		fmt.Printf("test#%d: [args:[", i)
		for ; j < len(fd.Params); j++ {
			arg := dataSources[j][i%len(dataSources[j])]
			fmt.Print(any2Str(arg))
			t.Args = append(t.Args, arg)
			if j < len(fd.Params)-1 {
				fmt.Print(", ")
			}
		}
		fmt.Print("]")
		// 生成expect值
		if j < len(dataSources) && len(dataSources[j]) > 0 {
			fmt.Print(", expect:")
			expect := dataSources[j][i%len(dataSources[j])]
			fmt.Print(any2Str(expect))
			t.Expect = expect
		}
		fmt.Println("]")
		common.FailExit(tryMarshal(t))
		tests = append(tests, t)
	}
	return tests, nil
}

// genStructJsonFile 将一种类型的结构转化为json文件，若数量为1则转化单个json对象，否则转化为切片
func genStructJsonFile(struType string, to string, num int) {
	if num <= 0 {
		return
	}
	f, err := os.Create(to)
	common.FailExit(err)
	defer f.Close()
	struType = strings.TrimSpace(struType)
	if len(struType) > 0 && struType[0] != '*' {
		struType = "*" + struType
	}
	s := convention.GetFullStruct(struType)
	if s == nil {
		common.FailExit(fmt.Errorf("unsupported struct type '%s'", struType))
	}
	// 若数量为1则转化单个对象
	if num == 1 {
		b, err := json.MarshalIndent(s, "", "  ")
		common.FailExit(err)
		f.Write(b)
		return
	}
	structs := make([]any, 0)
	structs = append(structs, s)
	for i := 1; i < num; i++ {
		structs = append(structs, convention.GetFullStruct(struType))
	}
	b, err := json.MarshalIndent(structs, "", "  ")
	common.FailExit(err)
	f.Write(b)
}

func runCmdGen(cmd *cobra.Command, _ []string) {
	common.SetCurrentCmd(Cmd.Use + " " + cmd.Use)
	out, _ := cmd.Flags().GetString("out")
	stru, _ := cmd.Flags().GetString("struct")
	if stru != "" {
		numStru, _ := cmd.Flags().GetInt("num")
		genStructJsonFile(stru, out, numStru)
	} else {
		path, _ := cmd.Flags().GetString("path")
		if path == "" {
			common.FailExit("missing plugin file path")
		}
		// 获取插件元信息
		inf, err := common.GetPluginInfo(path)
		common.FailExit(err)
		fd := convention.BuildFd(inf)
		var tests []*Test
		// 根据插件参数生成测试数据
		files, _ := cmd.Flags().GetString("files")
		tests, err = genTestsFiles(fd, files)
		common.FailExit(err)
		// 写入文件
		b, _ := json.MarshalIndent(tests, "", "  ")
		tryWrite(b, out)
	}
}
