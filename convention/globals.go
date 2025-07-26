package convention

var PluginFunNames = []string{"PayloadProcessor", "React", "PayloadGenerator", "ReqSender", "Preprocessor"}
var PluginTypes = []string{"payloadProc", "reactor", "payloadGen", "reqSender", "preprocess"}

// FuncDecls 每种插件的约定函数原型
var FuncDecls = map[string]FuncDecl{
	PluginTypes[0]: {
		Params:  []Param{{Name: "payload", Type: "string"}},
		RetType: "string",
	},
	PluginTypes[1]: {
		Params:  []Param{{Name: "req", Type: "*fuzzTypes.Req"}, {Name: "resp", Type: "*fuzzTypes.Resp"}},
		RetType: "*fuzzTypes.Reaction",
	},
	PluginTypes[2]: {
		Params:  []Param{},
		RetType: "[]string",
	},
	PluginTypes[3]: {
		Params:  []Param{{Name: "sendMeta", Type: "*fuzzTypes.SendMeta"}},
		RetType: "*fuzzTypes.Resp",
	},
	PluginTypes[4]: {
		Params:  []Param{{Name: "fuzz", Type: "*fuzzTypes.Fuzz"}},
		RetType: "*fuzzTypes.Fuzz",
	},
}
