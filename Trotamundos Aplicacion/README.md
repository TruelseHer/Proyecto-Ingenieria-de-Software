# Trotamundos V2

Plataforma web comercial de reservas multimodales, gestión inteligente de itinerarios, sincronización de confirmaciones por correo/PNR y guía interactiva en destino.

## Características Principales

- **Autenticación Real de Viajero (`/login`, `/register`, `/logout`):**
  - Registro e inicio de sesión con contraseña hasheada criptográficamente en PostgreSQL y sesiones firmadas `HttpOnly`.
  - Las cuentas nuevas inician limpias (`0` viajes y bandeja vacía) listas para explorar el catálogo y armar itinerarios propios.
- **Explorador de Viajes y Catálogo Precargado (`/explore`):**
  - **Paquetes completos con fechas de salida confirmadas:** Japón Imperial (Tokio & Kioto), Alta Velocidad Ibérica (Madrid & Barcelona), Patagonia Andina (Bariloche), Roma & Costa Amalfitana, Cusco & Machu Picchu y París & Londres.
  - **Buscador de tramos individuales:** Vuelos, hoteles, trenes de alta velocidad, autobuses y alquiler de autos con fechas, cupos, tarifas y códigos PNR.
- **Mis Viajes e Itinerario (`/trips` / `/`):**
  - Línea de tiempo cronológica multimodal, agrupación automática por destino/fechas, edición manual protegida contra sobrescrituras automáticas, archivado/restauración, uso compartido invitando acompañantes por email y apertura de casos de soporte con tu agencia preferida.
- **Confirmaciones, Correo y Código PNR (`/inbox`):**
  - Importación directa por código localizador (PNR) de aerolíneas, hoteles y trenes.
  - Bandeja personal de confirmaciones recibidas al reservar en el catálogo y gestión de lista blanca de dominios confiables.
- **Guía Interactiva en Destino (`/companion`):**
  - Mapa interactivo **Leaflet + OpenStreetMap + OSRM** con recorrido real desde el aeropuerto o estación hasta el hotel del viaje.
  - Clima actual en destino (**Open-Meteo API**) y conversión de moneda en vivo (**Frankfurter API**) con caché en Redis y respaldo local automático.
- **Recomendaciones Personalizadas (`/recommendations` y `/trends`):**
  - Motor de sugerencias proactivas de alojamiento, transporte y rutas basado en el análisis del historial y preferencias del viajero (con 3 estrategias: *Afinidad de viajeros*, *Tendencias y puntualidad* y *Optimización multimodal*).
  - Esquema de optimización interactivo (*human-in-the-loop*): ajuste de parámetros en vivo (tipo de alternativa, presupuesto, destino, prioridad) y retroalimentación por tarjeta (*Me interesa*, *Descartar*, *Incorporar a un viaje*) que refina continuamente los pesos del modelo, complementado con el ranking de destinos y puntualidad de operadores.

## Scripts de Inicio, Cierre y Respaldo de Datos

Para iniciar y cerrar la aplicación manteniendo todos los datos creados en PostgreSQL:

```bash
cd TrotamundosAplicacionV2
./iniciar.sh              # Inicia db, cache y app (restaura respaldo si la BD está vacía)
./iniciar.sh --build      # Reconstruye la imagen de la aplicación e inicia
./iniciar.sh --restaurar  # Fuerza la restauración desde backups/trotamundos_backup.sql
./cerrar.sh               # Exporta la BD a backups/trotamundos_backup.sql y apaga los contenedores
```

Al ejecutar `./cerrar.sh`, se genera un respaldo SQL portátil en `backups/trotamundos_backup.sql` y se conservan los volúmenes persistentes de Docker.

## Ejecución con Docker Compose

```bash
cd TrotamundosAplicacionV2
docker compose up --build -d
```

Luego abre **[http://localhost:8080](http://localhost:8080)** en tu navegador, crea tu cuenta en `/register` y comienza a explorar y reservar tus viajes.

## Ejecución Local y Pruebas Automatizadas

```bash
cd TrotamundosAplicacionV2
go test -v ./...
go run ./cmd/server
```
