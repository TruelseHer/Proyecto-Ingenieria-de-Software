# Iteración 3 \- Método ADD Sistema "Trotamundos"

## **Paso 2: Selección y justificación de Drivers Seleccionados**

#### **Objetivo de la Iteración**

* Robustecer la arquitectura del sistema mediante la selección e incorporación de tácticas de **Disponibilidad, Escalabilidad, Rendimiento y Seguridad** (*Reliability, Availability & Scalability*), asegurando la continuidad operativa ante fallos de infraestructura y picos disruptivos de tráfico. Se prioriza explícitamente el **control de acceso granular y la privacidad estricta por encima del rendimiento puro en el consumo de itinerarios**, así como la **consistencia transaccional estricta (ACID) en la confirmación de reservas frente a esquemas de disponibilidad eventual**.

#### **Clasificación y Justificación de Drivers Seleccionados**

* **QA-5 (Seguridad y Privacidad \- Control de Acceso y Datos Sensibles):**  
  * *Justificación:* El sistema procesa información personal, reservas y accesos a correos electrónicos. Exige garantizar que el 100% de los intentos de acceso no autorizados sean denegados y que los itinerarios compartidos solo puedan ser visualizados por personas explícitamente designadas. Este driver actúa como el **eje rector del primer gran *trade-off* arquitectónico**: se prioriza la seguridad y la verificación estricta de identidad/autorización en cada petición perimetral e interna por sobre la latencia mínima o el ahorro de cómputo (*Seguridad por sobre Performance*).  
* **QA-3 (Escalabilidad \- Soporte de Picos de \+500% ante Eventos Disruptivos):**  
  * *Justificación:* Ante contingencias climáticas o cancelaciones masivas de vuelos, la plataforma experimenta incrementos súbitos de hasta 5 veces su tráfico habitual de usuarios intentando reprogramar y consultar sus viajes. Resulta fundamental definir tácticas de asignación elástica de recursos de cómputo y desacoplamiento de carga para absorber la demanda sin degradar los servicios críticos.  
* **QA-2 (Disponibilidad \- Tolerancia a Fallas en Infraestructura Crítica):**  
  * *Justificación:* Al tratarse de un servicio global accesible en todo momento, el sistema debe tolerar fallas de nodos, contenedores o zonas de disponibilidad, manteniendo un tiempo de inactividad no planificado inferior a 5 minutos acumulados por mes (SLA 99.98%). Demanda tácticas de redundancia, detección temprana de caídas (*health checks*) y failover automático.  
* **QA-1 E1 (Rendimiento \- Latencias Límite de Operación):**  
  * *Justificación:* Define los umbrales máximos tolerables para la navegación: tiempo de respuesta en la web $\leq 800ms$ y *First Contentful Paint* móvil $<1.4s$. Debe balancearse activamente frente a las sobrecargas impuestas por los mecanismos criptográficos y de autorización de QA-5.

## **Paso 3: Elegir Uno o Más Elementos del Sistema a Refinar**

En correspondencia con el objetivo de robustecimiento ante fallos, picos disruptivos y control de acceso estricto (QA-2, QA-3, QA-5, QA-1), los elementos del sistema a refinar en esta iteración comprenden las capas perimetrales, los servicios transaccionales y la infraestructura de soporte operativo:

1. **trotamundos.perimeter (Edge API Gateway y BFFs):**  
   - *Tipo de refinamiento:* Asignación de tácticas de seguridad y rendimiento.  
   - *Propósito:* Refinar el punto de entrada para incorporar inspección criptográfica de tokens (autenticación y autorización granular para itinerarios compartidos \- QA-5), políticas de limitación de tasa (Rate Limiting perimetral) y terminación segura de sesiones para proteger los servicios internos ante picos de demanda.

   

2. **trotamundos.core\_travel y Repositorio Transaccional (OLTP Store):**  
   - *Tipo de refinamiento:* Configuración de concurrencia, topología de datos y elasticidad.  
   - *Propósito:* Refinar el servicio de viajes y su almacenamiento relacional para soportar el trade-off de consistencia estricta (escrituras ACID en nodo primario) combinándolo con escalabilidad de lectura (Read Replicas multi-zona y caché distribuida) para tolerar picos de tráfico del \+500% (QA-3) y garantizar disponibilidad continua ante fallas de nodos (QA-2).

   

