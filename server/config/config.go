package config

import (
	"io/ioutil"
	"os"

	"github.com/spf13/viper"
)

func BittrexApiKeys() (string, string) {
	return env("BITTREX_PUBLIC_KEY", ""), env("BITTREX_PRIVATE_KEY", "")
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func StrategyConfig(name string) map[string]string {
	viper.SetConfigType("json")
	file, err := os.Open("config/trading.json")
	if err != nil {
		panic(err)
	}
	viper.ReadConfig(file)
	return viper.GetStringMapString("strategies." + name)
}

func ExchangeConfig(name string) map[string]string {
	viper.SetConfigType("json")
	file, err := os.Open("config/config.json")
	if err != nil {
		panic(err)
	}
	viper.ReadConfig(file)
	return viper.GetStringMapString("exchanges." + name)
}

func TestbedFile(name string) []byte {
	viper.SetConfigType("json")
	file, err := os.Open("config/testbeds.json")
	if err != nil {
		panic(err)
	}
	viper.ReadConfig(file)
	testbedsDir := viper.GetString("dir")
	if testbedsDir == "" {
		panic("Testbed dir is not defined.")
	}
	files := viper.GetStringMapString("files." + name)
	if fileName, ok := files["name"]; ok {
		dat, err := ioutil.ReadFile(testbedsDir + "/" + fileName)
		if err != nil {
			panic(err)
		}
		return dat
	} else {
		panic("Testbed not found.")
	}
}

func TestbedMarket(name string) string {
	viper.SetConfigType("json")
	file, err := os.Open("config/testbeds.json")
	if err != nil {
		panic(err)
	}
	viper.ReadConfig(file)
	testbedsDir := viper.GetString("dir")
	if testbedsDir == "" {
		panic("Testbed dir is not defined.")
	}
	files := viper.GetStringMapString("files." + name)
	if market, ok := files["market"]; ok {
		return market
	} else {
		panic("Testbed not found.")
	}
}
