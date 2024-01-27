package main

import (
	"github.com/pelletier/go-toml/v2"
	"os"
)

type SignalingResteamConfig struct {
	Host string
	Port uint16
}

type IceRestreamConfig struct {
	Stun    string
	Turn    string
	TurnUsr string `toml:"turn_usr"`
	TurnPwd string `toml:"turn_pwd"`
}

type JwtRestreamConfig struct {
	Public string
}

type StreamRestreamConfig struct {
	Id  string
	Url string
}

type RestreamConfig struct {
	Signaling SignalingResteamConfig
	Ice       IceRestreamConfig
	Jwt       JwtRestreamConfig
	Stream    []StreamRestreamConfig
}

func LoadConfig() RestreamConfig {
	configPath := os.Getenv("CONFIG")

	if configPath == "" {
		configPath = "config.toml"
	}

	bytes, err := os.ReadFile(configPath)
	if err != nil {
		panic(err)
	}

	var config RestreamConfig
	tomlErr := toml.Unmarshal(bytes, &config)
	if tomlErr != nil {
		panic(tomlErr)
	}

	return config
}
