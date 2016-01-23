package model

import (
	"encoding/json"
	"io/ioutil"
)

type DistributedConfig struct {
	AppName       string
	IsMaster      bool
	ThreadNum     int
	LockerTimeout int
	ZkTimeOut     int
	ZkHosts       []string
}

func GetDistributedConfigFromPath(path string) (*DistributedConfig, error) {
	var config DistributedConfig
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return &config, err
	}
	err = json.Unmarshal(data, &config)
	return &config, err
}
