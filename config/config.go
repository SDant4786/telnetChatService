package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TelNetIp   string
	TelNetPort string
	HttpIp     string
	HttpPort   string
	LogFile    string
}

func LoadConfig(filepath string) (Config, error) {
	//Load file
	err := godotenv.Load(filepath)
	if err != nil {
		return Config{}, nil
	}

	//Get values
	telNetIP := os.Getenv("TELNET_IP")
	if telNetIP == "" {
		return Config{}, errors.New("TELNET_IP missing")
	}

	telNetPort := os.Getenv("TELNET_PORT")
	if telNetPort == "" {
		return Config{}, errors.New("TELNET_PORT missing")
	}

	httpIp := os.Getenv("HTTP_IP")
	if httpIp == "" {
		return Config{}, errors.New("HTTP_IP missing")
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		return Config{}, errors.New("HTTP_PORT missing")
	}

	logFile := os.Getenv("LOG_FILE")
	if logFile == "" {
		return Config{}, errors.New("LOG_FILE missing")
	}
	//Return config struct
	return Config{
		TelNetIp:   telNetIP,
		TelNetPort: telNetPort,
		HttpIp:     httpIp,
		HttpPort:   httpPort,
		LogFile:    logFile,
	}, nil
}
