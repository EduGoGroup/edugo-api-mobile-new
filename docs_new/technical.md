# 💻 Detalles Técnicos y Stack

La API Mobile de EduGo está forjada con herramientas de grado industrial. Hemos diseñado este servicio para ser concurrente, seguro por defecto y observable desde el minuto cero.

---

## 🛠️ El Stack de Infraestructura

* 🐹 **Lenguaje Core:** Go (Golang) v1.26+ — *Elegido por su performance ruteando miles de requests simultáneos mediante goroutines.*
* 🚀 **Framework HTTP:** [Gin Web Framework](https://gin-gonic.com/) — *Ultraligero, con ruteo de prefijos (Radix Tree) de latencia casi nula.*
* 🗄️ **Almacenamiento (Bases de Datos):**
  * **Relacional:** PostgreSQL via [GORM](https://gorm.io/) (Hospedado en Neon).
  * **Documental:** MongoDB via `mongo-driver` oficial (Hospedado en Atlas).
* ⚡ **Caché:** Redis (Implementación in-memory compartida de Edge).
* 🐇 **Mensajería:** RabbitMQ (Publicador asíncrono hacia Workers).
* 🪣 **Persistencia Binaria:** Amazon S3 (para links pre-firmados de `upload/download`).

---

## 🛡️ Línea de Defensa: Middlewares

La cañería HTTP está protegida por una serie de barreras de control implementadas como Middlewares en Gin (usualmente inyectadas desde `edugo-shared` o `internal/infrastructure/http/middleware/`):

1. 🚑 **Recovery:** Envuelve cada petición. Si se produce un `panic()`, evita que el nodo completo muera, lo captura, loguea el trace, y devuelve `500 Internal Server Error`.
2. 🆔 **Request ID:** Inyecta un UUID en cada llamada. Indispensable para rastrear la vida de una petición cruzando múltiples logs.
3. 👮‍♂️ **CORS Seguro:** Solo orígenes autorizados pueden consumir la API.
4. 📊 **Metrics (Prometheus):** Expone un endpoint oculto `/metrics` con histogramas de latencias y contadores de status code para graficar en Grafana.
5. 📝 **Structured Logging:** Utiliza [Zap Logger](https://github.com/uber-go/zap) de Uber para imprimir registros JSON veloces y parseables por herramientas tipo ELK/Datadog.
6. 🛂 **Autenticación (Remote Auth):** Intercepta el Token Bearer (JWT) y valida criptográficamente su integridad y vigencia, a menudo asistido por IAM Platform.
7. 🚦 **Autorización (Permissions):** Verifica que el rol del usuario contenga el derecho exacto (Ej: `PermissionAssessmentsAttempt`).
8. 🕵️ **Auditoría (AuditTrail):** Las peticiones destructivas (`POST`, `PUT`, `DELETE`) dejan un registro inmutable en MongoDB sobre *quién* cambió *qué* y *cuándo*.

---

## 📖 Contratos Claros: Swagger / OpenAPI

Nunca adivinamos endpoints. Todo el código fuente está anotado usando la sintaxis de `swaggo/swag`, lo que compila automáticamente una especificación estricta.

> [!TIP]
> 🔗 **Acceso a la Doc Viva:** Cuando levantas el servidor en local, visita `http://localhost:8065/swagger/index.html`.
> 
> *Para probar endpoints seguros en interfaz, usa el botón "Authorize" e ingresa el JWT con el prefijo `Bearer `.*

---

## 🧪 Ingeniería de Calidad: Testing

No hacemos deploy esperando que "funcione".

* 🔬 **Pruebas Unitarias (Mocking):** Utilizamos generadores de *Mocks* para aislar la Capa de Aplicación. Simulamos respuestas de BD para certificar que la lógica de negocio puramente matemática o de validación funciona.
* 🐳 **Pruebas de Integración (Testcontainers):** 
  En `test/integration/`, destruimos la excusa del *"en mi máquina sí funciona"*. Esta suite levanta contenedores **Efeméros Dockerizados reales** de PostgreSQL, MongoDB y RabbitMQ transitoriamente, inyecta seeds de datos, y ejecuta peticiones HTTP verificando la integración sistémica completa de extremo a extremo.
