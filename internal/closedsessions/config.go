package closedsessions

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the value loaded from the config file.
type Config struct {
	BaseSearchPath string
}

// LoadConfig reads the config file from ~/.config/closed-sessions/config.yaml and returns the runtime config.
func LoadConfig() (Config, error) {
	v := viper.New()
	cfgDir := ConfigDir()
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		return Config{}, err
	}

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(cfgDir)
	v.SetDefault("base_search_path", DefaultBaseSearchPath())

	if _, err := os.Stat(ConfigFile()); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(ConfigFile(), []byte(fmt.Sprintf("base_search_path: %s\n", DefaultBaseSearchPath())), 0o644); err != nil {
			return Config{}, err
		}
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return Config{}, err
		}
	}

	config := Config{BaseSearchPath: v.GetString("base_search_path")}
	if config.BaseSearchPath == "" {
		config.BaseSearchPath = DefaultBaseSearchPath()
		v.Set("base_search_path", config.BaseSearchPath)
		if err := v.WriteConfigAs(ConfigFile()); err != nil {
			return Config{}, err
		}
	}

	return config, nil
}

// SaveConfig persists the configured base search path in the YAML config file.
func SaveConfig(baseSearchPath string) error {
	cfgDir := ConfigDir()
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		return err
	}

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(cfgDir)
	v.Set("base_search_path", baseSearchPath)
	return v.WriteConfigAs(filepath.Join(cfgDir, "config.yaml"))
}
