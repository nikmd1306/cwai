package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/ini.v1"
)

const (
	DefaultAPIURL         = "https://api.openai.com/v1"
	DefaultModel          = "gpt-5-mini"
	DefaultLanguage       = "en"
	DefaultMaxTokensInput = 4096
	DefaultMaxTokensOut   = 500
)

type Config struct {
	APIKey           string
	APIURL           string
	Model            string
	Language         string
	MaxTokensInput   int
	MaxTokensOutput  int
	Temperature      float64
	HasTemperature   bool
	ReasoningEffort  string
	Verbosity        string
	StructuredOutput string
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".cwai"), nil
}

func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		APIURL:          DefaultAPIURL,
		Model:           DefaultModel,
		Language:        DefaultLanguage,
		MaxTokensInput:  DefaultMaxTokensInput,
		MaxTokensOutput: DefaultMaxTokensOut,
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	f, err := ini.Load(path)
	if err != nil {
		return nil, fmt.Errorf("cannot load config %s: %w", path, err)
	}

	sec := f.Section("")

	if v := sec.Key("CWAI_API_KEY").String(); v != "" {
		cfg.APIKey = v
	}
	if v := sec.Key("CWAI_API_URL").String(); v != "" {
		cfg.APIURL = v
	}
	if v := sec.Key("CWAI_MODEL").String(); v != "" {
		cfg.Model = v
	}
	if v := sec.Key("CWAI_LANGUAGE").String(); v != "" {
		cfg.Language = v
	}
	if v := sec.Key("CWAI_MAX_TOKENS_INPUT").String(); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.MaxTokensInput = n
		}
	}
	if v := sec.Key("CWAI_MAX_TOKENS_OUTPUT").String(); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.MaxTokensOutput = n
		}
	}
	if v := sec.Key("CWAI_TEMPERATURE").String(); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.Temperature = f
			cfg.HasTemperature = true
		}
	}
	if v := sec.Key("CWAI_REASONING_EFFORT").String(); v != "" {
		cfg.ReasoningEffort = v
	}
	if v := sec.Key("CWAI_VERBOSITY").String(); v != "" {
		cfg.Verbosity = v
	}
	if v := sec.Key("CWAI_STRUCTURED_OUTPUT").String(); v != "" {
		cfg.StructuredOutput = v
	}

	return cfg, nil
}

func Set(key, value string) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	var f *ini.File
	if _, err := os.Stat(path); os.IsNotExist(err) {
		f = ini.Empty()
	} else {
		f, err = ini.Load(path)
		if err != nil {
			return fmt.Errorf("cannot load config %s: %w", path, err)
		}
	}

	f.Section("").Key(key).SetValue(value)

	return f.SaveTo(path)
}

func Get(key string) (string, error) {
	path, err := configPath()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", nil
	}

	f, err := ini.Load(path)
	if err != nil {
		return "", fmt.Errorf("cannot load config %s: %w", path, err)
	}

	return f.Section("").Key(key).String(), nil
}

func (c *Config) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("CWAI_API_KEY is not set. Run 'cwai setup' or 'cwai config set CWAI_API_KEY <key>'")
	}
	return nil
}
