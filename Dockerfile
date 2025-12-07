# Etapa 1: Compilación (builder)
# Usamos una imagen oficial de Go para compilar nuestra aplicación.
FROM golang:1.25-alpine AS builder

# Establecemos el directorio de trabajo
WORKDIR /app

# Copiamos los archivos de dependencias para aprovechar el cache de Docker
COPY go.mod go.sum ./
# Descargamos las dependencias con cache
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copiamos el resto del código fuente
COPY . .

# Compilamos la aplicación con optimizaciones
# CGO_ENABLED=0 deshabilita CGO para crear un binario estático.
# -ldflags="-s -w" reduce el tamaño del binario (strip debug info)
# -trimpath elimina paths del sistema del binario
# Removemos -v para build más rápido (verbose output)
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -o /canvas_service .

# Etapa 2: Ejecución (final)
# Usamos una imagen base mínima de Alpine. Es muy pequeña pero incluye
# herramientas básicas y certificados CA, a diferencia de 'scratch'.
FROM alpine:latest

# Alpine necesita este paquete para ejecutar binarios Go.
RUN apk --no-cache add ca-certificates curl

# Copiamos el ejecutable compilado desde la etapa 'builder'
COPY --from=builder /canvas_service /canvas_service

# Exponemos el puerto en el que corre nuestro servicio
EXPOSE 8080

# Healthcheck so Docker can detect readiness via the service's /health endpoint
HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://127.0.0.1:8080/health || exit 1

# El comando para ejecutar la aplicación cuando el contenedor inicie
ENTRYPOINT ["/canvas_service"]