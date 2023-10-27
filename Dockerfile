# Usa un'immagine di Go come base
FROM golang:latest

WORKDIR /

# Copia il codice sorgente nella directory /
COPY . ./

# Installa le dipendenze Gin e pq
RUN go get -u github.com/gin-gonic/gin
RUN go get -u github.com/lib/pq
RUN go get -u github.com/gin-contrib/cors

# Installa le dipendenze del tuo progetto (se necessario)
#RUN go mod download

# Compila il tuo script Go
RUN go build -o main

# Espone la porta su cui il tuo server ascolterà
EXPOSE 5001

# Avvia l'applicazione Go
CMD ["./main"]
