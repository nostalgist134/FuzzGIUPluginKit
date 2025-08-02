package info

import (
	"FuzzGIUPluginKit/cmd/common"
	"FuzzGIUPluginKit/convention"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var Cmd = &cobra.Command{
	Use:   "info",
	Short: "get plugin information",
	Run:   runCmdInfo,
}

func init() {
	Cmd.Flags().StringP("path", "p", "", "plugin binary path")
	Cmd.Flags().StringP("format", "f", "", "output format(native, json)")
}

func outputPluginInfo(info *convention.PluginInfo, format string) {
	if format == "json" {
		jInfo, _ := json.MarshalIndent(info, "", "  ")
		fmt.Print(string(jInfo))
		return
	}
	formattedOut := func(title string, content ...any) {
		fmt.Printf("%-15s: %s\n", title, content)
	}
	formattedOut("plugin", info.Name)
	formattedOut("plugin type", info.Type)
	formattedOut("go version", info.GoVersion)
	formattedOut("usage", info.UsageInfo)
	fmt.Printf("parameters >")
	if len(info.Params) > 0 {
		os.Stdout.Write([]byte{'\n'})
	}
	for _, pm := range info.Params {
		fmt.Printf("        %-7s %-7s", pm.Param.Name, pm.Param.Type)
		if pm.ParaInfo != "" {
			fmt.Printf(" - \"%s\"", pm.ParaInfo)
		}
		os.Stdout.Write([]byte{'\n'})
	}
}

func runCmdInfo(cmd *cobra.Command, _ []string) {
	common.SetCurrentCmd(cmd.Use)
	path, _ := cmd.Flags().GetString("path")
	if path == "" {
		common.FailExit("missing plugin path")
	}
	pi, err := common.GetPluginInfo(path)
	common.FailExit(err)
	format, _ := cmd.Flags().GetString("format")
	outputPluginInfo(pi, format)
}
