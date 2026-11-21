# AGENTS.md — Guía Arquitectónica y de Producto (`TrotamundosAplicacionV2`)

## 1. Visión del Producto
`TrotamundosAplicacionV2` es una plataforma web comercial de gestión de viajes multimodales, exploración de catálogos de reservas y asistencia al viajero en destino.
La aplicación debe comportarse y verse como un **producto real en producción para el usuario final**, implementando internamente las decisiones arquitectónicas del repositorio (`READme.md`, `inputs/`, `iteraciones/`, `adr/`, `spec/`) sin exponer jamás jerga académica ni controles de simulación.

## 2. Identidad Visual y Diseño (Estilo Airbnb / Booking — Marca Roja)
- **Estética:** Luminosa, limpia, editorial y moderna inspirada en Airbnb y Booking, donde nuestro color corporativo principal es el **Rojo Trotamundos (`#E11D48` / `#FF385C`, hover `#BE123C`)** sobre fondos claros (`#FFFFFF`, `#F8FAFC`, `#FFF1F2` sutil para destacados) y tipografía de alta legibilidad (`Plus Jakarta Sans` + `Inter` + `JetBrains Mono` solo para códigos PNR/vuelos).
- **Jerarquía visual:** Tarjetas de paquetes y destinos con fotografías reales (Unsplash curadas), badges claros de transporte, precios transparentes y estados de reserva legibles.

## 3. Reglas Estrictas de Experiencia de Usuario (Prohibiciones)
1. **Cero notación del ADD en la UI o código:** Prohibido mostrar códigos (`US-01`, `QA-1`, `ADR-05`, `CRN-1`), nombres de tácticas (`Circuit Breaker`, `DLQ`, `Heartbeat`, `OLAP`) o textos académicos.
2. **Cero selectores de cambio de usuario (`SwitchUser`):** La identidad del usuario se gestiona exclusivamente mediante registro (`/register`), inicio de sesión con email y contraseña (`/login`) y cierre de sesión (`/logout`). Ningún usuario puede ver listados de otros usuarios ni viajes ajenos.
3. **Cero simuladores de fallos en la interfaz:** Prohibido incluir botones de "Simular caída de proveedor", "Forzar error 403", "Simular pico de tráfico" o "Enviar spam". Los mecanismos de resiliencia (timeouts, reintentos, circuit breaker, validación de permisos y auditoría) operan de forma transparente en el backend.
4. **Experiencia de Cuenta Nueva (Idea B):**
   - Cuando un usuario se registra desde cero, su sección **"Mis Viajes"** y su **"Bandeja de Confirmaciones (`/inbox`)"** inician limpias (0 viajes y 0 correos previos).
   - Cada vez que el usuario reserva un paquete o tramo desde el **Explorador de Viajes (`/explore`)** o importa una reserva externa por código PNR, el sistema emite automáticamente el correo de confirmación en su bandeja de entrada (`/inbox`) además de permitir sincronizarlo con su itinerario.
5. **Catálogo precargado listo para usar:** El usuario dispone desde el primer momento de un catálogo completo en PostgreSQL de **Paquetes de Viaje con fechas disponibles** y **Ofertas de Tramos Individuales** (vuelos, hoteles, trenes, autobuses y alquiler de autos) con fechas, horarios, cupos y precios para elegir y reservar, además de poder registrar reservas manuales externas.

## 4. Stack Tecnológico y Estructura de Capas
- **Directorio raíz:** `TrotamundosAplicacionV2/`
- **Backend:** Go 1.22+ (`net/http` estándar, `html/template`, `golang.org/x/crypto/bcrypt`, `github.com/jackc/pgx/v5`).
- **Frontend Híbrido (HATEOAS):** HTMX 2.0 + Tailwind CSS + Leaflet.js (OpenStreetMap) para mapas interactivos.
- **Persistencia y Caché:** PostgreSQL 16 (con migraciones automáticas y seed del catálogo de viajes/proveedores) + Redis 7.
- **APIs Públicas Gratuitas (sin API Key, con caché y fallback local):**
  - **Rutas y Mapas:** OpenStreetMap + OSRM (`router.project-osrm.org`) para cálculo real de rutas al hotel.
  - **Clima en Destino:** Open-Meteo (`api.open-meteo.com`) por coordenadas del destino.
  - **Tasas de Cambio:** Frankfurter API (`api.frankfurter.app`) con caché en Redis y respaldo local.
- **Estructura de paquetes Go:**
  - `cmd/server/` — Punto de entrada, configuración e inyección de dependencias.
  - `internal/config/` — Variables de entorno y configuración.
  - `internal/domain/` — Entidades de negocio ricas (`User`, `Trip`, `Reservation`, `CatalogPackage`, `CatalogOffer`, `EmailConfirmation`, `TravelAgency`, `SenderRule`, `SupportTicket`).
  - `internal/repository/` — Persistencia PostgreSQL + memoria para tests, migraciones automáticas y datos semilla del catálogo.
  - `internal/platform/` — Autenticación (`bcrypt` + sesiones firmadas JWT/cookie `HttpOnly`), caché Redis, bus de eventos interno, cliente HTTP resiliente y adaptadores de APIs públicas (Clima, Divisas, Rutas OSRM).
  - `internal/service/` — Casos de uso (`AuthService`, `CatalogService`, `TripService`, `SyncService`, `CompanionService`, `InsightsService`, `RecommendationService`).
  - `internal/handler/` — Middlewares HTTP (Auth, Rate Limit, Auditoría), controladores Web/HTMX y endpoints JSON.
  - `web/templates/` y `web/static/` — Layouts, páginas completas, parciales HTMX y hojas de estilo.
