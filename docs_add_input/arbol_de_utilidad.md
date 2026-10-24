Utilidad
│
├── Rendimiento (Performance)
│   ├── Latencia de Integración Externa
│   │   └── [A1R-E1] Notificar al usuario la disponibilidad de alojamiento externo en ≤ 5 min (registro en ≤ 2 min) [M, M]
│   ├── Tiempo de Respuesta Web
│   │   └── [A1R-E2] Desplegar contenido sugerido en navegador web con latencia ≤ 800 ms [A, B]
│   └── Experiencia de Carga Móvil
│       └── [A1R-E3] Renderizar el primer contenido visual (FCP) en la app móvil en < 1.4 s [M, M]
│
├── Disponibilidad (Availability)
│   └── Tolerancia a Fallas en Infraestructura
│       └── [A2D-E1] Recuperación ante caídas de nodos críticos vía failover manteniendo sesión activa con downtime ≤ 5 min/mes [A, A]
│
├── Escalabilidad (Scalability)
│   └── Capacidad de Carga y Autoescalado
│       └── [A3E-E1] Soportar picos repentinos de tráfico del +500% escalando horizontalmente sin pérdidas y con carga de páginas < 3 s [A, A]
│
├── Interoperabilidad (Interoperability)
│   └── Ingesta y Cruce Automático de Reservas
│       └── [A4I-E1] Escanear confirmaciones por correo, sincronizar con APIs de aerolíneas/hoteles con 100% de precisión sin intervención manual [A, M]
│
├── Seguridad y Privacidad (Security)
│   └── Control de Acceso y Protección de Datos Personales
│       └── [A5S-E1] Bloquear el 100% de accesos no autorizados a itinerarios/correos y registrar intentos en auditoría [A, M]
│
├── Evolucionabilidad y Mantenibilidad (Evolvability)
│   └── Extensibilidad de Nuevos Transportes
│       └── [A6E-E1] Integrar microservicios de autobuses/trenes en < 2 sprints bajo Spec-driven y ADRs sin modificar el core [M, M]
│
└── Usabilidad (Usability) e Internacionalización
    └── Adaptabilidad de Interfaz y Localización
        └── [A7U-E1] Detectar GPS y desplegar mapa de ruta localizado (moneda y hora) en < 2 s y ≤ 3 clics [M, B]
