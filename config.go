package main

import (
	"encoding/json"
	"io"
	"os"
	"log"
)

type OthersType struct {
	ExternalPort   string `json:"external_port"`
	ExternalIp string `json:"external_ip"`
	InternalIp   string `json:"internal_ip"`
	InternalPort string `json:"internal_port"`
	Protocol string `json:"protocol"`
}

type ServerType struct {
	ExternalPort   string `json:"external_port"`
	ExternalIp string `json:"external_ip"`
	InternalIp   string `json:"internal_ip"`
	InternalPort string `json:"internal_port"`
	Protocol string `json:"protocol"`
	Others []OthersType `json:"others"`
}

type Config struct {
	ApiUrl string `json:"api_url"`
	Username string `json:"username"`
	Password string `json:"password"`
	Timeout int `json:"timeout"`
	AutoShutdown bool `json:"auto_shutdown"`
	Addresses []ServerType `json:"addresses"`
	Blocks []string `json:"blocks"`
	Allows []string `json:"allows"`
	AbuseIPDBKey string `json:"abuse_ipdb_key"`
}

func loadConfig() Config {
	file, err := os.Open("./config.json")
	if err != nil {
		_, err = os.Create("./config.json")
		if(err!=nil){
			panic("Could not open config\n");
		}

		err = os.Chmod("./config.json", 0755)
		if (err !=nil){
			panic("Created a file but could not chmod it\n")
		}
		
		panic("Could not open config\nCreated one config file before exiting\n");
	}

	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		panic("Could not read config\n");
	}

	var config Config;
	err = json.Unmarshal(byteValue, &config)

	if err != nil{
		panic("Could not parse config\n");
	}

	return config;
}

var singleConfigInstance *Config

func getConfig() *Config{
	if singleConfigInstance == nil {
		var config = loadConfig();
		singleConfigInstance = &config;
	}

	return singleConfigInstance;
}

func addIpBlock(ip string) {
	config := getConfig()
	for _, blockedIp := range config.Blocks {
		if blockedIp == ip {
			return // IP is already blocked
		}
	}
	config.Blocks = append(config.Blocks, ip)
	err := saveConfig(*config)
	if err != nil {
		log.Printf("Error saving config after adding IP block: %s", err)
	}
}

func addIpAllow(ip string) {
	config := getConfig()
	for _, allowedIp := range config.Allows {
		if allowedIp == ip {
			return // IP is already allowed
		}
	}
	config.Allows = append(config.Allows, ip)
	err := saveConfig(*config)
	if err != nil {
		log.Printf("Error saving config after adding IP allow: %s", err)
	}
}

func saveConfig(config Config) error {
	file, err := os.Create("./config.json")
	if err != nil {
		return err
	}
	defer file.Close()
	
	byteValue, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	_, err = file.Write(byteValue)

	if err == nil {
        // Update the singleton instance to the new config
        singleConfigInstance = &config
    }

	return err
}
