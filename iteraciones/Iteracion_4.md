# Iteración 4 \- Método ADD Sistema "Trotamundos"

## **Paso 2: Selección y justificación de Drivers Seleccionados**

#### **Objetivo de la Iteración**

* Diseñar las estructuras arquitectónicas internas, pipelines de datos y contratos de integración para materializar la **Recopilación y Consulta de Datos Analíticos orientados a Monetización (US-07)** y la provisión de **Extensibilidad para Módulos Avanzados (CRN-5)** —en particular el motor de recomendaciones personalizadas—, completando a su vez la **Interoperabilidad de Ingesta Externa (QA-4, CON-2, US-01, US-03)** y garantizando la **Evolucionabilidad modular (QA-6, CRN-4)** sin degradar la consistencia ni el rendimiento del núcleo transaccional.

**Clasificación y Justificación de Drivers Seleccionados**

* **US-07 (Recopilación y Consulta de Datos Analíticos / Monetización):**  
  - *Justificación:* Habiéndose establecido en la Iteración 1 la segregación física entre el almacén operativo y el analítico (ADR-04) y el aislamiento del tráfico transaccional, resta modelar la lógica interna del subsistema trotamundos.analytics. Requiere diseñar la ingesta asíncrona de eventos históricos (TripLifecycleEvents), las rutinas de desidentificación/anonimización estricta de información personal (PII) y los esquemas de consulta agregada multidimensional sobre destinos y cancelaciones para el equipo de monetización, garantizando cero contención sobre el motor OLTP.  
* **CRN-5 (Extensibilidad para integración de módulos avanzados \- Motor de Recomendaciones):**  
  - *Justificación:* Representa el propósito de diseño fundacional de Trotamundos ("sistema desde cero enfocado en realizar recomendaciones para viajeros"). Su inclusión es prioritaria para definir las interfaces de desacoplamiento, eventos de contexto y puntos de enganche (*extension points*) que permitan acoplar un motor de sugerencias personalizadas de viaje (alimentado por los patrones analíticos de US-07) sin invadir ni alterar las invariantes transaccionales de trotamundos.core\_travel.  
* **QA-4 (Interoperabilidad \- Integración precisa con proveedores y correo)**  
  - *Justificación:* Driver no abordado en las iteraciones previas que evalúa el intercambio de datos con fuentes heterogéneas. Exige asegurar el 100% de precisión en el mapeo sintáctico/semántico bidireccional entre las confirmaciones/APIs de aerolíneas u hoteles y los modelos canónicos internos bajo esquemas estandarizados (JSON/REST), absorbiendo variaciones de formato sin demandar intervención manual del usuario.  
* **CON-2 / US-01 / US-03 (Consolidación de la Ingesta Externa y Adaptadores de Terceros):**  
  - *Justificación:* Mantenían estado **parcialmente abordado** desde la Iteración 2\. Se seleccionan conjuntamente para cerrar el diseño de la Capa Anticorrupción (ACL) en trotamundos.ingestion: se especifican los adaptadores concretos por tipo de proveedor (vuelos, hoteles, rentadoras) y los mecanismos de resiliencia frente a caídas de terceros (reintentos desacoplados, circuit breaker y buffers en el broker de mensajería).  
* **QA-6 (Evolucionabilidad y Mantenibilidad \- Extensibilidad para Nuevos Transportes):**  
  - *Justificación:* Demanda asegurar que la adición futura de medios de transporte adicionales (trenes, autobuses de larga distancia) se concrete en menos de 2 sprints sin refactorizar el código transaccional core. Permite validar y asentar la jerarquía polimórfica de reservas y la extensibilidad modular de adaptadores.  
* **CRN-4 (Modularización para distribución de trabajo y gobernanza ADR):**  
  - *Justificación:* Define los límites físicos de paquetes, proyectos y repositorios entre los equipos de desarrollo (especialmente entre el equipo analítico/monetización y el equipo transaccional), garantizando trazabilidad formal entre decisiones (ADRs) y contratos (*Spec-Driven*).

## **Paso 3: Elegir Uno o Más Elementos del Sistema a Refinar**

