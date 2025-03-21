# Order Management API

Esta es una API para gestionar órdenes, productos y su stock. La aplicación está desarrollada en Go y utiliza MySQL y Redis como servicios de backend.

---

## **Requisitos previos**

Antes de instalar y ejecutar la aplicación, asegúrate de tener instalados los siguientes componentes:

- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)
- [Go](https://go.dev/) (opcional, solo si deseas ejecutar la aplicación localmente sin Docker)

---

## **Instalación y ejecución**

### **1. Clonar el repositorio**

Clona este repositorio en tu máquina local:

```bash
git clone https://github.com/tu-usuario/order-management.git
cd order-management
```

### **2. Configurar el entorno**

Crea un archivo .env en el directorio raíz con las siguientes variables de entorno, en caso de que el repositorio no los tenga:

```bash
MYSQL_HOST=mysql
MYSQL_PORT=3306
MYSQL_USER=user
MYSQL_PASSWORD=password123
MYSQL_DATABASE=order_management

REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
```

### **3. Construir y ejecutar con Docker**

Levanta los servicios (API, MySQL y Redis) usando Docker Compose:

```bash
docker-compose up --build
```

Esto hará lo siguiente:

- Construirá la imagen de la aplicación Go.
- Levantará los contenedores de MySQL, Redis y la aplicación.

La API estará disponible en <http://localhost:8080>.

---

## **Endpoints disponibles**

### **1. Crear una orden**

- URL: `POST /orders`
- Headers:
-- `Content-Type: application/json`
-- `Idempotency-Key: <unique-key>`
- Body

```json
{
  "customer_name": "John Doe",
  "items": [
    {
      "product_id": 1,
      "quantity": 2
    },
    {
      "product_id": 2,
      "quantity": 3
    }
  ]
}
```

- Respuesta exitosa:

```json
{
  "order_id": 123,
  "customer_name": "John Doe",
  "total_amount": 150.00,
  "items": [
    {
      "product_id": 1,
      "quantity": 2,
      "subtotal": 50.00
    },
    {
      "product_id": 2,
      "quantity": 3,
      "subtotal": 100.00
    }
  ]
}
```

### **2. Obtener una orden por ID**

- URL `GET /orders/{order_id}`
- Headers:
-- `Content-Type: application/json`
- Respuesta exitosa:

```json
{
  "order_id": 123,
  "customer_name": "John Doe",
  "total_amount": 150.00,
  "items": [
    {
      "product_id": 1,
      "quantity": 2,
      "subtotal": 50.00
    },
    {
      "product_id": 2,
      "quantity": 3,
      "subtotal": 100.00
    }
  ]
}
```

### **3. Actualizar el stock de un producto**

- URL: `PUT /products/{product_id}/stock`
- Headers:
-- `Content-Type: application/json`
-- `Idempotency-Key: <unique-key>`
- Body:

```json
{
  "new_stock": 100
}
```

- Respuesta exitosa:

```bash
204 No content
```

### **4. Obtener todos los productos**

- URL: `GET /products`
- Headers:
-- `Content-Type: application/json`
- Respuesta exitosa:

```json
[
  {
    "product_id": 1,
    "name": "Producto A",
    "price": 25.00,
    "stock": 100
  },
  {
    "product_id": 2,
    "name": "Producto B",
    "price": 50.00,
    "stock": 200
  }
]
```

---

## **Notas importantes**

### **1. Idempotencia**

- Usa el header Idempotency-Key para garantizar que las solicitudes repetidas no creen duplicados. Si una solicitud ya fue procesada, la API devolverá la respuesta almacenada.

### **2. Errores comunes**

- `400 Bad Request`: Datos inválidos en el cuerpo de la solicitud.
- `409 Conflict`: La solicitud ya está en progreso o fue completada.
- `500 Internal Server Error`: Error interno del servidor.

### **3. Persistencia de datos**

- Los datos de MySQL y Redis se almacenan en volúmenes de Docker para garantizar la persistencia.

---

## **Comandos útiles**

### **Detener los servicios**

```bash
docker-compose down
```

### **Reconstruir los contenedores**

```bash
docker-compose up --build
```

### **Ver logs de la aplicación**

```bash
docker logs -f go_order_management
```
