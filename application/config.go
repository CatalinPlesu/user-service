package application

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort       uint16
	RedisAddress     string
	PostgresAddress  string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	RabitMQURL       string
}

func LoadConfig() Config {
	cfg := Config{
		ServerPort:       3000,
		RedisAddress:     "localhost:6379",
		PostgresAddress:  "localhost:5432",
		PostgresUser:     "user",
		PostgresPassword: "password",
		PostgresDB:       "user_service_db",
		RabitMQURL:		  "amqp://guest:guest@localhost:5672/",
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

	if rabitMQURL, exists := os.LookupEnv("RABITMQ_URL"); exists {
		cfg.RabitMQURL = rabitMQURL
	}

	if serverPort, exists := os.LookupEnv("SERVER_PORT"); exists {
		if port, err := strconv.ParseUint(serverPort, 10, 16); err == nil {
			cfg.ServerPort = uint16(port)
		}
	}

	return cfg
}