En correspondencia con el objetivo de la iteración y los drivers arquitectónicos seleccionados (US-07, CRN-5, QA-4, CON-2, US-01, US-03, QA-6 y CRN-4), los elementos del sistema a descomponer y refinar en esta cuarta iteración corresponden a los subsistemas analíticos, de ingesta y los puntos de extensión del dominio:

### **1\. trotamundos.analytics (Subsistema Analítico y de Monetización)**

* **Tipo de refinamiento:** Descomposición interna (vertical) y refinamiento de flujo de datos.  
* **Propósito:** Descomponer el subsistema instanciado a nivel macro en la Iteración 1 en sus módulos internos de procesamiento para materializar **US-07**:  
  - Un componente de consumo de eventos asíncronos (event\_consumer) que capture las transiciones del ciclo de vida de los viajes desde el Message Broker.  
  - Un módulo de anonimización y transformación de datos (data\_anonymizer) que desasocie la información personal identificable (PII) antes de la persistencia histórica.  
  - El motor de agregación y reportes de monetización (reporting\_engine) junto con sus adaptadores hacia la base de datos analítica (persistence\_olap), garantizando que las consultas analíticas pesadas no compitan por recursos con el almacén operativo (CRN-3 / ADR-04).

### **2\. trotamundos.recommendations (Módulo Avanzado de Recomendación Personalizada)**

* **Tipo de refinamiento:** Creación, contextualización y descomposición arquitectónica.  
* **Propósito:** Dar soporte directo a **CRN-5** y materializar el propósito de diseño fundacional de la plataforma (sistema enfocado en recomendaciones para viajeros):  
  - Definir la estructura interna del nuevo motor de recomendación inteligente (recommendation\_engine).  
  - Establecer los puntos de extensión desacoplados y los conectores asíncronos para consumir patrones y métricas del subsistema analítico (trotamundos.analytics), cruzándolos con el contexto del viaje activo del usuario sin introducir acoplamiento síncrono ni dependencias circulares con trotamundos.core\_travel.

### **3\. trotamundos.ingestion (Subsistema de Ingesta e Integración Externa)**

* **Tipo de refinamiento:** Refinamiento estructural interno y completitud de adaptadores.  
* **Propósito:** Cerrar el estado de los drivers parcialmente abordados (**CON-2, US-01, US-03**) y satisfacer **QA-4** (Interoperabilidad precisa):  
  - Refinar la Capa Anticorrupción (ACL) mediante la especialización de adaptadores dedicados (AirlineAdapter, HotelAdapter, RentalAdapter) con contratos normalizados.  
  - Incorporar políticas tácticas de resiliencia (circuit breakers, colas de descarte/dead-letter queues y reintentos exponenciales) para tolerar la latencia o indisponibilidad de proveedores externos sin degradar la precisión del mapeo sintáctico/semántico (100% de exactitud en el cruce de datos).

### **4\. Jerarquía de Clases y Extensibilidad del Dominio (trotamundos.core\_travel.domain)**

* **Tipo de refinamiento:** Refinamiento de diseño orientado a objetos y modularización.  
* **Propósito:** Abordar **QA-6** (Evolucionabilidad) y **CRN-4** (Gobernanza y modularización):  
  - Extender la jerarquía polimórfica de reservas dentro del modelo de dominio (Reservation como aggregate base) para admitir nuevos medios de transporte (TrainReservation, BusReservation) de forma aislada.  
  - Definir la separación en submódulos y librerías contractuales versionadas que permitan a equipos independientes desarrollar y desplegar integraciones en menos de dos sprints sin modificar el código fuente central de viajes.

## **Paso 4: Elegir Uno o Más Conceptos de Diseño que Satisfagan los Drivers**

En este paso se formalizan los conceptos arquitectónicos (patrones arquitectónicos, patrones de integración, tácticas de diseño y estilos de extensibilidad) necesarios para gobernar el procesamiento analítico y de monetización (US-07), el motor desacoplado de recomendaciones (CRN-5), la interoperabilidad precisa y resiliente con terceros (QA-4, CON-2, US-01, US-03) y la evolución modular del sistema (QA-6, CRN-4).

### 1\. Identificación y Selección de Conceptos de Diseño

