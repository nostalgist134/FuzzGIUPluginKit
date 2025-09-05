//go:build linux || darwin

package common

import (
	"encoding/json"
	"errors"
	"github.com/nostalgist134/FuzzGIUPluginKit/convention"
	goPlugin "plugin"
)

func GetPluginInfo(pluginFile string) (*convention.PluginInfo, error) {
	p, err := goPlugin.Open(pluginFile)
	if err != nil {
		return nil, err
	}

	piSym, err := p.Lookup("PluginInfo")
	if err != nil {
		return nil, err
	}

	pluginInfoFun, ok := piSym.(func() string)
	if !ok {
		return nil, errors.New("PluginInfo is not func() string")
	}

	pi := pluginInfoFun()
	if pi == "" {
		return nil, errors.New("PluginInfo returned empty")
	}

	ret := new(convention.PluginInfo)
	err = json.Unmarshal([]byte(pi), ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}
