# Use the official MySQL Docker image
FROM mysql:8.0.42-debian

WORKDIR /

COPY scripts/tools/csv  tools/csv

# COPY ./sql/* /docker-entrypoint-initdb.d/
# Set environment variables for MySQL configuration

# Build environment 
ENV ENVIRONMENT=dev

CMD ["mysqld"]


