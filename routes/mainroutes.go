package routes

import (
	"fmt"
	"sync"
	"time"

	"github.com/drd-engineering/TwinCape/db"
	"github.com/drd-engineering/TwinCape/environments"
	"github.com/gin-gonic/gin"
)

var instance *gin.Engine
var once sync.Once

// GetInstance will return gin engine that already been setup once
func GetInstance() *gin.Engine {
	// Initiate value if there is no instance
	once.Do(func() {
		instance = gin.New()

		// LoggerWithFormatter middleware will write the logs to gin.DefaultWriter
		// By default gin.DefaultWriter = os.Stdout
		instance.Use(gin.LoggerWithFormatter(auditRailsLogger))
		instance.Use(environmentUpdater())

		instance.Use(gin.Recovery())
		instance.Use(CORSMiddleware())
	})
	return instance
}

// CORSMiddleware controlling access for API
func CORSMiddleware() gin.HandlerFunc {
	// adding header to define the application detail
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Access-Control-Allow-Origin, Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Drd-Identification, Authorization")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Next()
	}
}

func environmentUpdater() gin.HandlerFunc {
	return func(c *gin.Context) {
		environments.LoadEnvironmentVariableFile()
		c.Next()
	}
}

func auditRailsLogger(param gin.LogFormatterParams) string {
	// save the log to db, then also return the log to default logger
	apiLog := db.APILog{
		Timestamp:      param.TimeStamp,
		TTL:            param.Latency.String(),
		ResponseStatus: param.StatusCode,
		Path:           param.Path,
		Method:         param.Method,
		ClientIP:       param.ClientIP,
		ClientTools:    param.Request.UserAgent(),
		Protocol:       param.Request.Proto,
	}
	dbInstance := db.GetDb()
	dbInstance.Create(&apiLog)

	return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
		param.ClientIP,
		param.TimeStamp.Format(time.RFC1123),
		param.Method,
		param.Path,
		param.Request.Proto,
		param.StatusCode,
		param.Latency,
		param.Request.UserAgent(),
		param.ErrorMessage,
	)
}
