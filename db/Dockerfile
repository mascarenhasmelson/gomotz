FROM postgres:16-alpine
COPY schema.sql /docker-entrypoint-initdb.d/schema.sql
ENV POSTGRES_USER=admin \
    POSTGRES_PASSWORD=StrongPassword123 \
    POSTGRES_DB=tunnel_services
EXPOSE 5432
