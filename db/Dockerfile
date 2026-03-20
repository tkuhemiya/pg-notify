FROM postgres:17-alpine

# Keep the image minimal: use official alpine postgres and only copy init SQL.
COPY init/*.sql /docker-entrypoint-initdb.d/
