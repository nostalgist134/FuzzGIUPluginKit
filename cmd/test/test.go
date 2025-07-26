package test

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "test",
	Short: "test a plugin",
	Run:   runCmdTest,
}

func init() {
	Cmd.Flags().StringP("path", "p", "", "path of plugin binary file")
	Cmd.Flags().StringP("expr", "e", "", "parameter expression to be passed to plugin")
}

func runCmdTest(cmd *cobra.Command, args []string) {
	// 还没想好怎么写
}
