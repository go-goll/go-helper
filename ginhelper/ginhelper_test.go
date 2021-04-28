package ginhelper

import (
	"github.com/gin-gonic/gin"
	"github.com/go-goll/go-helper/loghelper"
	"github.com/rs/zerolog"
	"testing"
	"time"
)

func TestGinLogMiddleware(t *testing.T) {
	e := gin.New()
	logger := loghelper.GetLogger(loghelper.LogTargetStdout)
	e.Use(GinLogMiddleware(logger))

	e.GET("/", ginHandler)

	e.Run(":8081")
}

func ginHandler(c *gin.Context) {
	start := time.Now()
	logger := zerolog.Ctx(c.Request.Context())
	logger.Info().Msg("Start processing...")

	logger.Info().Str("id", "userapp01").Str("name", "admin").Msg("")
	//body, _ := ioutil.ReadAll(c.Request.Body)
	//logmiddleware.Debug().RawJSON("body", body).Msg("")

	c.JSON(200, gin.H{"data": start})

	logger.Info().Dur("elapsed", time.Since(start)).Msg("Done")
}

func TestExampleMain(t *testing.T) {
	ExampleMain()
}