| Concepto de Diseño / Patrón | Tipo de Concepto | Drivers que Ataca | Justificación de Selección (Rationale) |
| :---- | :---- | :---- | :---- |
| **Pipeline de Ingesta Asíncrona Desacoplada (Event-Driven Data Pipeline)** | Patrón de Integración y Flujo de Datos | US-07, CRN-3 | Permite que trotamundos.analytics consuma de manera reactiva y no bloqueante los eventos emitidos por el núcleo transaccional (TripLifecycleEvents) desde el Message Broker. Garantiza que la recolección masiva de métricas no degrade la latencia del usuario activo ni compita por conexiones en el almacenamiento OLTP. |
| **Táctica de Seguridad: Sanitización y Seudonimización en Tránsito (Data Masking & De-identification)** | Táctica de Seguridad y Privacidad | US-07, QA-5 | Aplica transformaciones de hashing unidireccional y eliminación de identificadores directos (PII como emails, nombres y documentos) en el momento de la ingesta previa al almacenamiento analítico, garantizando que el repositorio histórico y los reportes de monetización sean legalmente ciegos a la identidad individual. |
| **Arquitectura Basada en Plugins / Extensión Desacoplada (Microkernel / Extension Points)** | Patrón Arquitectónico de Extensibilidad | CRN-5, QA-6 | Estructura el módulo de recomendaciones (trotamundos.recommendations) como un componente periférico desacoplado que expone puntos de extensión claros. Permite evolucionar algoritmos de sugerencia de rutas/hoteles y retroalimentación interactiva (*Human-in-the-Loop*) consumiendo vistas agregadas del almacén analítico sin introducir acoplamiento síncrono ni dependencias directas con el núcleo transaccional de viajes. |
| **Patrón Circuit Breaker & Retry con Dead-Letter Queue (DLQ)** | Táctica de Tolerancia a Fallos y Resiliencia | QA-4, CON-2, US-03 | Aísla caídas transitorias y tiempos de respuesta lentos en las APIs externas de aerolíneas y hoteles. Si un proveedor externo falla o supera umbrales de latencia, el disyuntor interrumpe las llamadas para evitar el agotamiento de hilos y desvía los mensajes fallidos a una cola de reprocesamiento (DLQ) para reintento diferido, garantizando que no se pierdan datos de reservas. |
| **Extensibilidad Polimórfica de Dominio y Contratos Abiertos (Open-Closed Principle / Strategy Pattern)** | Patrón de Diseño Orientado a Objetos / Mantenibilidad | QA-6, CRN-4 | Permite modelar tipos de transporte emergentes (trenes, autobuses) mediante subclases especializadas de la abstracción Reservation e implementaciones concretas de interfaces de adaptación sin alterar los métodos de cálculo ni las invariantes del agregado raíz Trip, posibilitando su integración en menos de 2 sprints. |

### **2\. Conceptos de Diseño Descartados**

#### **A. Ingesta Analítica Directa vía Consultas Periódicas Batch a la Base de Datos Transaccional (Direct OLTP Polling / Periodic Batch Dump)**

* **Descripción:** Ejecutar procesos cron/batch nocturnos que lean directamente mediante sentencias SELECT \* masivas sobre la base de datos PostgreSQL operativa para sincronizar el Data Warehouse.  
* **Motivo de descarte:** Viola frontalmente el aislamiento transaccional consagrado en el ADR-04 (CRN-3). La ejecución de consultas pesadas de extracción bloquea tablas, satura el pool de conexiones e incrementa severamente los tiempos de respuesta del panel en vivo de los viajeros. Se adopta la alimentación continua guiada por eventos a través del broker de mensajería.

#### **B. Integración Sincrónica Request-Reply entre el Motor de Recomendaciones y el Servicio Transaccional**

* **Descripción:** Que el trotamundos.core\_travel invoque por HTTP/gRPC al módulo de recomendaciones cada vez que un usuario consulta o actualiza un itinerario para adjuntarle alternativas en el mismo viaje de red.  
* **Motivo de descarte:** Genera un acoplamiento temporal crítico y duplica la latencia de respuesta del usuario final, arriesgando el límite web \$\\le 800\$ ms. Si el motor analítico de recomendaciones se ralentiza o cae, comprometería la disponibilidad del itinerario transaccional. La interacción se descarta a favor de una composición asíncrona a nivel de capa BFF o lectura desacoplada de sugerencias cacheadas.

#### **C. Tolerancia a Fallas Exclusiva por Reintentos Síncronos Indefinidos (Tight Synchronous Retries sin Circuit Breaker)**

