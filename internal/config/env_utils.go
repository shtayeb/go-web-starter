package config

import (
	"log"
	"net/url"
	"os"
	"strconv"
)

func GetEnv(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return defaultVal
}

func GetEnvAsInt(key string, defaultVal int) int {
	strVal := GetEnv(key, "")

	if val, err := strconv.Atoi(strVal); err == nil {
		return val
	}

	return defaultVal
}

func GetEnvAsBool(key string, defaultVal bool) bool {
	strVal := GetEnv(key, "")

	if val, err := strconv.ParseBool(strVal); err == nil {
		return val
	}

	return defaultVal
}

func GetEnvAsURL(key string, defaultVal string) *url.URL {
	strVal := GetEnv(key, "")

	if len(strVal) == 0 {
		u, err := url.Parse(defaultVal)
		if err != nil {
			log.Panicf("Failed to parse default value %s for env variable %s as URL: %v", defaultVal, key, err)
		}

		return u
	}

	u, err := url.Parse(strVal)
	if err != nil {
		log.Panicf("Failed to parse env variable %s with value %s as URL: %v", key, strVal, err)
	}

	return u
}
