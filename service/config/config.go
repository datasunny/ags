package config

import (
	"encoding/json"
	"fmt"
	log "github.com/featen/utils/log"
	"io/ioutil"
)

var configFile string
var sysConfig map[string]string

func getConfigs() {
	b, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal("Read test configuration file failed: ", err)
		fmt.Println("read file failed")
	}

	err = json.Unmarshal(b, &sysConfig)
	if err != nil {
		log.Fatal("Unmarshal test configuration failed: ", err)
		fmt.Println("unmarshal failed")
	}
}

func InitConfigs(cf string) {
	configFile = cf
	getConfigs()
}

func saveConfigs() {
	b, err := json.MarshalIndent(sysConfig, "", "  ")
	if err != nil {
		log.Fatal("not be able to save configs")
	}
	ioutil.WriteFile(configFile, b, 0644)
}

func GetValue(key string) string {
	if val, ok := sysConfig[key]; !ok {
		return ""
	} else {
		return val
	}
}

func SetValue(key, value string) {
	sysConfig[key] = value
	saveConfigs()
	fmt.Println(sysConfig)
}

func IsConfigInited() bool {
	value := GetValue("dbInited")
	if value == "Y" {
		return true
	} else {
		return false
	}
}
