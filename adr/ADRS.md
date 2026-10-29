# Arquitectural Decision Records

Bajo el formato Markdown Arquitectural Desicion Records

## **ADR-01: Selección del estilo macro arquitectónico basado en Microservicios Agrupados por Rol/Carga**

### **Contexto y Declaración del Problema**

El sistema "Trotamundos" es un desarrollo *Greenfield* concebido para soportar una escala de 15 millones de cuentas y 2 millones de usuarios activos semanales. La plataforma experimenta picos abruptos de concurrencia (+500%) ante eventos climáticos o disrupciones de viaje masivas, coexistiendo tres perfiles operativos disímiles: operaciones interactivas de itinerarios en tiempo real, sondeo de correos en segundo plano y procesamiento analítico de datos. Se debe definir la estructura organizativa macro inicial del sistema que soporte el escalamiento independiente sin comprometer los tiempos de respuesta.

* **Drivers arquitectónicos:** CRN-1 (Estructura general inicial), A3E-E1 / QA-3 (Escalabilidad y picos de \+500%), A6E-E1 (Evolucionabilidad).

## **Opciones Consideradas**

1. **Microservicios Agrupados por Carga / Rol Operativo**  
2. **Monolito Único Centralizado**  
3. **Microservicios Granulares Extremos (*Database-per-Service* por entidad)**

## **Resultado de la Decisión**

Opción elegida: **Microservicios Agrupados por Carga / Rol Operativo**. Se estructura el sistema en subsistemas macro autónomos según su comportamiento operacional (Servicio Transaccional de Viajes, Subsistema de Ingesta Externa y Subsistema Analítico).

### **Pros y Contras de las Opciones**

#### **Microservicios Agrupados por Carga / Rol Operativo**

* **Bueno**, porque permite el escalamiento horizontal elástico independiente de los subsistemas bajo mayor estrés durante picos masivos (+500%), sin necesidad de replicar todo el sistema.  
* **Bueno**, porque acota la complejidad operativa y técnica para la implementación del prototipo del TPE al evitar transacciones distribuidas excesivas.  
* **Malo**, porque demanda la gestión de despliegue de múltiples artefactos y el mantenimiento de contratos de red entre subsistemas.

#### **Monolito Único Centralizado**

* **Bueno**, porque simplifica el despliegue inicial y el modelo transaccional ACID local.  
* **Malo**, porque ante un pico de tráfico obliga a escalar la aplicación completa de forma ineficiente consumiendo recursos innecesarios.  
* **Malo**, porque las tareas pesadas en segundo plano y consultas de analítica compiten directamente por recursos de CPU/memoria con las peticiones interactivas de los viajeros.

#### **Microservicios Granulares Extremos**

* **Bueno**, porque ofrece máxima autonomía de despliegue a nivel de entidad de dominio.  
* **Malo**, porque introduce una complejidad de red, orquestación y transacciones distribuidas (Sagas) incompatible con el alcance (*scope*) y tiempos del prototipo exigidos por la cátedra.

## **ADR-02: Adopción del patrón Edge API Gateway con Backend for Frontend (BFF) diferenciado**

### **Contexto y Declaración del Problema**

Trotamundos requiere acceso simultáneo a través de interfaces Web de escritorio y aplicaciones Móviles nativas. Cada plataforma posee requerimientos de rendimiento dispares: First-Contentful Paint $<1.4s$ en conexiones celulares móviles y tiempos de respuesta web $\leq 800ms$. Comunicar a los clientes directamente con múltiples servicios internos provocaría degradación por alta latencia de red (*chattiness*), mientras que unificar la API acoplaría los formatos de presentación.

* **Drivers arquitectónicos:** CON-1 (Soporte Web y Móvil nativo), CRN-1 (Fronteras del sistema), QA-1 / A1R-E2 / A1R-E3 (Latencias objetivo).

### **Opciones Consideradas**

1. **Edge API Gateway combinado con BFFs dedicados (Mobile BFF y Web BFF)**  
2. **Comunicación Directa Cliente-a-Servicio (*Direct Client-to-Service*)**  
3. **API Gateway Único Agnóstico sin capa BFF**

### **Resultado de la Decisión**

Opción elegida: **Edge API Gateway combinado con BFFs dedicados (Mobile BFF y Web BFF)**. El Edge Gateway centraliza la terminación TLS, autenticación base por token y ruteo de perímetro; mientras que cada BFF adapta, filtra y agrega payloads según las capacidades de su cliente (Web o Móvil).

### **Pros y Contras de las Opciones**

#### **Edge API Gateway con BFFs dedicados**

* **Bueno**, porque optimiza los viajes de red: el Mobile BFF compacta y despoja campos del JSON para cumplir el FCP $<1.4s$ en móviles, mientras el Web BFF entrega estructuras ricas para pantalla ancha ($\leq 800ms$).  
* **Bueno**, porque centraliza las políticas de seguridad perimetral y autenticación en un único punto de entrada público.  
* **Bueno**, porque desacopla la evolución de las interfaces gráficas respecto a las APIs del dominio transaccional interno.  
* **Malo**, porque introduce una capa adicional en el flujo de peticiones que debe implementarse y mantenerse sincronizada.

#### **Comunicación Directa Cliente-a-Servicio**

* **Bueno**, porque prescinde de componentes perimetrales intermedios.  
* **Malo**, porque expone los contratos internos e IP de los servicios directamente a Internet, comprometiendo la seguridad perimetral.  
* **Malo**, porque fuerza a las aplicaciones móviles a ejecutar múltiples llamadas sucesivas en redes inestables, violando las restricciones de latencia.

#### **API Gateway Único Agnóstico**

* **Bueno**, porque simplifica la infraestructura al mantener un único componente de entrada.  
* **Malo**, porque unifica contratos obligando al cliente móvil a descargar información redundante pensada para la web, incrementando el consumo de ancho de banda y latencia.

## **ADR-03: Desacoplamiento de la Ingesta Externa mediante Conector Message Broker (Asynchronous Messaging)**

### **Contexto y Declaración del Problema**

La plataforma debe escanear bandejas de correo para detectar reservas (US-01) e integrarse con APIs de agencias externas (hoteles, aerolíneas, rentadoras) (US-03, CON-2). Dichos servicios externos tienen latencias variables, fallos temporales e interfaces heterogéneas. Un encadenamiento de llamadas sincrónicas degradaría la experiencia del usuario o causaría caídas en cascada ante la indisponibilidad de proveedores externos.

* **Drivers arquitectónicos:** CON-2 (Integración con terceros), QA-3 / A3E-E1 (Escalabilidad y absorción de picos), QA-2 (Disponibilidad global y tolerancia a fallos).

### **Opciones Consideradas**

1. **Desacoplamiento Asíncrono mediante Message Broker (Colas / Event Bus)**  
2. **Orquestación Sincrónica Request-Response bloqueante (REST/HTTP)**

### **Resultado de la Decisión**