3. **Infraestructura de Despliegue y Conectores Asíncronos (Cluster Runtime y Message Broker):**  
   - *Tipo de refinamiento:* Asignación de infraestructura y políticas de resiliencia.  
   - *Propósito:* Definir los mecanismos de orquestación de contenedores (autoescalado horizontal elástico, sondas de vida/salud \- liveness/readiness probes) y el dimensionamiento de buffers en el Message Broker para amortiguar sobrecargas transitorias sin pérdida de datos.

## **Paso 4: Elegir Uno o Más Conceptos de Diseño que Satisfagan los Drivers**

En este paso se formalizan las **tácticas arquitectónicas** orientadas a satisfacer los atributos de calidad de *runtime* (Disponibilidad QA-2, Escalabilidad QA-3, Rendimiento QA-1 y Seguridad QA-5). Asimismo, se definen y fundamentan los **trade-offs rectores** que gobiernan las decisiones de compromiso entre estos requerimientos en conflicto.

#### **1\. Análisis y Definición de Trade-offs Rectores de la Iteración**

##### **A. Trade-off Rector: Seguridad y Privacidad por sobre Rendimiento (**QA-5 **vs.** QA-1**)**

* **Tensión de Atributos:** Garantizar que el 100% de los accesos a itinerarios compartidos sean auditados y verificados bajo políticas estrictas de autorización (QA-5) requiere inspeccionar firmas asimétricas de tokens (JWT), validar listas de control de acceso (ACLs) y verificar revocaciones en tiempo real para cada solicitud entrante. Esto añade sobrecarga de procesamiento en CPU y micro-latencias en el perímetro, tensionando el objetivo de respuesta web $\leq 800ms$ (QA-1).  
* **Resolución de Compromiso:** **Se prioriza de manera estricta la Seguridad por sobre el Rendimiento puro.** No se permite eludir la validación de permisos ni servir itinerarios sensibles desde cachés públicas no autenticadas con el fin de ahorrar milisegundos. Para amortiguar la penalización en QA-1, se aplican verificaciones criptográficas optimizadas en memoria dentro del *API Gateway* y caché perimetral privada de corta duración para políticas de acceso frecuentes.

##### **B. Trade-off de Integridad: Consistencia Transaccional Estricta vs. Disponibilidad / Escalabilidad (**QA-2**,** QA-3 **vs. Integridad de Reservas)**

* **Tensión de Atributos:** Ante picos de tráfico del \+500% motivados por cancelaciones o tormentas (QA-3), una práctica usual para maximizar la disponibilidad de escritura es adoptar modelos de consistencia eventual (*Eventual Consistency*) o bases multimaster distribuidas. Sin embargo, la confirmación de reservas, cobros y modificaciones de tramos turísticos no tolera inconsistencias temporales, sobreventas ni lecturas desfasadas.  
* **Resolución de Compromiso:** **Se prioriza la Consistencia Estricta (ACID) en el núcleo transaccional de viajes y reservas.** Se mantiene la base de datos operativa relacional bajo transacciones ACID (ADR-09). Para salvaguardar la escalabilidad y disponibilidad (QA-2, QA-3), se segrega el tráfico: todas las escrituras impactan de forma consistente en el nodo primario, mientras que la demanda masiva de consultas de itinerarios se descarga hacia réplicas de lectura (*Read Replicas* multi-zona) y el Message Broker amortigua asíncronamente la ingesta externa.

#### **2\. Selección de Tácticas Arquitectónicas**

A continuación se catalogan las tácticas del SEI seleccionadas para materializar las medidas de respuesta y resolver los compromisos establecidos:

