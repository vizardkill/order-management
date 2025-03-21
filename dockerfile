# Usar una imagen base de Go
FROM golang:1.24.1 as builder

# Establecer el directorio de trabajo dentro del contenedor
WORKDIR /app

# Copiar los archivos de dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar todos los archivos fuente del proyecto
COPY . .

# Cambiar al directorio donde está el archivo main.go
WORKDIR /app/cmd

# Construir el binario de la aplicación
RUN go build -o /app/main .

# Crear una imagen más ligera para ejecutar la aplicación
FROM debian:bookworm-slim

# Establecer el directorio de trabajo dentro del contenedor
WORKDIR /app

# Copiar el binario desde la etapa de construcción
COPY --from=builder /app/main .

# Exponer el puerto en el que corre la aplicación
EXPOSE 8080

# Comando para ejecutar la aplicación
CMD ["./main"]