package main

import (
	"context"
	"log"
	"os"
	"time"

	_ "github.com/devxp-tech/demo-app/config"
	"github.com/devxp-tech/demo-app/controllers"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"google.golang.org/grpc/credentials"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// var serviceName string
var (
	serviceName  string //os.Getenv("SERVICE_NAME")
	collectorURL string //os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	insecure     string //os.Getenv("INSECURE_MODE")
)

func recordMetrics() {
	go func() {
		for {
			opsProcessed.Inc()
			time.Sleep(2 * time.Second)
		}
	}()
}

var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events",
	})
)

func initTracer() func(context.Context) error {

	secureOption := otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if len(insecure) > 0 {
		secureOption = otlptracegrpc.WithInsecure()
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(collectorURL),
		),
	)

	if err != nil {
		log.Fatal(err)
	}
	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		log.Printf("Could not set resources: ", err)
	}

	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(resources),
		),
	)
	return exporter.Shutdown
}

func main() {
	serviceName = "demo-app"
	collectorURL = "localhost:4317"
	insecure = "true"

	if os.Getenv("SERVICE_NAME") != "" {
		serviceName = os.Getenv("SERVICE_NAME")
	}

	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" {
		collectorURL = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	}

	if os.Getenv("INSECURE_MODE") != "" {
		serviceName = os.Getenv("INSECURE_MODE")
	}

	cleanup := initTracer()
	defer cleanup(context.Background())
	// Server
	log.Println("Starting server...")
	router := gin.New()
	p := ginprometheus.NewPrometheus("gin")
	p.Use(router)
	router.Use(otelgin.Middleware(serviceName))
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  true,
			"message": "Hello world for your app demo-app",
		})
	})

	router.GET("/health-check/liveness", controllers.HealthCheckLiveness)
	router.GET("/health-check/readiness", controllers.HealthCheckReadiness)
	router.Run()
}
