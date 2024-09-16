package main

import (
	"context"
	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"log"
	"net/http"
)

func initTracer() (*trace.TracerProvider, error) {
	ctx := context.Background()
	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint("localhost:4318"), otlptracehttp.WithInsecure())
	if err != nil {
		return nil, err
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithSampler(trace.AlwaysSample()), // AlwaysSample sampler
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("service-a"),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

func main() {
	tp, err := initTracer()
	if err != nil {
		log.Fatalf("failed to initialize tracer: %v", err)
	}
	defer func() { _ = tp.Shutdown(context.Background()) }()

	app := fiber.New()

	app.Use(otelfiber.Middleware())

	app.Get("/call-service-b", func(c *fiber.Ctx) error {
		tracer := otel.Tracer("service-a")
		ctx, span := tracer.Start(c.UserContext(), "callServiceB")
		defer span.End()

		correlationID := uuid.New().String()

		client := http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		}

		req, _ := http.NewRequestWithContext(ctx, "GET", "http://localhost:3001/orders/11", nil)
		req.Header.Set("x-correlationid", correlationID)

		resp, err := client.Do(req)
		if err != nil {
			span.RecordError(err)
			return c.Status(http.StatusInternalServerError).SendString("Failed to call Service B")
		}
		defer resp.Body.Close()

		span.SetAttributes(attribute.String("response_status", resp.Status))

		return c.SendString("Called service B successfully")
	})

	log.Fatal(app.Listen(":3000"))
}
