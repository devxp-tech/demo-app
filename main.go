package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/devxp-tech/demo-app/controllers"
	"github.com/devxp-tech/demo-app/pkg/monitoring"
	"github.com/gin-gonic/gin"
	ginprometheus "github.com/zsais/go-gin-prometheus"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	monit := monitoring.New()
	app := "demo-app"

	// Server
	fmt.Println("Welcome to", app)
	log.Println("Starting server...")
	router := gin.New()
	p := ginprometheus.NewPrometheus("gin")
	p.Use(router)
	router.Use(otelgin.Middleware(monit.ServiceName))
	router.GET("/", func(c *gin.Context) {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()
		shutdown, err := monit.InitProvider()
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := shutdown(ctx); err != nil {
				log.Fatal("failed to shutdown TracerProvider: %w", err)
			}
		}()

		tracer := otel.Tracer("demo-app")

		// Attributes represent additional key-value descriptors that can be bound
		// to a metric observer or recorder.
		commonAttrs := []attribute.KeyValue{
			attribute.String("attrA", "chocolate"),
			attribute.String("attrB", "raspberry"),
			attribute.String("attrC", "vanilla"),
		}

		// work begins
		ctx, span := tracer.Start(
			ctx,
			"CollectorExporter-Example",
			trace.WithAttributes(commonAttrs...))
		defer span.End()
		c.JSON(200, gin.H{
			"status":  true,
			"message": "Hello world for your fuck app is running demo-app",
		})
	})
	router.GET("/health-check/liveness", controllers.HealthCheckLiveness)
	router.GET("/health-check/readiness", controllers.HealthCheckReadiness)
	router.Run()
}
