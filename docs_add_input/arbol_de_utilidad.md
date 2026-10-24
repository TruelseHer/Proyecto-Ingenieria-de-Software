# Árbol de Utilidad — Sistema "Trotamundos"

Documento de consolidación y priorización arquitectónica siguiendo el estándar SEI (Kazman & Cervantes).

---

## 1. Estructura Jerárquica del Árbol de Utilidad

```text
Utilidad (Trotamundos)
|
+-- Rendimiento (Performance)
|   +-- Latencia de Integración Externa
|   |   \-- [A1R-E1] (M, M) Notificar al usuario la disponibilidad de alojamiento externo en <= 5 min (registro en <= 2 min).
|   +-- Tiempo de Respuesta Web
|   |   \-- [A1R-E2] (A, B) Desplegar contenido sugerido en navegador web con latencia <= 800 ms.
|   \-- Experiencia de Carga Móvil
|       \-- [A1R-E3] (M, M) Renderizar el primer contenido visual (FCP) en la aplicación móvil en < 1.4 s.
|
+-- Disponibilidad (Availability)
|   \-- Tolerancia a Fallas en Infraestructura
|       \-- [A2D-E1] (A, A) Failover automático ante caída de nodo crítico manteniendo sesión activa; inactividad no planificada <= 5 min/mes.
|
+-- Escalabilidad (Scalability)
|   \-- Capacidad de Carga y Autoescalado
|       \-- [A3E-E1] (A, A) Soportar pico repentino de tráfico de +500% escalando horizontalmente con carga de páginas < 3 s y sin pérdida de datos.
|
+-- Interoperabilidad (Interoperability)
|   \-- Ingesta y Cruce Automático de Reservas
|       \-- [A4I-E1] (A, M) Escanear confirmaciones por correo, sincronizar con APIs de aerolíneas/hoteles con 100% de precisión sin intervención manual.
|
+-- Seguridad y Privacidad (Security)
|   \-- Control de Acceso y Protección de Datos Personales
|       \-- [A5S-E1] (A, M) Bloquear el 100% de accesos no autorizados a itinerarios/correos y registrar los incidentes en auditoría.
|
+-- Evolucionabilidad y Mantenibilidad (Evolvability)
|   \-- Extensibilidad de Nuevos Transportes
|       \-- [A6E-E1] (M, M) Integrar microservicios de autobuses/trenes en < 2 sprints bajo Spec-driven y ADRs sin modificar el core.
|
\-- Usabilidad (Usability) e Internacionalización
    \-- Adaptabilidad de Interfaz y Localización
        \-- [A7U-E1] (M, B) Detectar GPS nativo, adaptar mapa/moneda/horario y brindar ruta al hotel en < 2 s y <= 3 clics.
```

---

## 2. Tabla Detallada de Hojas y Priorización

| Atributo de Calidad | Sub-refinamiento | ID Escenario | Hoja del Árbol (Estímulo + Respuesta) | Importancia | Dificultad | Prioridad Final |
| --- | --- | --- | --- | --- | --- | --- |
| **Rendimiento** | Latencia de Integración Externa | `A1R-E1` | Ante publicación de alojamiento en servicio externo, procesar y notificar al usuario en $\le$ 5 min (registro en $\le$ 2 min). | Media* | Media* | **(M, M)** |
| **Rendimiento** | Tiempo de Respuesta Web | `A1R-E2` | Ante navegación web del usuario, entregar contenido sugerido con latencia $\le$ 800 ms. | Alta | Baja* | **(A, B)** |
| **Rendimiento** | Experiencia de Carga Móvil | `A1R-E3` | Ante apertura o navegación en la app móvil, renderizar el primer contenido visual (FCP) en $<$ 1.4 s. | Media* | Media* | **(M, M)** |
| **Disponibilidad** | Tolerancia a Fallas en Infraestructura | `A2D-E1` | Ante caída de infraestructura o base de datos, enruta automáticamente el tráfico hacia un servidor de respaldo (failover) preservando la sesión con indisponibilidad $\le$ 5 min/mes. | Alta | Alta* | **(A, A)** |
| **Escalabilidad** | Capacidad de Carga y Autoescalado | `A3E-E1` | Ante evento disruptivo con pico del +500% de tráfico, La plataforma escala horizontalmente aprovisionando de manera automática nuevas instancias de servidores y equilibrando la carga para procesar las peticiones simultáneas sin degradar el servicio, cargando páginas en $<$ 3 s sin fallos. | Alta | Alta* | **(A, A)** |
| **Interoperabilidad** | Ingesta y Cruce de Reservas | `A4I-E1` | Ante recepción de confirmaciones por email, autenticar, extraer y cruzar datos vía APIs con 100% de precisión automática. | Alta | Media* | **(A, M)** |
| **Seguridad** | Control de Acceso y Privacidad | `A5S-E1` | Ante intentos de acceso no autorizado a correos/itinerarios, denegar el 100%, enmascarar datos y auditar. | Alta | Media* | **(A, M)** |
| **Evolucionabilidad** | Extensibilidad de Nuevos Transportes | `A6E-E1` | Ante solicitud de incorporar trenes/buses, construir módulos aislados (Spec-driven y ADRs) en $<$ 2 sprints sin tocar módulos core. | Media* | Media* | **(M, M)** |
| **Usabilidad** | Adaptabilidad y Localización | `A7U-E1` | Ante uso en el extranjero, leer GPS y presentar indicaciones localizadas (mapa, moneda, hora) en $<$ 2 s y $\le$ 3 clics. | Media* | Baja* | **(M, B)** |

