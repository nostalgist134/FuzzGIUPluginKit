package cmd

import (
	"FuzzGIUPluginKit/cmd/build"
	"FuzzGIUPluginKit/cmd/gen"
	"FuzzGIUPluginKit/cmd/info"
	"FuzzGIUPluginKit/cmd/test"
	"fmt"
	"github.com/spf13/cobra"
)

var entry = &cobra.Command{
	Use: "help",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("FuzzGIUPluginKit is a tool for create and common plugins of FuzzGIU" +
			"(https://github.com/nostalgist134/FuzzGIU)")
	},
}

func init() {
	entry.AddCommand(build.Cmd)
	entry.AddCommand(gen.Cmd)
	entry.AddCommand(info.Cmd)
	entry.AddCommand(test.Cmd)
}

func RunCmd() {
	if err := entry.Execute(); err != nil {
		fmt.Printf("Run command error - %v\n", err)
	}
}
