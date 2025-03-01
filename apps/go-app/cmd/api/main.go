package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	telemetry "github/hhertout/otel-example"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	if os.Getenv("OTEL_LOG_ENABLED") == "true" {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()

		shutdown := telemetry.InitOtel(ctx)
		defer func() {
			if err := shutdown(ctx); err != nil {
				log.Fatalf("failed to shutdown: %s", err)
			}
		}()

		p := telemetry.InitPyroscope()
		defer p.Stop()
	}

	r := gin.Default()
	r.Use(telemetry.OtelMiddleware())
	r.Use(gin.Logger())

	r.GET("/ping", func(c *gin.Context) {
		log.Println("🚀 Request received")

		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080
}