Opción elegida: **Desacoplamiento Asíncrono mediante Message Broker (Colas / Event Bus)**. El Subsistema de Ingesta procesa externamente correos y APIs externas emitiendo eventos normalizados al broker, de donde el servicio de viajes consume las reservas sin acoplamiento temporal.

### **Pros y Contras de las Opciones**

#### **Desacoplamiento Asíncrono mediante Message Broker**

* **Bueno**, porque amortigua picos repentinos de tráfico actuando como un *buffer* elástico de peticiones.  
* **Bueno**, porque aísla caídas externas: si la API de una aerolínea o proveedor falla, el mensaje se retiene y reintenta automáticamente en segundo plano sin propagar la falla ni bloquear la navegación del usuario.  
* **Bueno**, porque desacopla el ciclo de vida del proceso de ingesta continua respecto al servicio transaccional.  
* **Malo**, porque introduce un modelo de consistencia eventual (las reservas confirmadas impactan unos segundos después en el itinerario).

#### **Orquestación Sincrónica Request-Response (REST/HTTP)**

* **Bueno**, porque garantiza consistencia inmediata en las transacciones mediante llamadas directas.  
* **Malo**, porque el hilo de ejecución queda bloqueado esperando la respuesta de servidores de correo o agencias externas, consumiendo conexiones del servidor.  
* **Malo**, porque genera fallas en cascada y atenta directamente contra la disponibilidad requerida si un tercero experimenta indisponibilidad.

### **ADR-04: Segregación física de Almacenes de Datos Operativo (OLTP) y Analítico (OLAP)**

### **Contexto y Declaración del Problema**

Trotamundos debe ofrecer un panel de gestión en tiempo real de reservas, garantizando respuestas web rápidas ($\leq 800ms$). Paralelamente, se debe alimentar un equipo de monetización que ejecutará consultas y reportes analíticos pesados sobre tendencias, destinos y cancelaciones históricas (US-07, CRN-3). Si ambos perfiles operan sobre el mismo almacén, las consultas pesadas bloquearán las operaciones cotidianas de los usuarios.

* **Drivers arquitectónicos:** CRN-3 (Segregación de procesamiento analítico y transaccional), CRN-1 (Estructura general del sistema), QA-1 (Rendimiento y latencia web).

### **Opciones Consideradas**

1. **Segregación Dual: Base de Datos Operativa (OLTP) vs. Repositorio Analítico (OLAP)**  
2. **Base de Datos Relacional Única Tradicional (Traditional Relational Architecture)**

### **Resultado de la Decisión**

* Opción elegida: **Segregación Dual: Base de Datos Operativa (OLTP) vs. Repositorio Analítico (OLAP)**. Se destina una base transaccional optimizada para lecturas y escrituras atómicas de viajes y usuarios, y un Data Warehouse / almacén analítico dedicado para las consultas agregadas de monetización.

### **Pros y Contras de las Opciones**

#### **Segregación Dual (OLTP vs. OLAP)**

* **Bueno**, porque elimina la contención de recursos y el bloqueo de tablas o bloqueos transaccionales críticos durante la ejecución de consultas analíticas pesadas.  
* **Bueno**, porque permite optimizar los esquemas de manera independiente: esquema normalizado/indexado para transacciones y esquema desnormalizado columnar/dimensional para analítica.  
* **Bueno**, porque asegura el cumplimiento de la meta de latencia web $\leq 800ms$ al aislar la carga transaccional.  
* **Malo**, porque requiere implementar un mecanismo de propagación de datos (ETL por lotes o streaming/eventos) para mantener actualizado el almacén analítico.

#### **Base de Datos Relacional Única Tradicional**

* **Bueno**, porque mantiene una única fuente de verdad transaccional sin necesidad de pipelines de replicación.  
* **Malo**, porque las consultas de agregación y reportes de monetización saturarán el uso de CPU e I/O del motor de base de datos, violando las latencias de respuesta web.  
* **Malo**, porque es una alternativa descalificada en la literatura de referencia para arquitecturas que combinan alta concurrencia transaccional con analítica de datos a gran escala.

### **ADR-05: Adopción del Patrón Modelo de Dominio Enriquecido (*Rich Domain Model*)**

### Contexto y Declaración del Problema

Trotamundos requiere consolidar reservas dispares (vuelos, alojamientos, autos) en una entidad unificada de viaje (Trip), gestionando reglas de negocio complejas sobre su ciclo de vida: creación manual, agregación temporal, transiciones de estado (*Activo* $\rightarrow$ *Completado* $\rightarrow$ *Archivado*) y asociación con agencias de viaje preferidas (US-04, US-05, US-06). Dispersar estas reglas en procedimientos o controladores generaría duplicación de código e inconsistencias a medida que se agreguen nuevos tipos de transporte.

* **Drivers arquitectónicos:** US-04, US-05, US-06.

### Opciones Consideradas

1. **Modelo de Dominio Enriquecido (*Rich Domain Model Pattern*)**  
2. **Modelo de Dominio Anémico (*Anemic Domain Model*) con *Transaction Scripts***

### Resultado de la Decisión

Opción elegida: **Modelo de Dominio Enriquecido (*Rich Domain Model Pattern*)**. Las entidades Trip y Reservation encapsulan tanto los datos como los métodos que gobiernan sus invariantes y reglas de negocio del ciclo de vida.

### **Pros y Contras de las Opciones**

#### **Modelo de Dominio Enriquecido**

* **Bueno**, porque centraliza las reglas de negocio y transiciones de estado en la entidad raíz (Trip), evitando estados inválidos en los itinerarios.  
* **Bueno**, porque facilita la extensibilidad polimórfica para incorporar nuevos transportes (trenes, autobuses) mediante subclases de Reservation sin alterar la lógica de agregación.  
* **Bueno**, porque permite pruebas unitarias aisladas y exhaustivas de las reglas de negocio sin requerir dependencias de base de datos o frameworks.  
* **Malo**, porque requiere un mapeo objeto-relacional más sofisticado para persistir estructuras jerárquicas y polimórficas en la base de datos.

#### **Modelo de Dominio Anémico con *Transaction Scripts***

* **Bueno**, porque es trivial de mapear inicialmente en estructuras de persistencia directas.  
* **Malo**, porque dispersa la lógica del ciclo de vida y validaciones en los controladores o servicios procedurales, generando código duplicado y una arquitectura difícil de mantener (*"no hay atajos"*).

### **ADR-06: Estructuración Interna del Servicio de Viajes mediante *Service Application Pattern* (Arquitectura en Capas)**

### Contexto y Declaración del Problema

El subsistema transaccional trotamundos.core\_travel debe exponer APIs REST para los BFFs y consumir eventos asíncronos del broker, abstrayéndose por completo de la interfaz de usuario. Es necesario organizar su código interno en capas con responsabilidades desacopladas para delimitar unidades de trabajo claras asignables a los integrantes del equipo durante el desarrollo del prototipo.

* **Drivers arquitectónicos:** CRN-2, US-04, US-05.

### Opciones Consideradas