| Categoría de Atributo | Táctica SEI Seleccionada | Tipo de Táctica | Drivers / Trade-off | Justificación de Selección (Rationale) |
| :---- | :---- | :---- | :---- | :---- |
| **Disponibilidad (QA-2)** | **Detección: Ping / Echo & Heartbeat** | Detección de fallas | QA-2 | Sondas continuas (*Liveness/Readiness probes*) monitorean el estado de las instancias de Core Travel, Ingestion y el Gateway. Permite desviar tráfico instantáneamente ante la degradación de un contenedor. |
| **Disponibilidad (QA-2)** | **Recuperación: Redundant Spare / Active-Passive (Multi-AZ Failover)** | Recuperación de fallas | QA-2, Consistencia ACID | Mantiene un nodo primario relacional activo y una réplica en espera en una zona de disponibilidad secundaria. Si el nodo principal cae, se produce un failover automático en segundos, manteniendo la indisponibilidad no planificada en $\leq 5min/mes$. |
| **Disponibilidad (QA-2)** | **Recuperación: Restart** | Recuperación de fallas | QA-2 | El orquestador reinicia o sustituye automáticamente cualquier instancia de contenedor o worker que experimente fallos catastróficos o consumo anormal de memoria. |
| **Escalabilidad (QA-3)** | **Aumento de Recursos: Escalado Horizontal Elástico (HPA)** | Gestión de recursos | QA-3 | Aprovisionamiento dinámico de réplicas de contenedores para trotamundos.core\_travel y los BFFs en función de métricas de carga (CPU \> 70% o tasa de requests). Permite absorber picos de tráfico del \+500% ante contingencias. |
| **Escalabilidad (QA-3)** | **Múltiples Copias de Cómputo y Datos (*Read Replicas*)** | Gestión de recursos | QA-3, QA-1, Consistencia ACID | Balanceo de carga en cómputo y segregación de lecturas en base de datos: las consultas de itinerarios se dirigen a réplicas de lectura de PostgreSQL, liberando al nodo primario transaccional de contención. |
| **Escalabilidad (QA-3)** | **Gestión de Llegada de Eventos: Queue Requests (Buffer en Broker)** | Control de demanda | QA-3, CON-2 | El Message Broker actúa como colchón elástico para retener y desacoplar los mensajes masivos de confirmaciones e integraciones externas en situaciones de pico, impidiendo caídas en cascada. |
| **Rendimiento (QA-1)** | **Mantener Copias de Datos: Caching Distribuido** | Reducción de demanda | QA-1 | Instanciación de una capa en memoria (Redis) en los BFFs/Core para almacenar de forma temporal las consultas de itinerarios consolidados y datos de usuario frecuentes, garantizando tiempos de respuesta web $\leq 800ms$. |
| **Seguridad (QA-5)** | **Resistencia: Authenticate & Authorize Actors (JWT & Access Control Engine)** | Resistencia a ataques | QA-5, Seguridad vs. Performance | El Edge Gateway valida la autenticidad del token y un motor de control de acceso verifica exhaustivamente que el solicitante pertenezca a la lista de usuarios explícitamente autorizados para ver el itinerario. Cumple el requisito de 100% de accesos no autorizados denegados. |
| **Seguridad (QA-5)** | **Resistencia: Limit Access / Rate Limiting** | Resistencia a ataques | QA-5, QA-3 | Control de tasa de peticiones por IP y por token de usuario en el Gateway perimetral para mitigar ataques de fuerza bruta, scraping y agotamiento de recursos durante picos de uso. |
| **Seguridad (QA-5)** | **Detección y Reacción: Audit Trail (Trazabilidad de Auditoría)** | Trazabilidad y análisis | QA-5 | Registro inmutable y centralizado de todos los intentos de acceso denegados y modificaciones de itinerarios compartidos para posibilitar análisis forense. |

#### **3\. Conceptos de Diseño y Alternativas Descartadas**

* **Modelo de Consistencia Eventual Multimaster (*Multi-Master Eventual Consistency*):**  
  * *Descripción:* Permitir escrituras distribuidas en múltiples nodos de base de datos resolviendo conflictos a posteriori para maximizar la disponibilidad ante picos de demanda.  
  * *Motivo de descarte:* Viola frontalmente el trade-off de consistencia transaccional estricta (ACID) establecido para la confirmación de reservas y transacciones de viaje. Generaría anomalías críticas de negocio (doble reserva de habitaciones o pasajes) inaceptables en un sistema comercial de turismo.  
