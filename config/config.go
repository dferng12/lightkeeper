package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Mount struct {
	ID      int    `yaml:"ID"`
	Type    string `yaml:"type"`
	From    string `yaml:"from"`
	DstPath string `yaml:"dstpath"`
}

// Container contains the config to deploy the container
type Container struct {
	Name   string         `yaml:"name"`
	Image  string         `yaml:"image"`
	Ports  map[int]string `yaml:"ports"`
	Mounts []Mount        `yaml:"mounts"`
	Env    []string       `yaml:"env"`
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
func loadConfig() (cfg Configuration) {
	f, err := os.Open("conf.yml")
	checkErr(err)
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	checkErr(err)
	return
}

func GetContainerConfig(containerName string) Container {
	var containerConfig Container
	config := loadConfig()

	for _, config := range config.Containers {
		if "/"+config.Name == containerName {
			containerConfig = config
			break
		}
	}

	if containerConfig.Name == "" {
		panic("Container config not found")
	}

	return containerConfig
}

// ConfigData is the ready to be used object