1. **Arquitectura de Referencia *Service Application* en 3 Capas (Controladores $\rightarrow$ Servicios $\rightarrow$ Repositorios)**  
2. **Arquitectura de 2 Capas (*Smart Controllers* acoplados directamente al acceso a datos)**

### Resultado de la Decisión

Opción elegida: **Arquitectura de Referencia *Service Application* en 3 Capas**. El microservicio se descompone en: Capa de Presentación/API (Controllers y Event Consumers), Capa de Lógica de Negocio (TripService y entidades de dominio) y Capa de Persistencia (Repositories).

### **Pros y Contras de las Opciones**

#### **Arquitectura de Referencia *Service Application* en 3 Capas**

* **Bueno**, porque separa limpiamente la exposición de contratos HTTP/JSON de la lógica de dominio y del motor de persistencia.  
* **Bueno**, porque define unidades de implementación estandarizadas que facilitan la estimación, planificación de sprints y división del trabajo del equipo.  
* **Bueno**, porque permite sustituir o actualizar el motor de base de datos sin propagar cambios hacia la capa de controladores.  
* **Malo**, porque introduce clases intermedias (DTOs, Mappers) para transferir información a través de las fronteras de las capas.

#### **Arquitectura de 2 Capas (*Smart Controllers*)**

* **Bueno**, porque reduce la cantidad inicial de archivos de código en servicios simples.  
* **Malo**, porque mezcla la validación de protocolos web con consultas SQL/persistencia, imposibilitando la reutilización de lógica ante eventos asíncronos y dificultando las pruebas automatizadas.

### **ADR-07: Aislamiento de Fuentes Externas mediante Capa Anticorrupción (*Anti-Corruption Layer \- ACL*)**

### Contexto y Declaración del Problema

El sistema debe interoperar con proveedores turísticos heterogéneos (APIs de aerolíneas, cadenas hoteleras, rentadoras de autos) y extraer reservas desde correos electrónicos con formatos no normalizados (US-01, US-03). Integrar directamente los modelos externos provocaría que cambios arbitrarios en APIs de terceros o en plantillas de correo rompan el modelo transaccional central de Trotamundos.

* **Drivers arquitectónicos:** CON-2, US-01, US-03.

### Opciones Consideradas

1. **Capa Anticorrupción (*Anti-Corruption Layer \- ACL*) con Adaptadores y Traductor Canónico**  
2. **Modelo Compartido Directo (*Shared Kernel* / Consumo crudo de APIs de terceros)**

### Resultado de la Decisión

Opción elegida: **Capa Anticorrupción (*Anti-Corruption Layer \- ACL*) con Adaptadores y Traductor Canónico**. Dentro de trotamundos.ingestion, adaptadores dedicados (AirlineAdapter, HotelAdapter, etc.) consumen los servicios externos y delegan en un traductor (CanonicalDataTranslator) la conversión hacia el evento canónico ReservationDetectedEvent antes de publicarlo al broker.

### **Pros y Contras de las Opciones**

#### **Capa Anticorrupción (*ACL*)**

* **Bueno**, porque protege el dominio central de Trotamundos frente a discrepancias de esquemas, cambios de versión o terminología propietaria de proveedores externos.  
* **Bueno**, porque encapsula la complejidad del parsing de correos y llamadas HTTP externas en un subsistema periférico dedicado.  
* **Bueno**, porque simplifica la adición futura de nuevos proveedores turísticos creando un adaptador específico sin tocar el código transaccional de viajes.  
* **Malo**, porque requiere mantener módulos adicionales de traducción y mapeo para cada integración externa soportada.

#### **Modelo Compartido Directo (*Shared Kernel*)**

* **Bueno**, porque evita la sobrecarga computacional de transformar objetos entre capas.  
* **Malo**, porque acopla fuertemente el esquema de base de datos interno a las APIs de terceros, propagando fallos y cambios externos hacia todo el sistema.

### **ADR-08: Adopción del Enfoque *Contract-First* mediante OpenAPI y AsyncAPI (*Spec-Driven Development*)**

### Contexto y Declaración del Problema

La cátedra exige enfáticamente el enfoque de *Spec-Driven Development* para el diseño y construcción del prototipo (CRN-2). Asimismo, la interacción entre múltiples subsistemas (BFFs hacia Core Travel y Ingestion hacia Core Travel) requiere eliminar ambigüedades en las cargas útiles (*payloads*), rutas y códigos de estado previo a la codificación.

* **Drivers arquitectónicos:** CRN-2, CON-1, CON-2.

### Opciones Consideradas

1. **Diseño Dirigido por Especificaciones (*Spec-Driven / Contract-First*) con OpenAPI 3.0 y AsyncAPI**  
2. **Desarrollo Centrado en Código (*Code-First* con generación posterior de especificaciones)**

### Resultado de la Decisión

Opción elegida: **Diseño Dirigido por Especificaciones (*Spec-Driven / Contract-First*) con OpenAPI 3.0 y AsyncAPI**. Se formalizan y versionan los contratos TravelCoreAPI.yaml (endpoints REST síncronos) y ReservationDetectedEvent (JSON Schema / AsyncAPI) antes de la implementación de los servicios.

### **Pros y Contras de las Opciones**

#### ***Spec-Driven / Contract-First***

* **Bueno**, porque cumple explícitamente con la restricción metodológica CRN-2 de la cátedra.  
* **Bueno**, porque permite el trabajo en paralelo de los desarrolladores: los equipos de frontend/BFF pueden trabajar contra *mocks* generados desde el contrato mientras se programa el backend.  
* **Bueno**, porque asegura alta trazabilidad entre la arquitectura documentada y el código del prototipo.  
* **Malo**, porque exige disciplina inicial para definir y versionar esquemas YAML/JSON detallados antes de tirar las primeras líneas de código.

#### ***Code-First***

* **Bueno**, porque agiliza el inicio de la programación en equipos pequeños con un único desarrollador.  
* **Malo**, porque viola las consignas de evaluación de la cátedra respecto a *Spec-Driven Development*.  
* **Malo**, porque genera contratos inestables que cambian de forma imprevista ante refactorizaciones internas del código de backend.

### **ADR-09: Selección de Base de Datos Relacional (PostgreSQL) para el Modelo Transaccional de Viajes**

### Contexto y Declaración del Problema

El subsistema trotamundos.core\_travel requiere almacenar usuarios, viajes, itinerarios y múltiples reservas relacionadas de forma manual o automática (US-04, US-05). Se debe garantizar atomicidad y consistencia estricta (ACID) en operaciones de actualización y cancelación de tramos de viaje para evitar itinerarios corruptos, manteniendo tiempos de respuesta rápidos en las consultas cotidianas.

* **Drivers arquitectónicos:** US-04, US-05, CRN-3.

### Opciones Consideradas

1. **Motor de Base de Datos Relacional SQL (PostgreSQL)**  
2. **Base de Datos NoSQL Orientada a Documentos (MongoDB)**

### Resultado de la Decisión

