package convention

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output/outputFlag"
	"time"
)

const (
	IndPTypePlProc        = 0
	IndPTypeReact         = 1
	IndPTypePlGen         = 2
	IndPTypeReqSender     = 3
	IndPTypePreproc       = 4
	IndPTypeIterator      = 5
	IndPTypeIteratorMinor = 0
)

var PluginFunNames = []string{"PayloadProcessor", "React", "PayloadGenerator", "DoRequest", "Preprocess", "IterIndex"}
var PluginTypes = []string{"payloadProc", "reactor", "payloadGen", "reqSender", "preprocess", "iterator"}
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
	PluginTypes[IndPTypeReqSender]: {
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
		Version:    "2.0",
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
	HttpResponse:      nil,
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
	RetryCode:           "401-403,502,503",
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
	Preprocess: struct {
		PlTemp        map[string]fuzzTypes.PayloadTemp `json:"pl_temp,omitempty"`
		Preprocessors []fuzzTypes.Plugin               `json:"preprocessors,omitempty"` // 使用的自定义预处理器
		ReqTemplate   fuzzTypes.Req                    `json:"request_tmpl,omitempty"`  // 含有fuzz关键字的请求模板
	}{
		PlTemp: map[string]fuzzTypes.PayloadTemp{
			"FUZZ1": {
				Processors: []fuzzTypes.Plugin{
					{"addslashes", nil},
					{"base64", nil},
				},
				Generators: fuzzTypes.PlGen{
					Type: "wordlist",
					Gen: []fuzzTypes.Plugin{
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
					Type: "wordlist",
					Gen: []fuzzTypes.Plugin{
						{"dict.txt", nil},
					},
				},
			},
		},
		ReqTemplate: *fullReq,
	},
	Request: struct {
		Proxies             []string `json:"proxies,omitempty"`
		HttpFollowRedirects bool     `json:"http_follow_redirects,omitempty"`
		Retry               int      `json:"retry,omitempty"`
		RetryCode           string   `json:"retry_code,omitempty"`
		RetryRegex          string   `json:"retry_regex,omitempty"`
		Timeout             int      `json:"timeout,omitempty"`
	}{
		Proxies:             []string{"http://127.0.0.1:8080", "http://127.0.0.1:7890"},
		HttpFollowRedirects: true,
		Retry:               2,
		RetryCode:           "405",
		RetryRegex:          "giu",
		Timeout:             3,
	},
	React: struct {
		Reactor          fuzzTypes.Plugin `json:"reactor,omitempty"`      // 响应处理插件
		Filter           fuzzTypes.Match  `json:"filter,omitempty"`       // 过滤
		Matcher          fuzzTypes.Match  `json:"matcher,omitempty"`      // 匹配
		IgnoreError      bool             `json:"ignore_error,omitempty"` // 是否忽略发送过程中出现的错误
		RecursionControl struct {
			RecursionDepth    int               `json:"recursion_depth,omitempty"`     // 当前递归深度
			MaxRecursionDepth int               `json:"max_recursion_depth,omitempty"` // 最大递归深度
			Keyword           string            `json:"keyword,omitempty"`
			StatCodes         []fuzzTypes.Range `json:"stat_codes,omitempty"`
			Regex             string            `json:"regex,omitempty"`
			Splitter          string            `json:"splitter,omitempty"`
		} `json:"recursion_control,omitempty"`
	}{
		Reactor: fuzzTypes.Plugin{Name: "test", Args: []any{1, 2, 3}},
		Filter:  fullFilter,
		Matcher: fullFilter,
		RecursionControl: struct {
			RecursionDepth    int               `json:"recursion_depth,omitempty"`
			MaxRecursionDepth int               `json:"max_recursion_depth,omitempty"`
			Keyword           string            `json:"keyword,omitempty"`
			StatCodes         []fuzzTypes.Range `json:"stat_codes,omitempty"`
			Regex             string            `json:"regex,omitempty"`
			Splitter          string            `json:"splitter,omitempty"`
		}{
			RecursionDepth:    3,
			MaxRecursionDepth: 5,
			Keyword:           "FUZZ1",
			StatCodes:         Ranges,
			Regex:             "what",
			Splitter:          "/",
		},
	},
	Control: struct {
		PoolSize   int                     `json:"pool_size,omitempty"`   // 使用的协程池大小
		Delay      time.Duration           `json:"delay,omitempty"`       // 每次提交任务前的延迟
		OutSetting fuzzTypes.OutputSetting `json:"out_setting,omitempty"` // 输出设置
		IterCtrl   fuzzTypes.Iteration     `json:"iter_ctrl,omitempty"`   // 迭代控制
	}{
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
