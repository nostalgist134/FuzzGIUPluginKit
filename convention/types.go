package convention

// Param 参数
type Param struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type ParaMeta struct {
	Param    Param  `json:"param"`
	ParaInfo string `json:"para_info,omitempty"`
}

// FuncDecl 函数声明
type FuncDecl struct {
	Params  []Param
	RetType string
}

type PluginInfo struct {
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	GoVersion string     `json:"go_version"`
	UsageInfo string     `json:"usage_info,omitempty"`
	Params    []ParaMeta `json:"params"`
}
