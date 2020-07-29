package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// Container contains the config to deploy the container
type Container struct {
	Name   string      `yaml:"name"`
	Image  string      `yaml:"image"`
	Ports  map[int]int `yaml:"ports"`
	Mounts []struct {
		Type    string `yaml:"type"`
		From    string `yaml:"from"`
		DstPath string `yaml:"dstpath"`
	} `yaml:"mounts"`
	Env   map[string]string `yaml:"env"`
	Flags map[string]string
}

// Configuration contains the yaml data of the config file
type Configuration struct {
	Containers []Container `yaml:"containers"`
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// LoadConfig loads the yaml configuration file and returns a struct ready to be read
func LoadConfig() (cfg Configuration) {
	f, err := os.Open("conf.yml")
	checkErr(err)
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	checkErr(err)
	return
}

// ConfigData is the ready to be used object
var ConfigData = LoadConfig()