* **Descripción:** Reintentar de forma inmediata y continuada las llamadas HTTP salientes hacia las APIs de turismo externas que reporten códigos 5xx o timeout.  
* **Motivo de descarte:** Ante la caída prolongada de una aerolínea o cadena hotelera, agota rápidamente los hilos de ejecución de los workers de ingesta (*thread starvation*), degradando el rendimiento general y generando tormentas de reintentos sobre proveedores saturados. Se descarta en favor de disyuntores de circuito (*Circuit Breaker*) con retroceso exponencial (*Exponential Backoff*) y enrutamiento hacia Dead-Letter Queues.

#### **D. Esquema de Entidad Única Genérica sin Polimorfismo para Reservas (Universal Generic Reservation Entity)**

* **Descripción:** Utilizar una única clase Reservation con una columna genérica tipo payload\_json o campos nulos masivos para todos los tipos de servicios existentes y futuros.  
* **Motivo de descarte:** Destruye la cohesión del modelo de dominio enriquecido (ADR-05), degrada la tipificación estricta de invariantes de negocio y traslada la validación sintáctica a lógica condicional compleja (*if/else* o *switch* extensos), impidiendo la extensibilidad en menos de 2 sprints (QA-6).

## **Paso 5: Instanciar Elementos Arquitectónicos y Asignar Responsabilidades**

En este paso se instancian los componentes de software, conectores de mensajería, almacenes de datos y módulos de dominio que materializan los conceptos de diseño seleccionados en el Paso 4\. Se distribuyen las responsabilidades sobre los subsistemas trotamundos.analytics (US-07), trotamundos.recommendations (CRN-5), trotamundos.ingestion (QA-4, CON-2, US-01, US-03) y trotamundos.core\_travel.domain (QA-6, CRN-4).

### **1\. Catálogo de Elementos Instanciados y Asignación de Responsabilidades**

