package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func Load() error {
	return godotenv.Load()
}


func init(){
	Load()
}

func GetString(key string) string{
	value,ok := os.LookupEnv(key)

	if !ok{
		return ""
	}
	return value
}

func GetInt(key string) int{
	value,ok := os.LookupEnv(key)

	if !ok{
		return 0
	}
	intValue,err := strconv.Atoi(value)

	if err!=nil{
		return 0
	}
	
	return intValue
}