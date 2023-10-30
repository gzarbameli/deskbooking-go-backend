package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors" // Importa la libreria per il middleware CORS
    _ "github.com/lib/pq"

    "go.opentelemetry.io/otel/propagation"

    "context"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"

	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"

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

    //// Strumenta il database SQL
    //db = sql.OpenDB("postgres", db)
    //db = sqltrace.NewDB(db, sqltrace.WithSpanName(dbtrace.QueryName("Query")))

    cleanup := initTracerAuto()
	defer cleanup(context.Background())
    // Crea un'istanza di Gin
    r := gin.Default()

    otelginOption := otelgin.WithPropagators(propagation.TraceContext{})
    r.Use(otelgin.Middleware("backend-go", otelginOption))
    
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
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
            c.JSON(http.StatusUnauthorized, gin.H{"message": "Incorrect Username and/or Password"})
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
