package convention

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"time"
)

const (
	IndPTypePlProc    = 0
	IndPTypeReact     = 1
	IndPTypePlGen     = 2
	IndPTypeReqSender = 3
	IndPTypePreproc   = 4
)

var PluginFunNames = []string{"PayloadProcessor", "React", "PayloadGenerator", "SendRequest", "Preprocess"}
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

var fullReq = &fuzzTypes.Req{
	URL: "https://test.com",
	HttpSpec: fuzzTypes.HTTPSpec{
		Method:     "POST",
		Headers:    []string{"Hello: 1", "User-Agent: milaogiu browser(114.54)", "Giu: 12345"},
		Version:    "2.0",
		ForceHttps: false,
	},
	Data: "user=GIU&password=GIU12345",
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

var fullSendMeta = &fuzzTypes.SendMeta{
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
		Mode          string                           `json:"mode,omitempty"`          // 出现多个payload关键字时处理的模式
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
	Send: struct {
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
		Reactor          fuzzTypes.Plugin         `json:"reactors,omitempty"`
		OutSettings      fuzzTypes.OutputSettings `json:"output_settings,omitempty"`
		Filter           fuzzTypes.Match          `json:"filter,omitempty"`
		Matcher          fuzzTypes.Match          `json:"matcher,omitempty"`
		RecursionControl struct {
			RecursionDepth    int               `json:"recursion_depth,omitempty"`
			MaxRecursionDepth int               `json:"max_recursion_depth,omitempty"`
			Keyword           string            `json:"keyword,omitempty"`
			StatCodes         []fuzzTypes.Range `json:"stat_codes,omitempty"`
			Regex             string            `json:"regex,omitempty"`
			Splitter          string            `json:"splitter,omitempty"`
		} `json:"recursion_control,omitempty"`
	}{
		Reactor: fuzzTypes.Plugin{Name: "test", Args: []any{1, 2, 3}},
		OutSettings: fuzzTypes.OutputSettings{
			Verbosity:    3,
			IgnoreError:  false,
			OutputFormat: "json",
			OutputFile:   "test.json",
			NativeStdout: false,
		},
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
	Misc: struct {
		PoolSize         int           `json:"pool_size,omitempty"`
		Delay            int           `json:"delay,omitempty"`
		DelayGranularity time.Duration `json:"delay_granularity,omitempty"`
	}{
		PoolSize:         64,
		Delay:            50,
		DelayGranularity: time.Millisecond,
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
