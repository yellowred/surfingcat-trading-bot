package main

import (
	"os"
	"github.com/spf13/viper"
)

func BittrexApiKeys() (string, string) {
	return env("BITTREX_PUBLIC_KEY", ""), env("BITTREX_PRIVATE_KEY", "")
}

func ApiPort() string {
	return env("API_PORT", "3026")
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func StrategyConfig(name string) (map[string]string, error) {
	viper.SetConfigType("json")
	file, err := os.Open("config/trading.json")
	if err != nil { return nil, err }	
	viper.ReadConfig(file)
	return viper.GetStringMapString("strategies." + name), nil
}

func ExchangeConfig(name string) (map[string]string, error) {
	viper.SetConfigType("json")
	file, err := os.Open("config/config.json")
	if err != nil { return nil, err }	
	viper.ReadConfig(file)
	return viper.GetStringMapString("exchanges." + name), nil
}