* **Bypass de Autorización mediante Caché Pública sin Re-validación:**  
  * *Descripción:* Almacenar respuestas completas de itinerarios en cachés públicas o proxies perimetrales sin requerir la validación del token de usuario en cada petición, buscando mejorar la latencia (QA-1).  
  * *Motivo de descarte:* Viola el trade-off rector de *Seguridad por sobre Performance*. Crearía una vulnerabilidad severa donde un usuario con un token vencido, revocado o modificado podría visualizar datos privados de un itinerario ajeno. Toda consulta a un itinerario compartido debe validar identidad y permisos antes de entregar la información.  
* **Escalado Vertical Exclusivo (*Scale-Up* en un único nodo de servidor):**  
  * *Descripción:* Incrementar temporalmente memoria y procesadores en un único servidor para resistir los picos de \+500% de tráfico.  
  * *Motivo de descarte:* Constituye un punto único de falla (SPOF) que destruye el objetivo de disponibilidad de $\leq 5min/mes$ de indisponibilidad (QA-2) y tiene un techo físico y económico insostenible ante la escala proyectada de usuarios.

## **Paso 5: Instanciar Elementos Arquitectónicos y Asignar Responsabilidades**

Se materializa la resolución de los dos *trade-offs* rectores: **Seguridad por sobre Rendimiento** (QA-5 vs. QA-1) y **Consistencia Transaccional ACID en reservas frente a Disponibilidad/Escalabilidad eventual** (QA-2, QA-3).

### 

#### **Responsabilidades y Definir Interfaces**

En este paso se instancian los componentes de infraestructura, módulos de software y topologías de datos que materializan las tácticas de Disponibilidad (QA-2), Escalabilidad (QA-3), Rendimiento (QA-1) y Seguridad (QA-5) sobre los subsistemas trotamundos.perimeter, trotamundos.core\_travel, trotamundos.ingestion y la capa de almacenamiento.

#### **1\. Catálogo de Elementos Instanciados y Asignación de Responsabilidades**

