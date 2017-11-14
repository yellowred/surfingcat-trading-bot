package main

import (
	"os"
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