Opción elegida: **Motor de Base de Datos Relacional SQL (PostgreSQL)**. Se adopta como almacén operativo (*OLTP Store*) debido a su soporte maduro de integridad referencial, transaccionalidad ACID y capacidades de indexación sobre claves foráneas y campos temporales.

### **Pros y Contras de las Opciones**

#### **Base de Datos Relacional (PostgreSQL)**

* **Bueno**, porque garantiza transacciones ACID nativas para operaciones críticas como la cancelación o actualización coordinada de reservas dentro de un viaje.  
* **Bueno**, porque previene anomalías de datos mediante restricciones de clave foránea e integridad referencial estricta.  
* **Bueno**, porque se integra de manera directa y estándar con herramientas de persistencia y mapeo objeto-relacional (JPA / Spring Data) para implementar el patrón Repository.  
* **Malo**, porque requiere mantenimiento de esquemas de migración estructurados (*DDL scripts*) ante cambios en el modelo de datos.

#### **Base de Datos NoSQL Orientada a Documentos (MongoDB)**

* **Bueno**, porque maneja naturalmente la estructura jerárquica de un viaje con reservas anidadas en un solo documento JSON.  
* **Malo**, porque la gestión de transacciones atómicas distribuidas sobre múltiples colecciones (usuarios, itinerarios compartidos y logs de auditoría) resulta más compleja y menos eficiente en alta concurrencia transaccional.

### **ADR-10: Topología de Base de Datos Primario-Réplica con Failover Multi-AZ y Réplicas de Lectura**

* **Estado:** Aceptado  
* **Fecha:** 2026-10-28  
* **Decisores:** Equipo de Arquitectura (Trotamundos)

### Contexto y Declaración del Problema

Trotamundos experimenta picos de tráfico de hasta \+500% ante cancelaciones o disrupciones de viaje (QA-3), requiriendo al mismo tiempo un SLA de disponibilidad con indisponibilidad no planificada $\leq 5min/mes$ (QA-2). Las transacciones de viaje, pagos y cancelaciones (US-04, US-05) no toleran inconsistencias temporales ni lecturas sucias (incompatibles con consistencia eventual), pero concentrar todas las consultas masivas sobre una única base de datos agotaría el cómputo del servidor.

* **Drivers arquitectónicos:** QA-2, QA-3, QA-1, US-04, US-05.

## **Opciones Consideradas**

1. **Topología Primario-Réplica con Nodo Standby Sincrónico Multi-AZ y Pool de Réplicas de Lectura (*Read Replicas Pool*)**  
2. **Base de Datos Distribuida Multimaster con Consistencia Eventual (*Multi-Master Eventual Consistency*)**  
3. **Escalado Vertical Monolítico de Base de Datos en un Único Servidor (*Single-Instance Scale-Up*)**

## **Resultado de la Decisión**

Opción elegida: **Topología Primario-Réplica con Nodo Standby Sincrónico Multi-AZ y Pool de Réplicas de Lectura**. Se implementa segregación a nivel de conexión mediante RoutingDataSource: todas las transacciones de escritura (mutaciones de estado de viaje y reservas) se dirigen exclusivamente al nodo primario garantizando **Consistencia Transaccional Estricta (ACID)**, mientras que las consultas masivas del itinerario se balancean hacia el pool de réplicas de lectura. La disponibilidad se asegura mediante un nodo *Hot Standby* replicado sincrónicamente en una segunda zona de disponibilidad con promoción automática (*Failover* en $<60s$).

### **Pros y Contras de las Opciones**

#### **Topología Primario-Réplica Multi-AZ con Read Replicas**

* **Bueno**, porque resuelve el compromiso de integridad: mantiene propiedades ACID puras en modificaciones críticas sin admitir sobreventas ni datos corruptos.  
* **Bueno**, porque absorbe el incremento del \+500% en consultas de itinerarios añadiendo réplicas de lectura de forma horizontal (QA-3).  
* **Bueno**, porque cumple con el SLA de indisponibilidad $\leq 5min/mes$ ante desastres físicos de datacenter mediante el failover automático hacia el nodo Standby (QA-2).  
* **Malo**, porque existe un desfase mínimo de milisegundos (*replication lag*) en las réplicas de lectura, requiriendo que operaciones críticas de lectura inmediata tras escritura se fuercen temporalmente contra el nodo primario.

#### **Base de Datos Distribuida Multimaster (Consistencia Eventual)**

* **Bueno**, porque permite escrituras concurrentes distribuidas con alta tolerancia a particiones de red.  
* **Malo**, porque viola el trade-off de consistencia definido para Trotamundos; la resolución de conflictos por timestamp puede generar sobreescrituras accidentales en reservas y compras concurrentes.

#### **Escalado Vertical en Único Servidor**

* **Bueno**, porque elimina la complejidad de sincronización de réplicas.  
* **Malo**, porque introduce un punto único de falla (SPOF) catastrófico que destruye el SLA de disponibilidad (QA-2) y tiene un límite físico insostenible ante la escala proyectada.

### **ADR-11: Priorización de Seguridad y Validación Criptográfica en Memoria en el API Gateway**

* **Estado:** Aceptado  
* **Fecha:** 2026-10-28  
* **Decisores:** Equipo de Arquitectura (Trotamundos)

### **Contexto y Declaración del Problema**

El sistema debe procesar datos privados de viajes y permitir compartir itinerarios de forma colaborativa (QA-5). El requerimiento exige denegar de forma verificable el 100% de los accesos no autorizados. Evaluar firmas digitales, validar listas de acceso (ACLs) y consultar políticas en cada solicitud entrante consume ciclos de CPU e introduce micro-latencias, entrando en conflicto directo con el límite de respuesta de $\leq 800ms$ (QA-1).

* **Drivers arquitectónicos:** QA-5, QA-1, CRN-2.

### **Opciones Consideradas**

1. **Inspección Criptográfica en Memoria en el Edge Gateway \+ Motor de Autorización Estricto a Nivel de Servicio (*AccessControlEngine*)**  
2. **Bypass Perimetral de Autorización utilizando Caché Pública (*Edge Caching sin Re-autenticación*)**  
3. **Delegación de Validación Completa en un Servicio Remoto de Identidad (OAuth Token Introspection síncrono por cada petición)**

## **Resultado de la Decisión**

Opción elegida: **Inspección Criptográfica en Memoria en el Edge Gateway \+ Motor de Autorización Estricto a Nivel de Servicio**. Se consagra formalmente el trade-off: **Seguridad y Privacidad por sobre Rendimiento puro**. No se permite entregar itinerarios protegidos desde cachés públicas no autenticadas. El *Edge Gateway* descarga y cachea las claves públicas del emisor de identidad, validando firmas JWT localmente en memoria (tiempo $<1ms$). A nivel interno, el AccessControlInterceptor evalúa las políticas de pertenencia al itinerario y despacha auditorías estructuradas ante cualquier intento de intrusión (SecurityAuditLogger).

### **Pros y Contras de las Opciones**

#### **Inspección en Memoria \+ Control de Acceso Estricto**

