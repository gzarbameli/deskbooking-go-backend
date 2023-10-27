# Usa l'immagine ufficiale di PostgreSQL come base
FROM postgres:latest

ENV POSTGRES_PASSWORD="admin" 
ENV POSTGRES_USER="admin"
ENV POSTGRES_DB="project"
# Copia gli script SQL nella directory /docker-entrypoint-initdb.d
# Gli script in questa directory verranno eseguiti automaticamente all'avvio del container
COPY creation.sql /docker-entrypoint-initdb.d/

# Porta in cui PostgreSQL ascolterà (imposta su 5432 per impostazione predefinita)
EXPOSE 5432
