package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/drd-engineering/TwinCape/db"
	"github.com/drd-engineering/TwinCape/domains"
	"github.com/drd-engineering/TwinCape/environments"
	"github.com/drd-engineering/TwinCape/routes"
)

func makeDbConfig() *db.Config {
	return &db.Config{
		Host:     environments.Get("HOST_DB"),
		Username: environments.Get("USERNAME_DB"),
		DBName:   environments.Get("DB_NAME"),
		Password: environments.Get("PASSWORD_DB"),
	}
}
func getRoutingPort() string {
	return environments.Get("PORT")
}

func main() {
	var port string
	// Store the release type this engine will be run
	releaseType := flag.String("release", "localhost", "to define release type you are running this command, default value : localhost")
	if releaseType != nil {
		environments.Set("RELEASE_TYPE", strings.ToLower(*releaseType))
	} else {
		environments.Set("RELEASE_TYPE", strings.ToLower("localhost"))
	}
	environments.LoadEnvironmentVariableFile()
	err := db.InitPostgre(makeDbConfig())
	if err != nil {
		fmt.Println("POSTGRE is not started, there is something wrong with environment variable")
		return
	}
	port = getRoutingPort()
	// Add Specific router group to main router
	domains.InitiateRoutes()

	// Start Server
	r := routes.GetInstance()
	r.Run(":" + port)
}
