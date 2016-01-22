package model

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	AppName       string
	IsCluster     bool
	IsMaster      bool
	ThreadNum     int
	LockerTimeout int
	ZkTimeOut     int
	ZkHosts       []string
}

func GetConfigFromPath(path string) (*Config, error) {
	var config Config
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return &config, err
	}
	err = json.Unmarshal(data, &config)
	return &config, err
}