| Subsistema / Ámbito | Elemento Instanciado | Tipo de Elemento | Responsabilidades Asignadas (Responsibility Allocation) |
| :---- | :---- | :---- | :---- |
| **trotamundos.core\_travel** | **TripLifecycleEventPublisher** | Productor de Eventos / Capa de Dominio | • Capturar cada transición de estado en el ciclo de vida de los viajes (creación, confirmación, reprogramación, cancelación o archivado). • Serializar y publicar de forma no bloqueante el evento TripLifecycleEvent hacia el tópico particionado trip-lifecycle-events del Message Broker para su explotación desacoplada (US-07) |
| **trotamundos.analytics** | **TripEventStreamConsumer** | Consumidor Asíncrono / Listener | • Escuchar de forma continua el tópico trip-lifecycle-events desde el Message Broker. • Deserializar los mensajes entrantes y desacoplar temporalmente la carga de recolección de métricas respecto a las transacciones de los usuarios activos. |
| **trotamundos.analytics** | **DataAnonymizerFilter** | Módulo de Sanitización / Privacidad | • Interceptar los eventos en tránsito previo a su persistencia analítica (QA-5, US-07). • Eliminar identificadores personales directos (nombres, PNRs de tickets y direcciones de correo electrónico). • Aplicar algoritmos de *hashing* unidireccional con clave/salteado (SHA-256 \+ salt) sobre el userId para derivar un seudónimo inmutable (pseudonymId) que permita análisis de cohortes sin vulnerar la identidad real. |
| **trotamundos.analytics** | **MonetizationAggregationEngine** | Motor de Cálculo / Servicio Analítico | • Computar agregaciones analíticas pesadas por lotes o ventanas temporales: volúmenes de demanda por destino, tasas de disrupción por aerolínea y tendencias estacionales (US-07). • Exponer endpoints internos de reportería y consulta multidimensional para el equipo de monetización sin competir por recursos de cómputo con las transacciones operativas. |
| **trotamundos.analytics** | **Base de Datos Analítica (OLAP Store / ClickHouse o Timescale)** | Almacén Histórico Desnormalizado (OLAP) | • Persistir métricas de viajes agregadas en esquemas columnares y particionados por rangos temporales (CRN-3, ADR-04). • Atender consultas analíticas complejas y pesadas optimizando operaciones de lectura masiva. |
| **trotamundos.recommendations** | **RecommendationPluginHost** | Núcleo de Extensibilidad (Microkernel / Host) | Actuar como orquestador y punto de extensión desacoplado (CRN-5). Gestionar el ciclo de vida de los plugins de recomendación cargados en el sistema mediante la interfaz unificada RecommendationStrategy. Servir sugerencias contextualizadas en tiempo de lectura a los BFFs perimetrales con latencia menor a 50 ms. |
| **trotamundos.recommendations** | **RecommendationStrategy (Subtipos)** | Interfaces de Estrategia / Algoritmos | CollaborativeFilteringStrategy: Sugiere alojamientos y destinos basándose en correlaciones y perfiles similares extraídos periódicamente del almacén analítico. DestinationTrendingStrategy: Sugiere rutas alternativas basadas en los destinos con mayor tracción comercial y menor tasa de contingencias reportadas por el motor de monetización. |
| **trotamundos.recommendations** | **RecommendationCache (Redis Key-Value)** | Repositorio de Caché en Memoria | • Almacenar sugerencias precalculadas asociadas al contexto del viajero con tiempo de vida (TTL) configurable (QA-1). • Permitir que el Edge Gateway / BFF entregue recomendaciones sin invocar motores pesados en tiempo real. |
| **trotamundos.ingestion** | **CircuitBreakerGuard (Resilience4j)** | Filtro Táctico de Tolerancia a Fallos | • Supervisar el ratio de error y latencia en las llamadas HTTP salientes hacia proveedores turísticos externos (aerolíneas, hoteles, agencias) (QA-4, CON-2). • Abrir el circuito de comunicación si la tasa de fallo supera el 50% o la latencia supera los 3 segundos, impidiendo el agotamiento del pool de conexiones y el bloqueo de hilos. |
| **trotamundos.ingestion** | **IngestionDLQDispatcher** | Manejador de Contingencias / Despachador | • Capturar excepciones emitidas por el Circuit Breaker o agotamiento de reintentos. • Encapsular el mensaje original junto con el contexto del error y enrutarlo hacia la Dead-Letter Queue (DLQ) en el Message Broker para su posterior reprocesamiento asíncrono con retroceso exponencial (*exponential backoff* con *jitter*). |
| **trotamundos.core\_travel.domain** | **Jerarquía Polimórfica de Dominio (Reservation)** | Entidades y Clases de Negocio | • Reservation (Clase Base Abstracta): Modela la interfaz invariante de reserva (identificadores, códigos de confirmación, cálculo de lapsos temporales y vinculación a agencias). • TrainReservation y BusReservation (Especializaciones): Subclases modulares que encapsulan especificidades de transporte terrestre (número de tren/coche, terminal de transbordo, líneas de autobús) cumpliendo el principio abierto/cerrado sin modificar el agregado raíz Trip (QA-6, CRN-4). |

### 

### **2\. Definición Formal de Interfaces y Gobernanza Contractual (*Spec-Driven Development*)**

En cumplimiento estricto del enfoque *Spec-Driven Development* (CRN-2), se especifican las interfaces y contratos canónicos que gobiernan los nuevos elementos instanciados:  

