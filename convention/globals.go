package convention

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output/outputFlag"
	"time"
)

const (
	IndPTypePlProc = iota
	IndPTypeReact
	IndPTypePlGen
	IndPTypeRequester
	IndPTypePreproc
	IndPTypeIterator
	IndPTypeIteratorMinor = 0
)

var PluginFunNames = []string{"PayloadProcessor", "React", "PayloadGenerator", "DoRequest", "Preprocess", "IterIndex"}
var PluginTypes = []string{"payloadProc", "reactor", "payloadGen", "requester", "preprocess", "iterator"}
var PluginMinorFun = []string{"IterLen"}

// FuncDecls 每种插件的约定函数原型
var FuncDecls = map[string]FuncDecl{
	PluginTypes[IndPTypePlProc]: {
		Params:  []Param{{Name: "payload", Type: "string"}},
		RetType: "string",
	},
	PluginTypes[IndPTypeReact]: {
		Params:  []Param{{Name: "req", Type: "*fuzzTypes.Req"}, {Name: "resp", Type: "*fuzzTypes.Resp"}},
		RetType: "*fuzzTypes.Reaction",
	},
	PluginTypes[IndPTypePlGen]: {
		Params:  []Param{},
		RetType: "[]string",
	},
	PluginTypes[IndPTypeRequester]: {
		Params:  []Param{{Name: "requestCtx", Type: "*fuzzTypes.RequestCtx"}},
		RetType: "*fuzzTypes.Resp",
	},
	PluginTypes[IndPTypePreproc]: {
		Params:  []Param{{Name: "fuzz", Type: "*fuzzTypes.Fuzz"}},
		RetType: "*fuzzTypes.Fuzz",
	},
	PluginTypes[IndPTypeIterator]: {
		Params:  []Param{{Name: "lengths", Type: "[]int"}, {"ind", "int"}},
		RetType: "[]int",
	},
	PluginMinorFun[IndPTypeIteratorMinor]: {
		Params:  []Param{{Name: "lengths", Type: "[]int"}},
		RetType: "int",
	},
}

var fullReq = &fuzzTypes.Req{
	URL: "https://test.com",
	HttpSpec: fuzzTypes.HTTPSpec{
		Method:     "POST",
		Headers:    []string{"Hello: FUZZGIU", "User-Agent: milaogiu browser(114.54)", "Giu: 12345"},
		Proto:      "HTTP/2",
		ForceHttps: false,
	},
	Fields: []fuzzTypes.Field{
		{"NISHIGIU", "WOSHIGIU"},
		{"MILAOGIU", "NISHIGIU"},
		{"FUZZ", "NISHIGIU"},
	},
	Data: []byte("user=GIU&password=GIU12345"),
}

var fullResp = &fuzzTypes.Resp{
	StatCode:          200,
	ResponseTime:      5 * time.Millisecond,
	Size:              999,
	Words:             569,
	Lines:             12,
	HttpRedirectChain: "nishigiu->woshigiu->milaogiu",
	RawResponse:       []byte("FUZZGIU FUZZGIU"),
	ErrMsg:            "test error",
}

var fullRequestCtx = &fuzzTypes.RequestCtx{
	Request:             fullReq,
	Proxy:               "http://127.0.0.1:8080",
	HttpFollowRedirects: true,
	Retry:               3,
	RetryCodes:          []fuzzTypes.Range{{401, 403}, {502, 502}, {503, 503}},
	RetryRegex:          "nishigiu",
	Timeout:             10,
}

var Ranges = []fuzzTypes.Range{{500, 300}, {200, 100}, {404, 200}}
var fullFilter = fuzzTypes.Match{
	Code:  Ranges,
	Lines: Ranges,
	Words: Ranges,
	Size:  Ranges,
	Regex: "123",
	Mode:  "or",
}

var fullFuzz = &fuzzTypes.Fuzz{
	Preprocess: fuzzTypes.FuzzStagePreprocess{
		PlMeta: map[string]*fuzzTypes.PayloadMeta{
			"FUZZ1": {
				Processors: []fuzzTypes.Plugin{
					{"addslashes", nil},
					{"base64", nil},
				},
				Generators: fuzzTypes.PlGen{
					Wordlists: []string{"1.txt", "2.txt"},
					Plugins: []fuzzTypes.Plugin{
						{"dict.txt", nil},
					},
				},
			},
			"FUZZ2": {
				Processors: []fuzzTypes.Plugin{
					{"addslashes", nil},
					{"base64", nil},
				},
				Generators: fuzzTypes.PlGen{
					Wordlists: []string{"1.txt", "2.txt"},
					Plugins: []fuzzTypes.Plugin{
						{"dict.txt", nil},
					},
				},
			},
		},
		ReqTemplate: *fullReq,
	},
	Request: fuzzTypes.FuzzStageRequest{
		Proxies:             []string{"http://127.0.0.1:8080", "http://127.0.0.1:7890"},
		HttpFollowRedirects: true,
		Retry:               2,
		RetryCodes:          []fuzzTypes.Range{{404, 406}, {407, 409}},
		RetryRegex:          "giu",
		Timeout:             3,
	},
	React: fuzzTypes.FuzzStageReact{
		Reactor: fuzzTypes.Plugin{Name: "test", Args: []any{1, 2, 3}},
		Filter:  fullFilter,
		Matcher: fullFilter,
		RecursionControl: fuzzTypes.ReactRecursionControl{
			RecursionDepth:    3,
			MaxRecursionDepth: 5,
			Keyword:           "FUZZ1",
			StatCodes:         Ranges,
			Regex:             "what",
			Splitter:          "/",
		},
	},
	Control: fuzzTypes.FuzzControl{
		PoolSize: 64,
		Delay:    50,
		OutSetting: fuzzTypes.OutputSetting{
			Verbosity:    3,
			OutputFile:   "nishigiu.json",
			OutputFormat: "json",
			HttpURL:      "http://www.nishigiu.com/submit",
			ChanSize:     10,
			ToWhere:      outputFlag.OutToChan | outputFlag.OutToStdout,
		},
	},
}

var fullReaction = &fuzzTypes.Reaction{
	Flag: fuzzTypes.ReactOutput | fuzzTypes.ReactAddJob | fuzzTypes.ReactAddReq,
	Output: struct {
		Msg       string `json:"msg,omitempty"`
		Overwrite bool   `json:"overwrite,omitempty"`
	}{
		Msg:       "NISHIGIU",
		Overwrite: false,
	},
	NewJob: fullFuzz,
	NewReq: fullReq,
}
