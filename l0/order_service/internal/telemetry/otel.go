package telemetry

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.23.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	Tracer trace.Tracer

	HttpRequestCounter  metric.Int64Counter
	HttpRequestDuration metric.Float64Histogram

	DbOperationsCounter metric.Int64Counter
	DbOperationDuration metric.Float64Histogram

	CacheHitsCounter metric.Int64Counter
	CacheMissCounter metric.Int64Counter

	KafkaMessageCounter metric.Int64Counter
	KafkaErrorsCounter  metric.Int64Counter
)

func InitOTel(serviceName string) func() {
	initCtx := context.Background()

	// make resource with info about service
	res, err := resource.New(initCtx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			attribute.String("enviroment", "production"),
		),
	)
	if err != nil {
		log.Printf("Warning: failed to create resource: %v", err)
	}

	// traces(jaeger)
	// exporter
	traceExporter, err := jaeger.New(jaeger.WithCollectorEndpoint(
		jaeger.WithEndpoint("http://jaeger:14268/api/traces"),
	))
	if err != nil {
		// сервис работает без трейсов, не return тк еще метрики
		log.Printf("Warning: jaeger not available (traces disabled): %v", err)
	} else {
		// provider
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(traceExporter),
			sdktrace.WithResource(res),
		)
		otel.SetTracerProvider(tp)
		Tracer = tp.Tracer(serviceName)
	}

	// metrics(prometheus)
	// exporter
	metricExporter, err := prometheus.New()
	if err != nil {
		log.Printf("Warning: failed to create prometheus exporter: %v", err)
		// если метрики не работают - пустая функция cleanup
		return func() {}
	}
	// provider
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(metricExporter),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)
	meter := mp.Meter(serviceName)

	initMetrics(meter)

	log.Println("OpenTelemetry initialized successfully")
	// cleanup func
	return func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// если трейсер был создан
		if tp, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider); ok && tp != nil {
			if err := tp.Shutdown(shutdownCtx); err != nil {
				log.Printf("Error shutting down tracer provider: %v", err)
			}
		}

		// если метр провайдер был создан
		if mp, ok := otel.GetMeterProvider().(*sdkmetric.MeterProvider); ok && mp != nil {
			if err := mp.Shutdown(shutdownCtx); err != nil {
				log.Printf("Error shutting down meter provider: %v", err)
			}
		}
	}
}

// создает метрики
func initMetrics(meter metric.Meter) {
	var err error
	HttpRequestCounter, err = meter.Int64Counter("http.requests.total",
		metric.WithDescription("Total number of HTTP requests"))
	if err != nil {
		log.Printf("Warning: failed to create http requests counter: %v", err)
	}

	HttpRequestDuration, err = meter.Float64Histogram("http.request.duration.seconds",
		metric.WithDescription("Duration of HTTP requests"),
		metric.WithUnit("s"))
	if err != nil {
		log.Printf("Warning: failed to create http request duration: %v", err)
	}

	DbOperationsCounter, err = meter.Int64Counter("db.operations.total",
		metric.WithDescription("Total number of database operations"))
	if err != nil {
		log.Printf("Warning: failed to create db operations counter: %v", err)
	}

	DbOperationDuration, err = meter.Float64Histogram("db.operation.duration.seconds",
		metric.WithDescription("Duration of database operations"),
		metric.WithUnit("s"))
	if err != nil {
		log.Printf("Warning: failed to create db operation duration: %v", err)
	}

	CacheHitsCounter, err = meter.Int64Counter("cache.hits.total",
		metric.WithDescription("Total number of cache hits"))
	if err != nil {
		log.Printf("Warning: failed to create cache hits counter: %v", err)
	}

	CacheMissCounter, err = meter.Int64Counter("cache.misses.total",
		metric.WithDescription("Total number of cache misses"))
	if err != nil {
		log.Printf("Warning: failed to create cache misses counter: %v", err)
	}

	KafkaMessageCounter, err = meter.Int64Counter("kafka.messages.total",
		metric.WithDescription("Total number of Kafka messages processed"))
	if err != nil {
		log.Printf("Warning: failed to create kafka messages counter: %v", err)
	}

	KafkaErrorsCounter, err = meter.Int64Counter("kafka.errors.total",
		metric.WithDescription("Total number of Kafka errors"))
	if err != nil {
		log.Printf("Warning: failed to create kafka errors counter: %v", err)
	}
}

// collect http metrics
func RecordHTTPMetrics(ctx context.Context, method, path string, statusCode int, duration time.Duration) {
	if HttpRequestCounter != nil {
		HttpRequestCounter.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("http.method", method),
				attribute.String("http.path", path),
				attribute.Int("http.status_code", statusCode),
			),
		)
	}
	if HttpRequestDuration != nil {
		HttpRequestDuration.Record(ctx, duration.Seconds(),
			metric.WithAttributes(
				attribute.String("http.method", method),
				attribute.String("http.path", path),
			),
		)
	}
}

// collect db metrics
func RecordDBMetrics(ctx context.Context, operation, table string, duration time.Duration) {
	if DbOperationsCounter != nil {
		DbOperationsCounter.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("db.operation", operation),
				attribute.String("db.table", table),
			),
		)
	}
	if DbOperationDuration != nil {
		DbOperationDuration.Record(ctx, duration.Seconds(),
			metric.WithAttributes(
				attribute.String("db.operation", operation),
			),
		)
	}
}

// collect cache metrics
func RecordCacheMetrics(ctx context.Context, hit bool) {
	if hit {
		if CacheHitsCounter != nil {
			CacheHitsCounter.Add(ctx, 1)
		}
	} else {
		if CacheMissCounter != nil {
			CacheMissCounter.Add(ctx, 1)
		}
	}
}

// collect kafka metrics
func RecordKafkaMetrics(ctx context.Context, success bool) {
	if KafkaMessageCounter != nil {
		KafkaMessageCounter.Add(ctx, 1)
	}
	if !success && KafkaErrorsCounter != nil {
		KafkaErrorsCounter.Add(ctx, 1)
	}
}