| Subsistema / Ámbito | Elemento Instanciado | Tipo de Elemento | Responsabilidades Asignadas (Responsibility Allocation) |
| :---- | :---- | :---- | :---- |
| **trotamundos.perimeter** | RateLimitingFilter (en Edge Gateway) | Filtro de Seguridad Perimetral | • Interceptar cada solicitud HTTP entrante y verificar cuotas de consumo por dirección IP y por identificador de usuario (userId). • Mitigar ataques de denegación de servicio (DoS) y abusos de *scraping* durante picos disruptivos (QA-5, QA-3). • Responder con código HTTP 429 Too Many Requests ante excesos de cuota. |
| **trotamundos.perimeter** | SecurityTokenValidator (en Edge Gateway) | Componente Criptográfico de Inspección | • Validar criptográficamente en memoria la firma digital (asymmetric public key) y vigencia temporal de los tokens JWT de sesión. • Aplicar el trade-off de **Seguridad por sobre Rendimiento**: bloquear el 100% de accesos con firmas inválidas antes de enrutar a los servicios internos (QA-5). |
| **trotamundos.perimeter** | DistributedCacheCluster **(Redis Cluster)** | Repositorio en Memoria / Caché Distribuida | • Almacenar copias temporales (TTL configurable) de los itinerarios consolidados frecuentes y perfiles de viajero (QA-1). • Desacargar la demanda de consultas sobre el backend durante picos de navegación web y móvil. • Almacenar listas de revocación de tokens (*blocklist*) para invalidación inmediata de sesiones. |
| **trotamundos.core\_travel** | AccessControlInterceptor | Interceptor de Autorización / Dominio | • Inspeccionar cada acceso a itinerarios compartidos (/trips/{tripId}). • Evaluar contra la base de datos o contexto seguro si el solicitante es dueño, colaborador autorizado o invitado válido (QA-5). • Denegar el acceso (HTTP 403 Forbidden) y emitir eventos de violación de acceso hacia el registro de auditoría si el usuario no figura en la lista de acceso explícita. |
| **trotamundos.core\_travel** | SecurityAuditLogger | Componente de Trazabilidad y Detección | • Capturar de manera asíncrona todos los intentos de acceso no autorizados, rechazos de permisos y modificaciones de reservas sensibles (QA-5). • Escribir registros inmutables hacia un almacén de auditoría centralizado para análisis forense. |
| **trotamundos.core\_travel** | ClusterRuntime **(Kubernetes HPA & Health Probes)** | Infraestructura de Orquestación y Cómputo | • **Escalado Elástico (HPA):** Escalar horizontalmente las instancias de Core Travel de un mínimo de 3 a un máximo de 15 pods ante picos de demanda (umbral de CPU \> 70% o consumo de requests) (QA-3). • **Detección y Recuperación:** Ejecutar sondas de vida (livenessProbe) y de preparación (readinessProbe). Si una instancia falla o no responde, el orquestador la reemplaza instantáneamente (Restart) y desvía el balanceo (QA-2). |
| **trotamundos.core\_travel** | PrimaryDatabaseNode **(PostgreSQL Primary)** | Almacén Relacional Maestro (OLTP) | • Atender con exclusividad todas las operaciones de escritura transaccional (creación de viajes, modificación de reservas, confirmación y pagos). • Garantizar el trade-off de **Consistencia Estricta (ACID)** para impedir sobreventas, pérdidas de reservas o estados temporales corruptos. • Replicar sincrónicamente hacia el nodo de reserva y asincrónicamente hacia las réplicas de lectura. |
| **trotamundos.core\_travel** | StandbyDatabaseNode **(Multi-AZ Standby)** | Almacén de Respaldo Pasivo (*Hot Standby*) | • Mantener sincronización continua con el nodo primario ubicado en una zona de disponibilidad secundaria (*Multi-AZ*). • Promoverse de forma automática a nodo primario (*Automatic Failover*) ante caída física del maestro, garantizando el SLA de indisponibilidad $\leq 5min/mes$ (QA-2). |
| **trotamundos.core\_travel** | ReadReplicasPool **(PostgreSQL Read Replicas)** | Grupo de Réplicas de Lectura | • Atender el tráfico masivo de consultas del panel de itinerarios del usuario (GET /trips/active). • Balancear la carga de lectura liberando de contención al nodo transaccional primario ante picos del \+500% (QA-3, QA-1). |
| **Conectores / Runtime** | MessageBrokerCluster **(Kafka / RabbitMQ Particionado)** | Conector de Mensajería Elástico | • Implementar particionamiento y persistencia en disco de eventos de reserva (ReservationDetectedEvent). • Actuar como búfer elástico amortiguador para retener las reservas masivas de ingesta durante saturaciones de tráfico sin pérdida de datos (QA-3). |

#### **2\. Definición de Relaciones e Interfaces de Infraestructura y Operación**

##### **A. Interfaz de Monitoreo y Salud (Health Check Interface \-** QA-2**)**

* **Endpoint:** GET /actuator/health (Expositor HTTP estandarizado en cada microservicio).  
* **Protocolo:** HTTP/1.1 síncrono.  
* **Comportamiento:**

  * readinessProbe: Sondea cada 5 segundos si el servicio puede conectarse a PostgreSQL y al Message Broker. Si falla, el Gateway deja de enviarle tráfico.  
  * livenessProbe: Sondea cada 10 segundos el estado del hilo de procesamiento. Si devuelve error o *timeout*, el clúster destruye y reinicia el pod (Restart).

##### **B. Política de Enrutamiento de Base de Datos (Routing DataSource \- Consistencia vs. Escalabilidad)**

* **Patrón de Implementación:** AbstractRoutingDataSource (en la capa de Repositorios de Core Travel).  
* **Regla de Enrutamiento:**  
  * Métodos anotados con @Transactional(readOnly \= false) (Creación manual US-04, cancelación, estados de viaje US-05) se direccionan hacia la IP del PrimaryDatabaseNode (JDBC, puerto 5432\) garantizando aislamiento transaccional estricto.  
  * Métodos anotados con @Transactional(readOnly \= true) (Consulta del itinerario consolidador GET /trips) se balancean hacia el pool de ReadReplicasPool (JDBC, réplicas 1..N).

