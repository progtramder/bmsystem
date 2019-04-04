package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var privateData = struct {
	Domain        string `yaml:"domain"`
	AppId         string `yaml:"app_id"`
	AppSecret     string `yaml:"app_secret"`
	AdminPassword string `yaml:"admin_password"`
}{}

func InitPrivate() error {
	path := systembasePath + "/private.yaml"
	setting, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(setting, &privateData)
	if err != nil {
		return err
	}

	return nil
}
