// config/config.go

package config

type Config struct {
	ServerPort string
}

func LoadConfig() Config {
	return Config{
		ServerPort: "8081", // Default port
	}
}
