package cmd

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIUPluginKit/cmd/build"
	"github.com/nostalgist134/FuzzGIUPluginKit/cmd/gen"
	"github.com/nostalgist134/FuzzGIUPluginKit/cmd/info"
	"github.com/nostalgist134/FuzzGIUPluginKit/cmd/test"
	"github.com/nostalgist134/FuzzGIUPluginKit/version"
	"github.com/spf13/cobra"
)

var entry = &cobra.Command{}

func init() {
	entry.AddCommand(build.Cmd)
	entry.AddCommand(gen.Cmd)
	entry.AddCommand(info.Cmd)
	entry.AddCommand(test.Cmd)
	oldHelp := entry.HelpFunc()
	entry.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Printf("FuzzGIUPluginKit %s - a tool for develop/test plugins for"+
			" FuzzGIU(https://github.com/nostalgist134/FuzzGIU)\n", version.GetVersion())
		oldHelp(cmd, args)
	})
}

func RunCmd() {
	if err := entry.Execute(); err != nil {
		fmt.Printf("Run command error - %v\n", err)
	}
}
