package apollo

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"

	"github.com/ledgerwatch/log/v3"
)

func setFlagsFromBytes(ctx *cli.Context, value interface{}) error {
	fileConfig := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(value.(string)), fileConfig)
	if err != nil {
		log.Error(fmt.Sprintf("failed to load config: %v error: %v", value, err))
		return err
	}

	// sets global flags to value in yaml/toml file
	for key, value := range fileConfig {
		if !ctx.IsSet(key) {
			if reflect.ValueOf(value).Kind() == reflect.Slice {
				sliceInterface := value.([]interface{})
				s := make([]string, len(sliceInterface))
				for i, v := range sliceInterface {
					s[i] = fmt.Sprintf("%v", v)
				}
				err := ctx.Set(key, strings.Join(s, ","))
				if err != nil {
					return fmt.Errorf("failed setting %s flag with values=%s error=%s", key, s, err)
				}
			} else {
				err := ctx.Set(key, fmt.Sprintf("%v", value))
				if err != nil {
					return fmt.Errorf("failed setting %s flag with value=%v error=%s", key, value, err)
				}
			}
		}
	}

	return nil
}
