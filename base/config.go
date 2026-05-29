package base

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/kylelemons/go-gypsy/yaml"
)

var env *ConfigEnv

type ConfigEnv struct {
	configFile *yaml.File
}

func GetConfig() *ConfigEnv {
	if env == nil {
		filePath := fmt.Sprintf("config/env.yml")
		env = NewEnv(filePath)
	}
	return env
}

func NewEnv(configFile string) *ConfigEnv {
	env := &ConfigEnv{
		configFile: yaml.ConfigFile(configFile),
	}
	if env.configFile == nil {
		panic("go-configenv failed to open configFile: " + configFile)
	}
	return env
}

func (env *ConfigEnv) Get(spec, defaultValue string) string {
	value, err := env.configFile.Get(spec)
	if err != nil {
		value = defaultValue
	}
	return value
}

func (env *ConfigEnv) GetInt(spec string, defaultValue int) int {
	str := env.Get(spec, "")
	if str == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(str)
	if err != nil {
		log.Panic("go-configenv GetInt failed Atoi", spec, str)
	}
	return val
}

func (env *ConfigEnv) GetDuration(spec string, defaultValue string) time.Duration {
	str := env.Get(spec, "")
	if str == "" {
		str = defaultValue
	}
	duration, err := time.ParseDuration(str)
	if err != nil {
		log.Panic("go-configenv GetDuration failed ParseDuration", spec, str)
	}
	return duration
}

// WorkerConfig 表示 config/env.yml 中单个 worker 条目的配置
type WorkerConfig struct {
	Name          string
	DefaultPrefix bool
	Queue         string
	Log           string
	Threads       int
	Durable       bool
}

// GetWorkerConfigs 从 config/env.yml 读取 worker 配置列表
func (env *ConfigEnv) GetWorkerConfigs() []WorkerConfig {
	count, err := env.configFile.Count("worker")
	if err != nil || count <= 0 {
		return nil
	}

	node, err := yaml.Child(env.configFile.Root, "worker")
	if err != nil {
		return nil
	}

	list, ok := node.(yaml.List)
	if !ok {
		return nil
	}

	var configs []WorkerConfig
	for i := 0; i < list.Len(); i++ {
		item := list.Item(i)
		m, ok := item.(yaml.Map)
		if !ok {
			continue
		}

		wc := WorkerConfig{
			Durable: true, // 默认持久化，避免 transient_nonexcl_queues 错误
		}

		if v := m.Key("name"); v != nil {
			if s, ok := v.(yaml.Scalar); ok {
				wc.Name = s.String()
			}
		}
		if v := m.Key("defaultPrefix"); v != nil {
			if s, ok := v.(yaml.Scalar); ok {
				wc.DefaultPrefix = s.String() == "true"
			}
		}
		if v := m.Key("queue"); v != nil {
			if s, ok := v.(yaml.Scalar); ok {
				wc.Queue = s.String()
			}
		}
		if v := m.Key("log"); v != nil {
			if s, ok := v.(yaml.Scalar); ok {
				wc.Log = s.String()
			}
		}
		if v := m.Key("threads"); v != nil {
			if s, ok := v.(yaml.Scalar); ok {
				if n, err := strconv.Atoi(s.String()); err == nil && n > 0 {
					wc.Threads = n
				} else {
					wc.Threads = 1
				}
			}
		}
		if v := m.Key("durable"); v != nil {
			if s, ok := v.(yaml.Scalar); ok {
				wc.Durable = s.String() == "true"
			}
		}

		configs = append(configs, wc)
	}

	return configs
}
