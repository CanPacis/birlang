package config

import (
	"encoding/json"
	"os"
	"path"

	"github.com/canpacis/birlang/src/engine"
	"github.com/canpacis/birlang/src/thrower"
	"github.com/canpacis/birlang/src/util"
)

type Config struct {
	ColoredOutput        bool `json:"colored_output"`
	VerbosityLevel       int  `json:"verbosity_level"`
	MaximumCallstackSize int  `json:"maximum_callstack_size"`
}

func HandleConfig(instance *engine.BirEngine) {
	config_path := path.Join(instance.Directory, "bir.config.json")

	if _, err := os.Stat(config_path); !os.IsNotExist(err) {
		config := Config{}
		raw, _ := os.ReadFile(config_path)
		err := json.Unmarshal(raw, &config)

		if err != nil {
			instance.Thrower.WarnAnonymous("Could not properly parse the config file")
		} else {
			instance.ColoredOutput = config.ColoredOutput
			instance.VerbosityLevel = config.VerbosityLevel

			if config.MaximumCallstackSize != 0 {
				instance.MaximumCallstackSize = config.MaximumCallstackSize
			}

			instance.Thrower = thrower.Thrower{Owner: instance, Color: util.NewColor(instance.ColoredOutput)}
		}
	}
}
