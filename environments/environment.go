package environments

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnvironmentVariableFile should be run once before you use the environment variable
func LoadEnvironmentVariableFile() {
	releaseType := os.Getenv("RELEASE_TYPE")
	fileEnvLocation := "./environments/.env." + releaseType
	err := godotenv.Load(fileEnvLocation)
	if err != nil {
		fmt.Println("Error loading " + fileEnvLocation + " file")
	}
}

// Get environment variable from environments file loaded
func Get(key string) string {
	return os.Getenv(key)
}

// Set additional environment variable needed
func Set(key string, value string) {
	os.Setenv(key, value)
}
