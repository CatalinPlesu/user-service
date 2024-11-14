package application

import (
	"os"
	"strconv"
)

type Config struct {
	RedisAddress     string
	PostgresAddress  string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	ServerPort       uint16
}

func LoadConfig() Config {
	cfg := Config{
		RedisAddress:    "localhost:6379",
		PostgresAddress: "localhost:5432",
		PostgresUser:    "postgres",
		PostgresPassword: "password",
		PostgresDB:      "user_service",
		ServerPort:      3000,
	}

	if redisAddr, exists := os.LookupEnv("REDIS_ADDR"); exists {
		cfg.RedisAddress = redisAddr
	}

	if postgresAddr, exists := os.LookupEnv("POSTGRES_ADDR"); exists {
		cfg.PostgresAddress = postgresAddr
	}

	if postgresUser, exists := os.LookupEnv("POSTGRES_USER"); exists {
		cfg.PostgresUser = postgresUser
	}

	if postgresPassword, exists := os.LookupEnv("POSTGRES_PASSWORD"); exists {
		cfg.PostgresPassword = postgresPassword
	}

	if postgresDB, exists := os.LookupEnv("POSTGRES_DB"); exists {
		cfg.PostgresDB = postgresDB
	}

	if serverPort, exists := os.LookupEnv("SERVER_PORT"); exists {
		if port, err := strconv.ParseUint(serverPort, 10, 16); err == nil {
			cfg.ServerPort = uint16(port)
		}
	}

	return cfg
}
