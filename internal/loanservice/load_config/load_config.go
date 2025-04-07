package loadconfig
import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

)

type HTTPConfig struct {
    Port string `yaml:"port"`
}

type GRPCConfig struct {
    Port string `yaml:"port"`
}

type Config struct {
    HTTP HTTPConfig `yaml:"http"`
    GRPC GRPCConfig `yaml:"grpc"`
}

func LoadConfig(path string) (*Config, error) {
    config := &Config{
        HTTP: HTTPConfig{Port: "8080"},
        GRPC: GRPCConfig{Port: "50051"},
    }

    file, err := os.ReadFile(path)
    if err != nil {
        return config, fmt.Errorf("error reading config file: %v, using defaults", err)
    }

    err = yaml.Unmarshal(file, config)
    if err != nil {
        return config, fmt.Errorf("error parsing config file: %v, using defaults", err)
    }

    return config, nil
}
