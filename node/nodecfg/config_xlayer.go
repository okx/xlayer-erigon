package nodecfg

import (
	"fmt"

	"github.com/mitchellh/copystructure"
)

// TryClone is the helper method to return a deep copy of the ethconfig instance
func (c *Config) TryClone() (Config, error) {
	clone, err := copystructure.Copy(*c)
	if err != nil {
		return Config{}, err
	}
	ret, ok := clone.(Config)
	if !ok {
		return Config{}, fmt.Errorf("type assertion failed")
	}
	return ret, nil
}
