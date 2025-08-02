package test

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use: "test",
	Short: "test a plugin's functionality. this command expect plugin tested be\n" +
		"built with PluginInfo(-i of build command)",
}

type Test struct {
	Args   []any `json:"args,omitempty"`   // 测试参数列表
	Expect any   `json:"expect,omitempty"` // 期望返回值
}

type ResultTest struct {
	T      Test `json:"test"`
	Result any  `json:"result"`
	Passed bool `json:"passed,omitempty"`
}

func init() {
	Cmd.AddCommand(subCmdRun)
	Cmd.AddCommand(subCmdGen)
}
