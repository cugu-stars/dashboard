package main

import (
	"flag"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

func parseInput() (config Config, err error) {
	yamlFile, err := ioutil.ReadFile(flag.Args()[0])
	if err != nil {
		return
	}
	return config, yaml.Unmarshal(yamlFile, &config)
}