* **Bueno**, porque garantiza el 100% de rechazo a accesos ilegítimos o con identificadores forjados en URLs (QA-5).  
* **Bueno**, porque mitiga el impacto de latencia sobre QA-1 al evitar llamadas remotas síncronas al servidor de autenticación para validar firmas.  
* **Bueno**, porque genera un registro de auditoría inmutable de accesos denegados para detección de incidentes (Audit Trail).  
* **Malo**, porque impone una sobrecarga constante de verificación criptográfica de CPU en el API Gateway durante ráfagas de tráfico.

#### **Bypass Perimetral con Caché Pública**

* **Bueno**, porque minimiza la latencia de respuesta entregando contenido directamente desde la red perimetral (QA-1).  
* **Malo**, porque presenta una vulnerabilidad crítica donde un usuario no autorizado o con sesión revocada podría acceder a itinerarios cacheados ajenos, violando frontalmente QA-5.

#### **Token Introspection Remoto por Petición**

* **Bueno**, porque garantiza control en tiempo real del estado de revocación del token.  
* **Malo**, porque añade entre $50ms$ y $150ms$ adicionales de latencia por cada llamada HTTP, degradando severamente QA-1 bajo concurrencia masiva.

### ADR-12: Autoescalado Horizontal Elástico (HPA) y Amortiguación de Ingesta mediante Message Broker

* **Estado:** Aceptado  
* **Fecha:** 2026-10-28  
* **Decisores:** Equipo de Arquitectura (Trotamundos)

## **Contexto y Declaración del Problema**

Durante contingencias climáticas o disrupciones de huelgas aeroportuarias, las conexiones a Trotamundos se multiplican por 5 (+500%), tanto por consultas concurrentes al panel como por la llegada masiva de correos y avisos de aerolíneas (QA-3, US-01, US-03). Una arquitectura con dimensionamiento estático se saturaría rápidamente por falta de hilos de cómputo o caídas de workers por sobrecarga de memoria.

* **Drivers arquitectónicos:** QA-3, QA-2, US-01, US-03.

## **Opciones Consideradas**

1. **Autoescalado Horizontal de Pods (Kubernetes HPA) con Amortiguación Asíncrona en Broker Particionado (*Queue Requests*)**  
2. **Dimensionamiento Estático Sobredimensionado (*Over-Provisioning* Permanente)**  
3. **Control de Flujo Basado Exclusivamente en Descarte Temprano (*Load Shedding agresivo / 503 Service Unavailable*)**

## **Resultado de la Decisión**

Opción elegida: **Autoescalado Horizontal de Pods (Kubernetes HPA) con Amortiguación Asíncrona en Broker Particionado**. Se aplican dos tácticas complementarias del SEI: *Increase Resources* para la capa de cómputo de Core Travel y BFFs (escalando dinámicamente de 3 a 15 réplicas cuando el uso de CPU supera el 70% o la tasa de requests supera las 250 req/s por pod), y *Queue Requests* en el clúster de mensajería para retener de forma persistente los eventos de ingesta masiva sin saturar a los servicios de dominio.

### **Pros y Contras de las Opciones**

#### **HPA \+ Amortiguación en Message Broker**

* **Bueno**, porque absorbe el pico del \+500% dinámicamente sin degradación catastrófica de las latencias de usuario (QA-3).  
* **Bueno**, porque desacopla el ritmo de procesamiento de confirmaciones externas; si Core Travel se satura temporalmente, los eventos se encolan en el broker sin pérdida de información (QA-2).  
* **Bueno**, porque optimiza los costos operativos de infraestructura en la nube reduciendo instancias en periodos valle (*scale-down*).  
* **Malo**, porque el aprovisionamiento de nuevos pods de backend introduce una pequeña latencia inicial de inicialización (*cold start*), requiriendo mantener una base mínima de 3 réplicas calientes en todo momento.

#### **Over-Provisioning Permanente**

* **Bueno**, porque no sufre demoras de arranque de nuevos contenedores.  
* **Malo**, porque genera costos financieros de nube inasumibles e injustificados durante los días de tráfico regular.

#### **Load Shedding Agresivo**

* **Bueno**, porque protege la estabilidad física de los servidores ante saturación.  
* **Malo**, porque rechaza peticiones legítimas de viajeros en situaciones de emergencia y contingencia, destruyendo la confianza en el producto y violando las medidas de respuesta de QA-3 y QA-2.

### **ADR-13: Ingesta Analítica Asíncrona Guiada por Eventos para Consultas de Monetización**

* **Estado:** Aceptado  
* **Fecha:** 2026-10-28  
* **Decisores:** Equipo de Arquitectura (Trotamundos)

#### **Contexto y Declaración del Problema**

El sistema debe recopilar y procesar métricas históricas de viaje (destinos preferidos, cancelaciones masivas, demanda por temporada) para alimentar al equipo de monetización mediante consultas analíticas pesadas (US-07). Estas consultas de negocio consumen cómputo intensivo de agregación e indexación. Se debe definir el mecanismo de ingesta y alimentación del almacén analítico segregado (ADR-04) sin comprometer las transacciones operativas ni degradar la latencia web ($\leq 800$ ms) de los 2 millones de usuarios activos semanales.

* **Drivers arquitectónicos:** US-07, CRN-3, QA-1, QA-3.

#### **Opciones Consideradas**

1. **Pipeline Asíncrono Basado en Eventos (Event-Driven Data Pipeline con Kafka y ClickHouse/Timescale)**  
2. **Extracción Periódica por Lotes sobre la Base Operativa (Direct OLTP Polling / Batch Dump)**  
3. **Replicación Lógica Transaccional Directa (CDC — Change Data Capture a nivel de WAL)**

#### **Resultado de la Decisión**

Opción elegida: **Pipeline Asíncrono Basado en Eventos (Event-Driven Data Pipeline con Kafka y ClickHouse/Timescale)**. El subsistema transaccional trotamundos.core\_travel emite eventos de ciclo de vida (TripLifecycleEvents) hacia el broker de mensajería en cada transición de estado relevante. El microservicio trotamundos.analytics consume de manera autónoma y asíncrona estos eventos, aplicándoles transformaciones y persistiendo las filas en un almacén columnar/analítico optimizado.

##### **Pros y Contras de las Opciones**

###### **Pipeline Asíncrono Basado en Eventos**

* **Bueno**, porque elimina al 100% el acoplamiento y contención de recursos sobre la base de datos relacional operativa de PostgreSQL, blindando el rendimiento web (QA-1).  
* **Bueno**, porque absorbe de forma elástica picos masivos de cambios de estado sin sobrecargar al equipo analítico.  
* **Bueno**, porque permite transformar, enriquecer y anonimizar los eventos en vuelo antes de la persistencia histórica.  
* **Malo**, porque las métricas de monetización presentan consistencia eventual respecto a las últimas operaciones del usuario.

###### **Extracción Periódica por Lotes (Batch Dump)**

