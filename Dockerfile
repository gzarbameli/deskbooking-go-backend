# Usa un'immagine di Go come base
FROM golang:latest

WORKDIR /

# Copia il codice sorgente nella directory /
COPY . ./

# Installa le dipendenze Gin e pq
RUN go get -u github.com/gin-gonic/gin
RUN go get -u github.com/lib/pq
RUN go get -u github.com/gin-contrib/cors
RUN go get -u go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin
RUN go get -u go.opentelemetry.io/otel
RUN go get -u go.opentelemetry.io/otel/attribute
RUN go get -u go.opentelemetry.io/otel/exporters/otlp/otlptrace
RUN go get -u go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc
RUN go get -u go.opentelemetry.io/otel/sdk/resource
RUN go get -u go.opentelemetry.io/otel/sdk/trace
RUN go get -u github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql
RUN go get -u github.com/signalfx/splunk-otel-go/instrumentation/github.com/lib/pq/splunkpq
# Installa le dipendenze del tuo progetto (se necessario)
#RUN go mod download

# Compila il tuo script Go
RUN go build -o main

# Espone la porta su cui il tuo server ascolterà
EXPOSE 5001

# Avvia l'applicazione Go
CMD ["./main"]
