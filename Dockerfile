# Etapa 1: Compilación (builder)
# Usamos una imagen oficial de Go para compilar nuestra aplicación.
FROM golang:1.25-alpine AS builder

# Establecemos el directorio de trabajo
WORKDIR /app

# Copiamos los archivos de dependencias para aprovechar el cache de Docker
COPY go.mod go.sum ./
# Descargamos las dependencias
RUN go mod download

# Copiamos el resto del código fuente
COPY . .

# Compilamos la aplicación.
# CGO_ENABLED=0 deshabilita CGO para crear un binario estático.
# -o /canvas_service especifica el nombre del archivo de salida.
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /canvas_service .

# Etapa 2: Ejecución (final)
# Usamos una imagen base mínima de Alpine. Es muy pequeña pero incluye
# herramientas básicas y certificados CA, a diferencia de 'scratch'.
FROM alpine:latest

# Alpine necesita este paquete para ejecutar binarios Go.
RUN apk --no-cache add ca-certificates

# Copiamos el ejecutable compilado desde la etapa 'builder'
COPY --from=builder /canvas_service /canvas_service

# Exponemos el puerto en el que corre nuestro servicio
EXPOSE 8080

# El comando para ejecutar la aplicación cuando el contenedor inicie
ENTRYPOINT ["/canvas_service"]