*\* Clasificado como supuesto justificado al no venir prefijado por el enunciado original.*

---

## 3. Justificación de Prioridades (Importancia, Dificultad)

- **[A1R-E2] Tiempo de Respuesta Web (Alta, Baja):**
  - *Importancia (Alta):* Impacta de forma inmediata en la retención y conversión del usuario en la web principal.
  - *Dificultad (Baja):* Se aborda mediante tácticas consolidadas de caching distribuido (CDN, Redis) y optimización de assets.
- **[A2D-E1] Tolerancia a Fallas (Alta, Alta):**
  - *Importancia (Alta):* Una tolerancia máxima de 5 min/mes equivale a un 99.99% de disponibilidad (SLA crítico para viajeros en tránsito).
  - *Dificultad (Alta):* Requiere infraestructura multizona, replicación continua de base de datos y preservación de estado de sesión durante el failover.
- **[A3E-E1] Capacidad ante Picos (Alta, Alta):**
  - *Importancia (Alta):* El sistema debe operar con 15 millones de cuentas y 2 millones de usuarios activos semanales, respondiendo ante crisis masivas.
  - *Dificultad (Alta):* Absorber picos súbitos del +500% demanda desacoplamiento asíncrono con colas elásticas, sharding y políticas de autoescalado horizontal rápido.
- **[A4I-E1] Ingesta de Reservas (Alta, Media):**
  - *Importancia (Alta):* Representa el valor diferencial de la plataforma (centralizador automático de viajes sin carga manual).
  - *Dificultad (Media):* Requiere adaptadores modulares para tolerar variaciones sintácticas y semánticas de múltiples proveedores y formatos de correo.
- **[A5S-E1] Control de Acceso (Alta, Media):**
  - *Importancia (Alta):* Manejo de datos personales (PII) e itinerarios privados con exigencias normativas estrictas.
  - *Dificultad (Media):* Implementación de esquemas de autorización por atributos/roles (ABAC/RBAC) y encriptación de datos en reposo y tránsito.
- **[A1R-E1] Latencia de Integración Externa (Media, Media):**
  - *Importancia (Media):* Proceso asíncrono en segundo plano; unos minutos de holgura no degradan la experiencia interactiva del usuario.
  - *Dificultad (Media):* Requiere colas de mensajería y control de tasa (rate limiting) para interactuar con servicios de terceros.
- **[A1R-E3] Carga Móvil (Media, Media):**
  - *Importancia (Media):* Crítica para la experiencia en apps, pero mitigable con renderizado progresivo y estados de carga (skeletons).
  - *Dificultad (Media):* Implica optimización del bundle móvil, carga diferida y soporte de condiciones de red variables.
- **[A6E-E1] Extensibilidad Transportes (Media, Media):**
  - *Importancia (Media):* Driver a mediano plazo para la evolución del negocio que no condiciona el MVP inicial.
  - *Dificultad (Media):* Se mitiga con diseño desacoplado guiado por contratos (Spec-driven) y microservicios independientes.
- **[A7U-E1] Adaptabilidad y Localización (Media, Baja):**
  - *Importancia (Media):* Mejora la satisfacción en destino del viajero internacional.
  - *Dificultad (Baja):* Tecnologías móviles estándar resuelven nativamente la geolocalización y bibliotecas de formato resuelven i18n/moneda.

---

## 4. Hojas de Mayor Prioridad (Drivers Primarios para ADD)

1. **Resolución Inmediata (Quick Win):**
  - `[A1R-E2]` — Rendimiento Web `(Alta, Baja)`: Alto impacto directo al usuario con bajo costo de implementación táctica.
2. **Drivers Críticos de Arquitectura (Mayor Riesgo Estructural):**
  - `[A2D-E1]` — Disponibilidad `(Alta, Alta)` y `[A3E-E1]` — Escalabilidad `(Alta, Alta)`: Decisiones que determinan el diseño de infraestructura, topología de despliegue y patrones de resiliencia desde la Iteración 1 de ADD.
3. **Drivers Esenciales de Negocio y Dominio:**
  - `[A4I-E1]` — Interoperabilidad `(Alta, Media)` y `[A5S-E1]` — Seguridad `(Alta, Media)`: Garantizan la propuesta funcional central y la protección de información sensible.
