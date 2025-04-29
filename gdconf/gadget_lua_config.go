package gdconf

import (
	"fmt"
	"os"
	"strings"

	"github.com/flswld/halo/logger"
	lua "github.com/yuin/gopher-lua"
)

type GadgetLuaConfig struct {
	LuaStr   string
	LuaState *lua.LState
}

func (g *GameDataConfig) loadGadgetLuaConfig() {
	g.GadgetLuaConfigMap = make(map[string]*GadgetLuaConfig)
	g.loadGadgetLuaConfigLoop(g.luaPrefix + "gadget")
	logger.Info("GadgetLuaConfig Count: %v", len(g.GadgetLuaConfigMap))
}

func (g *GameDataConfig) loadGadgetLuaConfigLoop(path string) {
	fileList, err := os.ReadDir(path)
	if err != nil {
		info := fmt.Sprintf("open file error: %v, path: %v", err, path)
		panic(info)
	}
	for _, file := range fileList {
		fileName := file.Name()
		filePath := path + "/" + fileName
		if file.IsDir() {
			g.loadGadgetLuaConfigLoop(filePath)
		}
		split := strings.Split(fileName, ".")
		if split[len(split)-1] != "lua" {
			continue
		}
		if len(split) != 2 {
			continue
		}
		gadgetLuaName := split[0]
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			info := fmt.Sprintf("open file error: %v", err)
			panic(info)
		}
		if fileData[0] == 0xEF && fileData[1] == 0xBB && fileData[2] == 0xBF {
			fileData = fileData[3:]
		}
		gadgetLuaConfig := new(GadgetLuaConfig)
		gadgetLuaConfig.LuaStr = string(fileData)
		g.GadgetLuaConfigMap[gadgetLuaName] = gadgetLuaConfig
	}
}

func GetGadgetLuaConfigByName(name string) *GadgetLuaConfig {
	gadgetLuaConfig := CONF.GadgetLuaConfigMap[name]
	if gadgetLuaConfig == nil {
		return nil
	}
	if gadgetLuaConfig.LuaState == nil {
		luaState, err := newLuaState(gadgetLuaConfig.LuaStr)
		if err != nil {
			logger.Error("lua parse error: %v, name: %v", err, name)
		}
		scriptLib := luaState.NewTable()
		luaState.SetGlobal("ScriptLib", scriptLib)
		for _, scriptLibFunc := range SCRIPT_LIB_FUNC_LIST {
			luaState.SetField(scriptLib, scriptLibFunc.fnName, luaState.NewFunction(scriptLibFunc.fn))
		}
		gadgetLuaConfig.LuaState = luaState
	}
	return gadgetLuaConfig
}
