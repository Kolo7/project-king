package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/Kolo7/project-king/pkg/cache"
	"github.com/Kolo7/project-king/pkg/db"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"
)

var (
	serviceName  = os.Getenv("SERVICE_NAME")
	signozToken  = os.Getenv("SIGNOZ_ACCESS_TOKEN")
	collectorURL = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	insecure     = os.Getenv("INSECURE_MODE")

	rdAddr = "0.0.0.0:6379"
	rdPwd  = "gHmNkVBd88sZybj"
	dsn    = "root:123456@tcp(localhost:3306)/chatgpt_web?charset=utf8mb4&parseTime=True&loc=Local"

	tracer trace.Tracer
)

func main() {
	cleanup := initTracer()
	defer cleanup(context.Background())

	initCustomTracer()

	r := gin.Default()
	r.Use(otelgin.Middleware(serviceName))

	r.GET("/book", func(c *gin.Context) {
		ctx, _ := tracer.Start(c.Request.Context(), "custom", []trace.SpanStartOption{
			trace.WithSpanKind(trace.SpanKindInternal),
			trace.WithAttributes(
				attribute.String("hah", "hello"),
			),
		}...)
		var (
			count int
			err   error
		)
		n := rand.Intn(100)
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(attribute.String("a", c.Query("a")), attribute.String("b", c.Query("b")))
		defer span.End()
		if n < 10 {
			c.String(500, "%s", "模拟业务错误")
			err = errors.New("custom error")
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return
		}
		if n >= 10 && n <= 20 {
			err = db.GetDB(dsn).WithContext(ctx).Raw("select count(*) from user11").Scan(&count).Error
		} else {
			err = db.GetDB(dsn).WithContext(ctx).Raw("select count(*) from user").Scan(&count).Error
		}
		if err != nil {
			log.Print(err)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			c.String(500, "模拟数据库错误，%v", err)
			return
		}
		if n >= 21 && n <= 30 {
			err = cache.GetDBCache(rdAddr, rdPwd).Eval(ctx, "hmp", []string{}, []interface{}{}).Err()
		} else {
			err = cache.GetDBCache(rdAddr, rdPwd).Set(ctx, "kkkk", fmt.Sprintf("%d", count), time.Second).Err()
		}
		if err != nil {
			log.Print(err)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			c.String(500, "模拟redis错误，%v", err)
			return
		}

		c.String(200, "%d", count)
	})

	if err := r.Run(":8000"); err != nil {
		db.GetDB(dsn)
		log.Fatal(err)
	}
}

func initTracer() func(context.Context) error {

	headers := map[string]string{
		"signoz-access-token": signozToken,
	}

	secureOption := otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if len(insecure) > 0 {
		secureOption = otlptracegrpc.WithInsecure()
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(collectorURL),
			otlptracegrpc.WithHeaders(headers),
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
			// sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.2)),
			sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter)),
			sdktrace.WithSyncer(exporter),
			sdktrace.WithResource(resources),
		),
	)
	return exporter.Shutdown
}

func initCustomTracer() {
	tracer = otel.GetTracerProvider().Tracer("custom", trace.WithInstrumentationVersion("v0.0.1"))

}