##### **C. Especificación del Mecanismo de Control de Acceso (Enforcement Interface \-** QA-5**)**

* **Protocolo:** Interceptor interno de seguridad a nivel de capa de aplicación.  
* **Contrato de Validación de Acceso:**

public interface AccessControlEngine {  
    /\*\*  
     \* Valida el acceso estricto a un itinerario compartido (QA-5).  
     \* @param requesterUserId Identificador extraído del token JWT validado.  
     \* @param targetTripId Identificador del viaje solicitado.  
     \* @return true si el usuario es propietario o figura en la lista de autorizados; de lo contrario false.  
     \*/  
    boolean canAccessTrip(UUID requesterUserId, UUID targetTripId);

* **Acción ante rechazo:** Emisión sincrónica de alerta de seguridad a SecurityAuditLogger y corte inmediato del flujo con excepción HTTP 403 Forbidden.

##### **D. Configuración de Políticas de Auto-Escalado (Kubernetes HPA Spec \-** QA-3**)**

* **Métrica objetivo:** Utilización promedio de CPU \$\\ge 70\\%\$ o ratio de peticiones entrantes $>250req/s$ por pod.  
* **Rango de réplicas:** Mínimo \= 3 réplicas (alta disponibilidad basal Multi-AZ); Máximo \= 15 réplicas (absorción del pico de \+500%).  
* **Política de estabilización:** Ventana de escalado descendente (*scale-down stabilization window*) de 300 segundos para evitar oscilaciones de terminación abrupta de pods (*flapping*).

## **Paso 6: Vistas y ADR Registrados**

* De los ADR-10 al ADR-12 corresponden a la Iteración 3

[ADR-10](../adr/ADRS.md#adr-10-topología-de-base-de-datos-primario-réplica-con-failover-multi-az-y-réplicas-de-lectura)
[ADR-11](../adr/ADRS.md#adr-11-priorización-de-seguridad-y-validación-criptográfica-en-memoria-en-el-api-gateway)
[ADR-12](../adr/ADRS.md#adr-12-autoescalado-horizontal-elástico-hpa-y-amortiguación-de-ingesta-mediante-message-broker)

* Vistas  
  * Vista 1 Deployment  
  * Vista 2 C\&C

## **Paso 7: Analizar el Diseño, Evaluar Resultados y Actualizar el Design Backlog**

### **1\. Evaluación del Objetivo de la Iteración y Revisión del Diseño**

* **Objetivo de la Iteración:** Robustecer la arquitectura del sistema incorporando tácticas de **Disponibilidad, Escalabilidad, Rendimiento y Seguridad** (*Reliability, Availability & Scalability*), garantizando la operación continua ante fallos y picos disruptivos del \+500% de tráfico, priorizando la seguridad y privacidad sobre el rendimiento en itinerarios compartidos (QA-5 vs. QA-1), y preservando la consistencia transaccional estricta (ACID) en reservas y pagos frente a esquemas de disponibilidad eventual (QA-2, QA-3).  
    
* **Evaluación General:** El objetivo de la iteración ha sido **cumplido satisfactoriamente**. Las estructuras de despliegue y los componentes de tiempo de ejecución modelados en el Paso 6 proporcionan el soporte táctico necesario para tolerar caídas de infraestructura y absorber ráfagas masivas de demanda. Se formalizaron los compromisos en los registros **ADR-10** (Topología Primario-Réplica Multi-AZ y Read Replicas), **ADR-11** (Seguridad y Verificación Criptográfica en Memoria) y **ADR-12** (Autoescalado Elástico HPA y Amortiguación en Message Broker), logrando una arquitectura equilibrada y trazable respecto a los drivers de calidad.


### **2\. Justificación del Cumplimiento de Drivers y Definición de Abordaje (*Definition of Addressed*)**

Siguiendo el criterio metodológico del SEI (Cervantes & Kazman), los drivers se evalúan en función de si las decisiones arquitectónicas adoptadas son suficientes para satisfacer las medidas de respuesta o si requieren validación empírica en etapas de testing/operación:

| Driver / Escenario | Estado al Finalizar It. 3 | Evidencia Arquitectónica y Justificación (Definition of Addressed) |
| :---- | :---- | :---- |
| **QA-5** *(Seguridad: Control de Acceso Granular y Privacidad 100%)* | **Abordado Completamente** | Se diseñó el esquema de seguridad integral: terminación TLS, RateLimitingFilter y validación criptográfica en memoria de JWT en el Edge Gateway, junto con el interceptor formal IAccessControlEngine en Core Travel y el registro inmutable en Security Audit Store (ADR-11). Se garantiza el rechazo del 100% de peticiones no autorizadas y el aislamiento estricto de itinerarios compartidos.  |
| **QA-2** *(Disponibilidad: SLA $\leq 5min/mes$ de indisponibilidad)* | **Abordado Completamente** | Se aplicaron las tácticas de detección (*liveness/readiness probes*) y recuperación automática (*restart* por K8s). La persistencia transaccional se protege con un nodo *Hot Standby* replicado sincrónicamente en una zona de disponibilidad secundaria (*Multi-AZ*), asegurando un failover automático en menos de 60 segundos ante desastres físicos de datacenter (ADR-10).  |
| **QA-3** *(Escalabilidad: Picos de \+500% ante eventos disruptivos)* | **Parcialmente Abordado** | Se instanciaron las tácticas de *Increase Resources* (escalado elástico HPA de 3 a 15 pods para Core Travel y BFFs), el pool de réplicas de lectura de PostgreSQL para liberar al nodo primario (ADR-10) y el buffer en Kafka para ingesta masiva (ADR-12). Permanece **parcial** porque, tal como establece la cátedra (*«QA-3 sigue parcial: faltan tests que validen sus response measures»*), la confirmación cuantitativa de latencias $<3s$ ante picos masivos requiere pruebas de carga de estrés (*load testing*) sobre el clúster.  |
| **QA-1** *(Rendimiento: Latencia Web $\leq 800ms$, Móvil $<1.4s$)* | **Parcialmente Abordado** | Se instanció la capa de caché distribuida en memoria (Redis Cluster) en los BFFs para absorber consultas frecuentes y se redujo la sobrecarga de validación de tokens a nivel de Gateway (ADR-11). Permanece **parcial** debido al trade-off asumido (*Seguridad por sobre Performance*) y a la necesidad de contrastar los tiempos reales de red y renderizado móvil en el prototipo desplegado.  |
| **US-04, US-05, US-06, US-02, CRN-2** | **Abordados Completamente** | Cerrados formalmente en iteraciones previas; se mantiene su soporte estructural y transaccional sin regresiones arquitectónicas. |
| **US-01, US-03, CON-2** | **Parcialmente Abordados** | Mantienen su estado funcional desde la Iteración 2 a la espera de implementación concreta y contratos detallados con proveedores de aerolíneas/hoteles en el código fuente. |

#### 

#### **Actualización del Design Backlog**

| Categoría | Especificación | No abordado (Hasta el momento) | Parcialmente abordado | Abordado Completamente |
| :---- | :---- | :---- | :---- | :---- |
| Incumbencias Arquitectónicas | CNR 1 |  |   |  x |
|   | CNR 2 |  |   |  x |
|   | CNR 3 |  |   |  x |
|   | CNR 4 | x |   |   |
|   | CNR 5 | x |   |   |
| Restricciones | CON-1 |  |   |  x |
|   | CON-2 |  |  x |   |
| Atributos de Calidad | QA-1 |  |  x |   |
|   | QA-2 |  |   |  x |
|   | QA-3 |  |  x |   |
|   | QA-4 | x |   |   |
|   | QA-5 |  |   |  x |
|   | QA-6 | x |   |   |
|   | QA-7 | x |   |   |
| Requerimientos Funcionales Primarios | US-01 |  |  x |   |
|   | US-02 |  |   |  x |
|   | US-03 |  | x  |   |
|   | US-04 |  |   |  x |
|   | US-05 |  |   |  x |
|   | US-06 |  |   |  x |
|   | US-07 | x |   |   |

### 

