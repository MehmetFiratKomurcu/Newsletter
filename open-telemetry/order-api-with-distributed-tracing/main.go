package main

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	_ "yt-opentelemetry-tracing-blueprint/docs"
	"yt-opentelemetry-tracing-blueprint/src/application/controller"
	"yt-opentelemetry-tracing-blueprint/src/application/domain/persistance"
	"yt-opentelemetry-tracing-blueprint/src/application/domain/services"
	"yt-opentelemetry-tracing-blueprint/src/infra/Trace"
	"yt-opentelemetry-tracing-blueprint/src/infra/middleware"
	"yt-opentelemetry-tracing-blueprint/src/infra/validation"
)

var validate = validator.New()

// @title			Order Api
// @version		1.0
// @description	This is an Order Api just for young people
// @termsOfService	http://swagger.io/terms/
func main() {
	tp, err := Trace.InitTracer()
	if err != nil {
		log.Fatalf("failed to initialize tracer: %v", err)
	}
	defer func() { _ = tp.Shutdown(context.Background()) }()

	app := fiber.New()

	app.Use(otelfiber.Middleware())

	app.Get("/swagger/*", swagger.HandlerDefault)

	connString := "host=localhost user=postgres password=mysecretpassword dbname=postgres port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(connString), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	if err := db.Use(otelgorm.NewPlugin()); err != nil {
		panic(err)
	}

	customValidator := &validation.CustomValidator{
		Validator: validate,
	}

	app.Use(recover.New())
	app.Use(cors.New())

	// middleware
	middleware.AddCorrelationId(app)

	// repositories
	orderRepository := persistance.NewOrderRepository(db)

	// services
	orderService := services.NewOrderService(orderRepository)

	// endpoints
	controller.GetOrderById(app, orderService)
	controller.CreateOrder(app, customValidator, orderService)
	controller.ShipOrderByCargoCode(app, orderService)

	app.Listen(":3001")
}
