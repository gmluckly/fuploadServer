// Copyright © 19-11-12 Shenzhen  Technology CO., LTD. All rights reserved.
//
// @author: GM
// @date:   2019/11/12
//

package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

var config Config

type Config struct {
	Env    string `yaml:"env"`
	Server struct {
		Port      int    `yaml:"port"`
		Addr      string `yaml:"addr"`
		ProxyAddr string `yaml:"proxyAddr"`
	}
	cacheAddr   string             `yaml:"cacheAddr"`
	MysqlDB     mysqlConf          `yaml:"mysqlDB"`
	CheckTmpMd5 bool               `yaml:"checkTmpMd5"`
	TmpDir      string             `yaml:"tmpDir"`
	StoreDir    string             `yaml:"storeDir"`
	TaskTimeout time.Duration      `yaml:"taskTimeout"`
	BServer     map[string]bServer `yaml:"bServer"`
}

type mysqlConf struct {
	Host     string `yaml:"host"`
	Database string `yaml:"database"`
	UserName string `yaml:"userName"`
	Password string `yaml:"password"`
}

type bServer struct {
	//Name      string `yaml:"name"`
	Name      string `yaml:"name"`
	NotifyUrl string `yaml:"notifyUrl"`
}

func NewConfig(confPath string) (*Config, error) {

	f, err := os.Open(confPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	d, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return NewConfigFromBytes(d)
}

func NewConfigFromBytes(content []byte) (*Config, error) {
	c := &Config{}
	err := yaml.Unmarshal(content, c)
	if err != nil {
		fmt.Println("yaml unmarshal err:", err)
		return nil, err
	}
	config = *c
	return c, nil
}

func GetConfig() Config {
	return config
}