* **Bueno**, porque simplifica el desarrollo inicial al requerir solo scripts de exportación programados.  
* **Malo**, porque bloquea tablas operativas transaccionales durante la ejecución de los dumps masivos nocturnos.  
* **Malo**, porque genera ventanas de desactualización de hasta 24 horas en las métricas de negocio.

###### **Replicación Lógica Directa (CDC)**

* **Bueno**, porque captura de forma instantánea cambios de base de datos sin alterar el código de aplicación.  
* **Malo**, porque acopla el esquema de base de datos analítico a la estructura interna de tablas relacionales de viajes.  
* **Malo**, porque dificulta la aplicación de reglas complejas de anonimización y agregación previas al almacenamiento.

### **ADR-14: Anonimización y Desidentificación de Datos en Tránsito para el Repositorio Analítico**

* **Estado:** Aceptado  
* **Fecha:** 2026-10-28  
* **Decisores:** Equipo de Arquitectura (Trotamundos)

#### **Contexto y Declaración del Problema**

Trotamundos procesa nombres, documentos, itinerarios detallados y correos de millones de usuarios (QA-5). El requerimiento US-07 exige de forma explícita que todas las métricas de tendencias y destinos almacenadas en el Data Warehouse estén completamente anonimizadas y desprovistas de información personal identificable (PII), mitigando riesgos de filtración masiva ante accesos analíticos de terceros.

* **Drivers arquitectónicos:** US-07, QA-5, CRN-3.

#### **Opciones Consideradas**

1. **Sanitización y Hashing Salteado en Tránsito (In-Flight Data Masking & De-identification)**  
2. **Anonimización Dinámica en Capa de Visualización (Query-Time Masking en el Data Warehouse)**  
3. **Almacenamiento de Identificadores Reales con Control de Acceso Estricto (Role-Based Access Control sin Anonimización Física)**

#### **Resultado de la Decisión**

Opción elegida: **Sanitización y Hashing Salteado en Tránsito (In-Flight Data Masking & De-identification)**. Dentro del componente DataAnonymizerFilter de trotamundos.analytics, se eliminan campos sensibles (nombres, PNRs de tickets y correos) y se reemplaza el identificador de usuario (userId) por un hash criptográfico salteado no invertible (pseudonymId) antes de escribir en el repositorio analítico OLAP.

##### **Pros y Contras de las Opciones**

###### **Sanitización y Hashing en Tránsito**

* **Bueno**, porque garantiza cumplimiento normativo y privacidad por diseño (*Privacy by Design*): el Data Warehouse no contiene físicamente datos personales (QA-5, US-07).  
* **Bueno**, porque preserva la capacidad de agrupar viajes por usuario seudonimizado para análisis de comportamiento sin revelar la identidad real.  
* **Malo**, porque una vez ingresado el dato anonimizado es técnicamente imposible desanonimizarlo o reconstruir el itinerario original si se requiriera soporte particular sobre un reporte.

###### **Anonimización Dinámica en Capa de Visualización**

* **Bueno**, porque conserva el registro transaccional original intacto para consultas administrativas.  
* **Malo**, porque el almacén analítico persiste datos personales en texto claro; un fallo de configuración en las vistas expondría información confidencial.

###### **RBAC Exclusivo sin Anonimización**

* **Bueno**, porque no impone procesamiento criptográfico ni mapeo de identificadores.  
* **Malo**, porque viola explícitamente el criterio de aceptación de US-07 que exige anonimización física de los datos recopilados.

### **ADR-15: Desacoplamiento del Motor de Recomendaciones mediante Arquitectura de Plugins y Puntos de Extensión**

* **Estado:** Aceptado  
* **Fecha:** 2026-10-28  
* **Decisores:** Equipo de Arquitectura (Trotamundos)

#### **Contexto y Declaración del Problema**

Trotamundos tiene como propósito de diseño fundacional ser una plataforma enfocada en recomendaciones inteligentes para viajeros (CRN-5). Es necesario acoplar el motor de sugerencias personalizadas (alojamientos alternativos, destinos de temporada, actividades locales) de forma tal que no invada el dominio transaccional de viajes ni genere bloqueos de ejecución durante las operaciones rutinarias del usuario.

* **Drivers arquitectónicos:** CRN-5, QA-6, QA-1.

#### **Opciones Consideradas**

1. **Módulo Desacoplado con Extensión por Plugins y Consumo de Vistas Agregadas (Microkernel / Strategy Pattern con Caché)**  
2. **Invocación Sincrónica Directa Core-a-Recomendador (Request-Reply RPC en cada consulta de viaje)**  
3. **Módulo Empotrado Monolítico dentro de Core Travel (In-Process Recommendation Engine)**

#### **Resultado de la Decisión**

Opción elegida: **Módulo Desacoplado con Extensión por Plugins y Consumo de Vistas Agregadas (Microkernel / Strategy Pattern con Caché)**. El subsistema trotamundos.recommendations opera como un servicio autónomo que procesa periódicamente modelos analíticos provenientes del Data Warehouse y precalcula recomendaciones en una base clave-valor de baja latencia (Redis). La interfaz RecommendationStrategy actúa como punto de extensión abierto para incorporar nuevos algoritmos predictivos o modelos colaborativos sin tocar el core.

##### **Pros y Contras de las Opciones**

###### **Módulo Desacoplado con Extensión por Plugins**

* **Bueno**, porque cumple el propósito de diseño CRN-5 aislando por completo la complejidad matemática o de machine learning fuera de trotamundos.core\_travel.  
* **Bueno**, porque permite cambiar y probar estrategias algorítmicas de recomendación de forma pluggable e independiente (QA-6).  
* **Bueno**, porque protege el tiempo de respuesta web $\leq 800$ ms sirviendo recomendaciones cacheadas en $<50$ ms sin latencias de cálculo en vivo (QA-1).  
* **Malo**, porque las recomendaciones sugeridas pueden presentar un breve desfase de actualización respecto a las decisiones de viaje tomadas en los últimos minutos.

###### **Invocación Sincrónica Core-a-Recomendador**

* **Bueno**, porque genera sugerencias con contexto 100% fresco en cada petición.  
* **Malo**, porque enlaza la disponibilidad del recomendador a la del itinerario y duplica las latencias de red, degradando severamente QA-1.

###### **Módulo Empotrado Monolítico**

* **Bueno**, porque reduce la sobrecarga de despliegue y red.  
* **Malo**, porque el consumo intensivo de memoria y CPU de los algoritmos de recomendación penalizaría directamente las operaciones transaccionales críticas de reservas.

### **ADR-16: Resiliencia de Ingesta Externa mediante Circuit Breaker y Dead-Letter Queue (DLQ)**

* **Estado:** Aceptado  
* **Fecha:** 2026-10-28  
* **Decisores:** Equipo de Arquitectura (Trotamundos)

#### **Contexto y Declaración del Problema**

La ingesta e integración con proveedores externos (aerolíneas, hoteles, sistemas de correo) enfrenta APIs de terceros con variabilidad de latencia, ventanas de mantenimiento y fallos intermitentes (CON-2, US-01, US-03, QA-4). Los reintentos síncronos continuos colapsan los hilos de ejecución de los workers de ingesta. Se debe asegurar tolerancia a fallos, no propagación de caídas y 100% de precisión sin descarte inadvertido de confirmaciones.

