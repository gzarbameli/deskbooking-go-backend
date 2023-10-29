package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors" // Importa la libreria per il middleware CORS
    _ "github.com/lib/pq"
)

const (
    host     = "postgres"
    port     = 5432
    user     = "admin"
    password = "admin"
    dbname   = "project"
)

var db *sql.DB

func main() {
    // Crea la stringa di connessione al database PostgreSQL
    connStr := fmt.Sprintf("host=%s port=%d user=%s "+
        "password=%s dbname=%s sslmode=disable",
        host, port, user, password, dbname)

    // Apre la connessione al database
    var err error
    db, err = sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }

    defer db.Close()

    // Crea un'istanza di Gin
    r := gin.Default()

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
        err := db.QueryRow("SELECT employee_id FROM employee WHERE employee_id = $1 AND password = $2", employeeID, password).Scan(&employeeIDDB)
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
        gin.SetMode(gin.ReleaseMode)
    }
}
