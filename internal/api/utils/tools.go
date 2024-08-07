// common.go

package utils

import (
	"c2c/internal/config"
	"c2c/internal/tools"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

func remove(element int, data []int) []int {
	lt := []int{}
	for _, v := range data {
		if v != element {
			lt = append(lt, v)
		}
	}
	return lt
}

func indexOf(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1 //not found.
}

func Contains(slist []string, value string) bool {
	for _, v := range slist {
		if v == value {
			return true
		}
	}
	return false
}

func LoadLogger(config *config.Config) error {
	tools.NewLogger(config)

	return nil
}

// LoadConfiguration loads a yaml config file into a Config struct.
func LoadConfiguration(file *string) (config.Config, error) {
	var config config.Config
	config.Init()
	configFile, err := ioutil.ReadFile(*file)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

func WriteConfiguration(filename string, config *config.Config) error {
	data, err := yaml.Marshal(*config)

	if err != nil {
		tools.Logger.Errorf("Error while marshalling %s => %s", config, err)
		return err
	}

	err = ioutil.WriteFile(filename, data, 0)

	if err != nil {
		tools.Logger.Errorf("Error while writing %s => %s", config, err)
		return err
	}

	return err
}