* **Drivers arquitectónicos:** QA-4, CON-2, US-03, QA-2.

#### **Opciones Consideradas**

1. **Patrón Circuit Breaker combinado con Dead-Letter Queue (DLQ) y Retroceso Exponencial**  
2. **Reintentos Síncronos Indefinidos en Bucle Bloqueante (Tight Retry Loop)**  
3. **Descarte Inmediato con Notificación de Error al Usuario (Fail-Fast sin Buffer)**

#### **Resultado de la Decisión**

Opción elegida: **Patrón Circuit Breaker combinado con Dead-Letter Queue (DLQ) y Retroceso Exponencial**. En trotamundos.ingestion, cada invocación a proveedores externos es supervisada por una instancia de Circuit Breaker. Ante una tasa sustained de errores del 50% o timeouts repetidos, el circuito se abre evitando llamadas improductivas; las peticiones pendientes o fallidas se redirigen automáticamente a una cola de mensajes de reintento diferido (DLQ en Kafka) con retroceso exponencial (*exponential backoff* con jitter).

##### **Pros y Contras de las Opciones**

###### **Circuit Breaker con DLQ y Backoff**

* **Bueno**, porque preserva la integridad y completitud de las reservas: ningún mensaje válido de confirmación se pierde por indisponibilidad de terceros (QA-4, CON-2).  
* **Bueno**, porque evita la saturación de conexiones y consumo de CPU en workers durante caídas masivas de agencias o aerolíneas.  
* **Bueno**, porque protege al proveedor externo de tormentas de peticiones permitiendo su recuperación gradual (*Half-Open state*).  
* **Malo**, porque demanda monitoreo y alarmas operativas dedicadas para la cola DLQ ante escenarios de fallo prolongado.

###### **Reintentos Síncronos Indefinidos**

* **Bueno**, porque es trivial de codificar en librerías cliente.  
* **Malo**, porque satura el pool de conexiones HTTP, bloquea hilos del sistema operativo y provoca caídas en cascada del subsistema de ingesta.

###### **Descarte Inmediato (Fail-Fast)**

* **Bueno**, porque libera inmediatamente los recursos de cómputo.  
* **Malo**, porque viola los criterios de QA-4 y US-01 al descartar confirmaciones legítimas recibidas por los viajeros.

### **ADR-17: Extensibilidad de Transportes mediante Modelo Abierto/Cerrado y Contratos Modulares**

* **Estado:** Aceptado  
* **Fecha:** 2026-10-28  
* **Decisores:** Equipo de Arquitectura (Trotamundos)

#### **Contexto y Declaración del Problema**

El sistema debe estar diseñado para integrar futuros medios de transporte (autobuses de larga distancia, redes de trenes) en menos de 2 sprints de desarrollo sin modificar el código de los servicios existentes ni generar regresiones en el cálculo de itinerarios consolidados (QA-6, CRN-4).

* **Drivers arquitectónicos:** QA-6, CRN-4, US-05.

#### **Opciones Consideradas**

1. **Jerarquía Polimórfica Abierta/Cerrada (Domain Subtyping con Contratos Canónicos Modulares)**  
2. **Entidad Genérica Universal sin Tipificación Estricta (Universal JSON Property Bag)**  
3. **Microservicio Independiente Dedicado por Cada Medio de Transporte (Database-per-Transport Microservices)**  
     
* Opción elegida: **Jerarquía Polimórfica Abierta/Cerrada (Domain Subtyping con Contratos Canónicos Modulares)**. La clase abstracta Reservation dentro de trotamundos.core\_travel.domain define el contrato polimórfico de operaciones temporales y espaciales comunes. Nuevos medios de transporte (como TrainReservation o BusReservation) se incorporan como extensiones concretas sin modificar la lógica interna del agregador Trip. Paralelamente, los esquemas de eventos en trotamundos.common\_messaging emplean extensiones polimórficas del tipo enum reservationType, facilitando el desarrollo paralelo de los equipos de integración (CRN-4).

* Opción elegida: **Jerarquía Polimórfica Abierta/Cerrada (Domain Subtyping con Contratos Canónicos Modulares)**. La clase abstracta Reservation dentro de trotamundos.core\_travel.domain define el contrato polimórfico de operaciones temporales y espaciales comunes. Nuevos medios de transporte (como TrainReservation o BusReservation) se incorporan como extensiones concretas sin modificar la lógica interna del agregador Trip. Paralelamente, los esquemas de eventos en trotamundos.common\_messaging emplean extensiones polimórficas del tipo enum reservationType, facilitando el desarrollo paralelo de los equipos de integración (CRN-4).

### **Pros y Contras de las Opciones**

### **Jerarquía Polimórfica Abierta/Cerrada**

* **Bueno**, porque cumple la métrica de evolucionabilidad QA-6: la incorporación de trenes y autobuses se completa en  sprints creando una subclase de entidad y su respectivo adaptador, pasando el 100% de las pruebas automatizadas de regresión.  
* **Bueno**, porque aísla responsabilidades de código permitiendo a desarrolladores trabajar en paralelo sin conflictos de merge en el core (CRN-4).  
* **Bueno**, porque conserva la rigurosidad tipada del *Rich Domain Model* consolidado en ADR-05.  
* **Malo**, porque demanda esquemas de persistencia polimórfica (por ejemplo, estrategia *Joined* o *Single Table with Discriminator* en JPA/PostgreSQL).

### **Entidad Genérica Universal (Property Bag)**

* **Bueno**, porque permite persistir cualquier medio de transporte sin alterar esquemas ni clases de código.  
* **Malo**, porque traslada las validaciones de negocio a complejas sentencias condicionales desestructuradas, aumentando la tasa de errores de runtime y destruyendo la mantenibilidad.

### **Microservicio Dedicado por Cada Medio de Transporte**

* **Bueno**, porque ofrece máxima autonomía de despliegue a cada vertical de transporte.  
* **Malo**, porque atomiza innecesariamente el dominio, obligando a coordinar transacciones distribuidas complejas (patrón Saga) para consolidar un único itinerario de viaje.

### **ADR-18: Adopción del Patrón MVVM y Abstracción de Hardware de Localización para la Aplicación Móvil**

* **Estado:** Aceptado  
* **Fecha:** 2026-10-28  
* **Decisores:** Equipo de Arquitectura (Trotamundos)

#### **Contexto y Declaración del Problema**

El requerimiento **QA-7** exige que un viajero en tránsito internacional obtenga indicaciones de localización y rutas interactivas hacia su alojamiento en menos de 2 segundos y con un esfuerzo de interacción mínimo ($\leq 3$ clics). La restricción **CON-1** establece la obligatoriedad de dar soporte nativo a clientes móviles. Integrar directamente APIs de mapas y sensores propietarios del sistema operativo dentro de la interfaz gráfica acoplaría el código al ciclo de vida del teléfono, provocaría bloqueos del hilo principal de dibujo (*UI Thread freeze*) y consumiría batería excesiva al consultar el satélite GPS continuamente.