[A. Contrato Asíncrono de Ciclo de Vida: TripLifecycleEvent (JSON Schema / AsyncAPI)](../spec/Spec-Driven_Devolpement.md#a-contrato-asíncrono-de-ciclo-de-vida-triplifecycleevent-json-schema--asyncapi)  
[B. Contrato OpenAPI del Servicio Analítico de Monetización: AnalyticsAPI.yaml](../spec/Spec-Driven_Devolpement.md#b-contrato-openapi-del-servicio-analítico-de-monetización-analyticsapiyaml)  
[C. Interfaz de Software para Plugins de Recomendación (RecommendationStrategy)](../spec/Spec-Driven_Devolpement.md#c-interfaz-de-software-para-plugins-de-recomendación-recommendationstrategy)  
[D. Especificación Táctica del Circuit Breaker en Ingesta (Configuración de Resiliencia)](../spec/Spec-Driven_Devolpement.md#d-especificación-táctica-del-circuit-breaker-en-ingesta-configuración-de-resiliencia)


## **Paso 6: Vistas y ADR Registrados**

* De los ADR-13  al ADR-17 corresponden a la Iteración 4

[ADR-13](../adr/ADRS.md#adr-13-ingesta-analítica-asíncrona-guiada-por-eventos-para-consultas-de-monetización)
[ADR-14](../adr/ADRS.md#adr-14-anonimización-y-desidentificación-de-datos-en-tránsito-para-el-repositorio-analítico)
[ADR-15](../adr/ADRS.md#adr-15-desacoplamiento-del-motor-de-recomendaciones-mediante-arquitectura-de-plugins-y-puntos-de-extensión)
[ADR-16](../adr/ADRS.md#adr-16-resiliencia-de-ingesta-externa-mediante-circuit-breaker-y-dead-letter-queue-dlq)
[ADR-17](../adr/ADRS.md#adr-17-extensibilidad-de-transportes-mediante-modelo-abiertocerrado-y-contratos-modulares)

* Vistas  
  * Vista 1 C\&C  
  * Vista 2 Módulos  
  * Vista 3 Secuencia

## **Paso 7: Analizar el Diseño, Evaluar Resultados y Actualizar el Design Backlog**

### **1\. Evaluación del Objetivo de la Iteración y Revisión del Diseño**

* **Objetivo de la Iteración:** Diseñar las estructuras arquitectónicas internas, pipelines de datos y contratos de integración para materializar la **Recopilación y Consulta de Datos Analíticos orientados a Monetización (US-07)** y la provisión de **Extensibilidad para Módulos Avanzados (CRN-5)** —en particular el motor de recomendaciones personalizadas—, completando a su vez la **Interoperabilidad de Ingesta Externa (QA-4, CON-2, US-01, US-03)** y garantizando la **Evolucionabilidad modular (QA-6, CRN-4)** sin degradar la consistencia ni el rendimiento del núcleo transaccional.  
* **Evaluación General:** El objetivo de la iteración ha sido **cumplido satisfactoriamente**. A través de las vistas arquitectónicas y los registros ADR-13 a ADR-17, se cerró el frente de interoperabilidad con fuentes externas mediante mecanismos de protección activa (Circuit Breaker) y contingencia asíncrona (DLQ). Se consolidó el pipeline reactivo y anónimo de eventos para consultas pesadas de monetización en un almacén OLAP dedicado, se materializó la arquitectura basada en plugins para el motor de recomendaciones como propósito fundacional del sistema, y se aseguró la extensibilidad polimórfica del dominio para incorporar medios de transporte futuros (trenes y autobuses) sin alterar el núcleo transaccional de viajes.

### **2\. Justificación del Cumplimiento de Drivers y Definición de Abordaje (*Definition of Addressed*)**

| Driver / Escenario | Estado al Finalizar It. 4 | Evidencia Arquitectónica y Justificación (Definition of Addressed) |
| :---- | :---- | :---- |
| **US-07** *(Recopilación y Consulta de Datos Analíticos / Monetización)* | **Abordado Completamente** | Se diseñó el pipeline asíncrono no bloqueante guiado por eventos (trip-lifecycle-events sobre Kafka) consumido por trotamundos.analytics (**ADR-13**). Se garantizó la desidentificación criptográfica en tránsito (**ADR-14**) y la persistencia segregada en el motor columnar OLAP, permitiendo consultas multidimensionales pesadas para monetización sin generar contención sobre la base transaccional PostgreSQL. |
| **CRN-5** *(Extensibilidad para Módulos Avanzados \- Recomendaciones)* | **Abordado Completamente** | Se formalizó el módulo autónomo trotamundos.recommendations basado en el patrón Strategy / Plugins (**ADR-15**). El motor consume patrones históricos agregados desde el almacén OLAP y precalcula sugerencias en caché Redis, cumpliendo el propósito de diseño fundacional de la plataforma sin acoplamiento síncrono ni impacto sobre la latencia transaccional. |
| **QA-4** *(Interoperabilidad precisa con APIs y Correos)* | **Abordado Completamente** | Se estructuraron los adaptadores dedicados en la Capa Anticorrupción (ACL) y se incorporaron disyuntores de circuito (Circuit Breakers) y colas de desvío (DLQ) (**ADR-16**). Se garantiza que las fallas de terceros no bloqueen la ingesta ni provoquen pérdida de confirmaciones, logrando sincronización e intercambio bajo formatos canónicos estandarizados. |
| **CON-2** *(Integración obligatoria con proveedores turísticos)* | **Abordado Completamente** | Promovido desde el estado parcial previo. Quedaron sellados los adaptadores específicos por tipo de proveedor (AirlineAdapter, HotelAdapter, RentalAdapter), sus contratos canónicos y la política de tolerancia a fallos ante caídas externas (**ADR-16**). |
| **US-01** *(Escaneo y Detección Automática de Correos)* | **Abordado Completamente** | Promovido desde parcial. Se completó el pipeline de extracción sintáctica, validación frente a listas blancas y publicación del evento normalizado ReservationDetectedEvent con soporte de reprocesamiento diferido ante contingencias. |
| **US-03** *(Integración con Sistemas de Viajes vía API)* | **Abordado Completamente** | Promovido desde parcial. Se formalizó la sincronización continua mediante adaptadores ACL protegidos por Circuit Breakers y DLQ, asegurando la llegada de actualizaciones de proveedores en menos de 5 minutos sin riesgo de caídas en cascada. |
| **QA-6** *(Evolucionabilidad: Extensibilidad para Nuevos Transportes)* | **Abordado Completamente** | Se validó en la vista de módulos de dominio la extensión polimórfica abierta/cerrada (**ADR-17**). Nuevos transportes (TrainReservation, BusReservation) heredan de la clase base Reservation e implementan interfaces de adaptación sin alterar los métodos ni invariantes del agregado Trip, permitiendo su construcción y despliegue en menos de 2 sprints. |
| **CRN-4** *(Modularización y Gobernanza ADR)* | **Abordado Completamente** | Se estructuraron los paquetes de código y contratos independientes entre subsistemas (analítica, recomendaciones, ingesta y core), habilitando el desarrollo concurrente y manteniendo la trazabilidad total mediante los registros formales ADR-13 a ADR-17. |
| **QA-1** *(Rendimiento: Web $\leq 800$ ms, Móvil $<1.4$ s)* | **Parcialmente Abordado** | Las decisiones estructurales (caché de recomendaciones en Redis, segregación OLTP/OLAP y llamadas no bloqueantes) resguardan los umbrales de latencia, pero su ratificación numérica final permanece subordinada a la ejecución de pruebas de carga en el prototipo desplegado. |
| **QA-3** *(Escalabilidad: Picos de \+500%)* | **Parcialmente Abordado** | Se mantienen las tácticas de autoescalado elástico HPA, segregación a réplicas de lectura y buffers en el broker. Conforme al criterio metodológico del SEI, permanece parcial hasta constatar experimentalmente latencias $<3$ s bajo escenarios de estrés masivo. |
| **QA-7** *(Usabilidad nativa móvil e Internacionalización)* | **Not Yet Addressed** | Atributo centrado en la interacción de UI del cliente móvil nativo (sensores GPS, conversión de divisas localizadas y cálculo de rutas al hotel en $<2$ s). Queda diferido para el desarrollo específico de la capa de presentación/frontend fuera del scope del backend transaccional y analítico del prototipo. |

**Actualizacion del Design Backlog:**

| Categoría | Especificación | No abordado (Hasta el momento) | Parcialmente abordado | Abordado Completamente |
| :---- | :---- | :---- | :---- | :---- |
| Incumbencias Arquitectónicas | CNR 1 |  |   |  x |
|   | CNR 2 |  |   |  x |
|   | CNR 3 |  |   |  x |
|   | CNR 4 |  |   |  x |
|   | CNR 5 |  |   |  x |
| Restricciones | CON-1 |  |   |  x |
|   | CON-2 |  |   |  x |
| Atributos de Calidad | QA-1 |  |  x |   |
|   | QA-2 |  |   |  x |
|   | QA-3 |  |  x |   |
|   | QA-4 |  |   |  x |
|   | QA-5 |  |   |  x |
|   | QA-6 |  |   |  x |
|   | QA-7 | x |   |   |
| Requerimientos Funcionales Primarios | US-01 |  |   |  x |
|   | US-02 |  |   |  x |
|   | US-03 |  |  |  x |
|   | US-04 |  |   |  x |
|   | US-05 |  |   |  x |
|   | US-06 |  |   |  x |
|   | US-07 |  |   |  x |

### 