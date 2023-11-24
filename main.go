package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "bytes"
	"io"
	"time"

	ginzap "github.com/gin-contrib/zap"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors" // Importa la libreria per il middleware CORS
    _ "github.com/lib/pq"

    "go.opentelemetry.io/otel/propagation"

    "context"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"

	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"

    "github.com/Depado/ginprom"

	// Make sure to import this so the instrumented driver is registered.
	_ "github.com/signalfx/splunk-otel-go/instrumentation/github.com/lib/pq/splunkpq"

)

const (
    host     = "postgres"
    port     = 5432
    user     = "admin"
    password = "admin"
    dbname   = "project"
)

var db *sql.DB

func initTracerAuto() func(context.Context) error {

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint("otel-collector-daemonset-collector.otel-collector.svc.cluster.local:4317"),
		),
	)

	if err != nil {
		log.Fatal("Could not set exporter: ", err)
	}
	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", "backend-go"),
			attribute.String("application", "backend-go"),
		),
	)
	if err != nil {
		log.Fatal("Could not set resources: ", err)
	}

	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter)),
			sdktrace.WithSyncer(exporter),
			sdktrace.WithResource(resources),
		),
	)

    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return exporter.Shutdown
}

func loggerWithTraceInfo(ctx context.Context, logger *zap.Logger) *zap.Logger {
    // Recupera il contesto di tracciamento
    span := trace.SpanFromContext(ctx)
    traceID := span.SpanContext().TraceID().String()
    spanID := span.SpanContext().SpanID().String()

    // Crea un nuovo logger con trace_id e span_id aggiunti come campi
    logger = logger.With(
        zap.String("trace_id", traceID),
        zap.String("span_id", spanID),
    )

    return logger
}

func main() {

    // Crea la stringa di connessione al database PostgreSQL
    connStr := fmt.Sprintf("host=%s port=%d user=%s "+
        "password=%s dbname=%s sslmode=disable",
        host, port, user, password, dbname)

    // Apre la connessione al database
    var err error
    db, err = splunksql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }

    defer db.Close()

    cleanup := initTracerAuto()
	defer cleanup(context.Background())
    // Crea un'istanza di Gin
    r := gin.Default()

    otelginOption := otelgin.WithPropagators(propagation.TraceContext{})
    r.Use(otelgin.Middleware("backend-go", otelginOption))
    
    p := ginprom.New(
		ginprom.Engine(r),
		ginprom.Subsystem("backend"),
        ginprom.Namespace("go"),
		ginprom.Path("/metrics"),
	)
	r.Use(p.Instrument())

    logger, _ := zap.NewProduction()

	r.Use(ginzap.GinzapWithConfig(logger, &ginzap.Config{
		UTC:        true,
		TimeFormat: time.RFC3339,
		Context: ginzap.Fn(func(c *gin.Context) []zapcore.Field {
			fields := []zapcore.Field{}
			// log request ID
			if requestID := c.Writer.Header().Get("X-Request-Id"); requestID != "" {
				fields = append(fields, zap.String("request_id", requestID))
			}

			// log trace and span ID
			if trace.SpanFromContext(c.Request.Context()).SpanContext().IsValid() {
				fields = append(fields, zap.String("trace_id", trace.SpanFromContext(c.Request.Context()).SpanContext().TraceID().String()))
				fields = append(fields, zap.String("span_id", trace.SpanFromContext(c.Request.Context()).SpanContext().SpanID().String()))
			}

			// log request body
			var body []byte
			var buf bytes.Buffer
			tee := io.TeeReader(c.Request.Body, &buf)
			body, _ = io.ReadAll(tee)
			c.Request.Body = io.NopCloser(&buf)
			fields = append(fields, zap.String("body", string(body)))

            return fields
		}),
	}))

    // Aggiungi il middleware CORS
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"*"}, // Consenti richieste da qualsiasi origine
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
    }))

    // Endpoint per la gestione della richiesta di login
    r.POST("/login", func(c *gin.Context) {
        // Esempio di gestione del login

        var requestData struct {
            EmployeeID string `json:"employee_id"`
            Password   string `json:"password"`
        }
        if err := c.ShouldBindJSON(&requestData); err != nil {
            c.AbortWithStatus(http.StatusInternalServerError) 
            return
        }

        employeeID := requestData.EmployeeID
        password := requestData.Password

        // Esegui la query per il login nel database PostgreSQL
        var employeeIDDB string

        // Esegui la query SQL con tracciamento
        ctx := c.Request.Context()
        //_, err = db.ExecContext(ctx, "INSERT INTO table (column1, column2) VALUES ($1, $2)", value1, value2)
        err := db.QueryRowContext(ctx, "SELECT employee_id FROM employee WHERE employee_id = $1 AND password = $2", employeeID, password).Scan(&employeeIDDB)
        if err != nil {
            loggerWithTraceInfo(c.Request.Context(), logger).Error("Error in recovering credentials from DB", zap.Error(err))
            c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"Error in recovering credentials from DB": err.Error()})
            return
        }

        log.Println("Correct Access: " + employeeIDDB)
        c.JSON(http.StatusOK, gin.H{"token": employeeIDDB})
    })

    // Avvia il server
    if err := r.Run(":5001"); err != nil {
        log.Fatal(err)
    }
}