* **Drivers arquitectónicos:** QA-7, CON-1, QA-1 \[A1R-E3\].

#### **Opciones Consideradas**

1. **Patrón MVVM (Model-View-ViewModel) con Abstracción de Hardware (ILocationService) y Posicionamiento Rápido bajo Demanda**  
2. **Arquitectura Móvil Monolítica orientada a Actividades (Fat Activity / UI Controller acoplado a LocationManager nativo)**  
3. **Contenedor WebView Híbrido con Consumo de la Geolocation API del Navegador**

#### **Resultado de la Decisión**

Opción elegida: **Patrón MVVM con Abstracción de Hardware (ILocationService) y Posicionamiento Rápido bajo Demanda**. La aplicación móvil organiza su capa de presentación mediante ViewModels que manejan el estado a través de flujos reactivos (StateFlow), asegurando cero bloqueos en la interfaz. La interacción con los sensores GPS del dispositivo se encapsula tras la interfaz ILocationService, que implementa adaptadores específicos para Android (FusedLocationProviderClient) y iOS (CLLocationManager). Para cumplir la meta de $<2$ segundos, se emplea la técnica de *Last Known Location* como coordenada inicial inmediata mientras se afina la precisión satelital de fondo.

##### **Pros y Contras de las Opciones**

###### **MVVM con Abstracción de Hardware**

* **Bueno**, porque cumple el límite de tiempo de QA-7: el ViewModel dispara en paralelo la lectura de la última posición conocida del sensor y la consulta al BFF, renderizando la pantalla en $<2$ segundos.  
* **Bueno**, porque minimiza el esfuerzo cognitivo del usuario mediante navegación reactiva en $\leq 3$ clics sin pasos de configuración manuales.  
* **Bueno**, porque aísla el código de dominio de la plataforma móvil, facilitando pruebas unitarias sobre la lógica de rutas sin requerir emuladores ni dispositivos físicos.  
* **Bueno**, porque protege la batería del viajero evitando el sondeo GPS ininterrumpido.  
* **Malo**, porque demanda escribir y mantener implementaciones puente de la interfaz ILocationService para cada sistema operativo soportado.

###### **Fat Activity / UI Controller Acoplado**

* **Bueno**, porque reduce la cantidad de capas de código iniciales en prototipos simples.  
* **Malo**, porque acopla los callbacks asíncronos del GPS a la destrucción de pantallas (rotaciones de pantalla, llamadas entrantes), generando fugas de memoria (*memory leaks*) y excepciones de tiempo de ejecución.  
* **Malo**, porque mezcla responsabilidades de renderizado con cálculos de hardware, degradando la mantenibilidad.

###### **Contenedor WebView Híbrido**

* **Bueno**, porque permite reutilizar código web de la SPA de escritorio.  
* **Malo**, porque viola frontalmente la restricción CON-1 de cliente móvil nativo.  
* **Malo**, porque la sobrecarga de inicialización del motor web de renderizado supera los 1.4 segundos solo en arranque, haciendo imposible satisfacer QA-7 y QA-1.

### **ADR-19: Localización e Internacionalización (i18n) Híbrida en Memoria con Agregación Contextual en Mobile BFF**

* **Estado:** Aceptado  
* **Fecha:** 2026-10-28  
* **Decisores:** Equipo de Arquitectura (Trotamundos)

#### **Contexto y Declaración del Problema**

Al operar en escenarios internacionales (multirregión), el sistema debe mostrar instantáneamente monedas extranjeras convertidas, husos horarios correctos y mapas locales (QA-7). En redes celulares extranjeras (*roaming*), la latencia de red es elevada y los datos móviles son costosos y fluctuantes. Si la aplicación realizara llamadas HTTP separadas para traducir textos, otra para consultar tasas de cambio de divisas y otra para pedir la ubicación del hotel, el número de peticiones consecutivas (*chattiness*) sobrepasaría el umbral tolerable de 2 segundos.

* **Drivers arquitectónicos:** QA-7, CON-1, QA-1 \[A1R-E3\].

#### **Opciones Consideradas**

1. **Internacionalización Híbrida (Formateo en Memoria en Cliente \+ Endpoint Consolidado en Mobile BFF)**  
2. **Formateo y Traducción Dinámica del Lado del Servidor (Server-Side Screen Rendering / Localized HTML Strings)**  
3. **Múltiples Consultas REST Granulares Directas a Microservicios Internos (Chatty Client Architecture)**

#### **Resultado de la Decisión**

Opción elegida: **Internacionalización Híbrida (Formateo en Memoria en Cliente \+ Endpoint Consolidado en Mobile BFF)**. Se adopta una estrategia desacoplada de internacionalización:

1. El Mobile BFF agrega en una única llamada REST (GET /mobile/v1/trips/{id}/navigation-context) las coordenadas del alojamiento, el identificador del huso horario del destino y la tasa oficial de cambio provista por el adaptador de divisas cacheado en Redis.  
2. El componente ClientI18nEngine de la app móvil procesa localmente en memoria y en tiempo real ($<1$ ms) las conversiones numéricas de moneda y los desfases de calendario utilizando diccionarios estáticos compilados en el binario.  
3. El componente de mapas utiliza caché local en disco para los mosaicos vectoriales (*vector tiles*), disminuyendo el tráfico de red.

##### **Pros y Contras de las Opciones**

###### **Internacionalización Híbrida con Mobile BFF**

* **Bueno**, porque elimina la latencia de red: la pantalla móvil recibe todos los datos contextuales en un único viaje de red liviano ($<15$ KB), asegurando la respuesta en $<2$ segundos (QA-7) y respetando el FCP $<1.4$ s (QA-1).  
* **Bueno**, porque permite el formateo offline de precios y horarios locales una vez descargado el viaje, tolerando pérdidas temporales de señal celular en túneles o subterráneos.  
* **Bueno**, porque resguarda la escalabilidad del backend al delegar el renderizado visual y las operaciones matemáticas de formateo en la CPU del dispositivo cliente.  
* **Malo**, porque requiere empaquetar catálogos de cadenas traducidas dentro de la aplicación móvil, aumentando moderadamente el tamaño del instalable (bundle size).

###### **Renderizado y Traducción del Lado del Servidor**

* **Bueno**, porque centraliza todos los diccionarios y tasas en los servidores de backend.  
* **Malo**, porque cada cambio de vista o cambio de pantalla demanda viajes de red síncronos, haciendo inviable sostener $<2$ segundos bajo conexiones lentas de roaming.

###### **Consultas Granulares Directas**

* **Bueno**, porque reutiliza endpoints REST genéricos existentes.  
* **Malo**, porque viola el propósito del patrón BFF (ADR-02), obligando a la app móvil a ejecutar 3 o 4 peticiones HTTP sucesivas, destruyendo la latencia móvil y agotando la batería del dispositivo.

