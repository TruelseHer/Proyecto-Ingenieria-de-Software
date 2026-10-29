# Iteración 2 \- Método ADD Sistema "Trotamundos"

## **Paso 2: Selección y justificación de Drivers Seleccionados**

#### **Objetivo de la Iteración**

* Diseñar las estructuras arquitectónicas internas y los contratos de interfaz que dan soporte a la **funcionalidad primaria del sistema** (*Primary Functionality*), definiendo el modelo de dominio (*Domain Model*), la descomposición interna en capas de los servicios principales y la gobernanza de contratos bajo el enfoque *Spec-Driven Development*.

#### **Clasificación y Justificación de Drivers Seleccionados**

* **US-04 (Gestión Manual de Reservas):**  
  * *Justificación:* Es un requerimiento funcional central que habilita al usuario a crear, modificar y cancelar reservas de forma directa desde las interfaces. Demanda el diseño del modelo de dominio transaccional interno, la definición de contratos REST para operaciones CRUD y la persistencia en la base de datos operativa sin depender de procesos automáticos.  
* **US-05 (Ciclo de Vida del Itinerario y Viajes):**  
  * *Justificación:* Define la lógica medular de agregación de la plataforma: agrupar reservas bajo la entidad unificadora "Viaje" y gestionar sus transiciones de estado (activo, completado, archivado). Es indispensable para diseñar la estructura de entidades del dominio, las invariantes de negocio y los endpoints de consulta del itinerario consolidados que alimentan el panel principal.  
* **US-01 (Procesamiento y Escaneo Automático de Correos):**  
  * *Justificación:* Constituye la fuente primaria de automatización del sistema para la detección de reservas. Su inclusión en esta etapa es necesaria para modelar internamente los componentes de extracción, parsing y normalización dentro del *Subsistema de Ingesta*, así como el esquema del evento que este componente debe publicar hacia el broker de mensajería.  
* **US-02 (Filtros y Whitelists de Remitentes Seguros):**  
  * *Justificación:* Complementa directamente el procesamiento de correos, estableciendo las reglas de descarte temprano y validación de seguridad de remitentes permitidos para evitar el procesamiento innecesario de mensajes no vinculados a viajes. Requiere definir entidades y repositorios de configuración de usuario dentro de la lógica del worker de ingesta.  
* **US-03 (Integración con Proveedores Externos y Agencias):**  
  * *Justificación:* Aborda la conexión con plataformas de viajes externas (aerolíneas, hoteles, arrendadoras) para la obtención y actualización de reservas. Resulta fundamental para definir el patrón de adaptación (Anti-Corruption Layer \- ACL) y la especificación del contrato común de reservas que aísla al dominio interno de las variaciones y contratos heterogéneos de terceros.  
* **US-06 (Resolución Rápida de Problemas con Agencia Preferida):**  
  * *Justificación:* Especifica el mecanismo funcional mediante el cual el usuario puede canalizar reclamos o contingencias sobre una reserva hacia su agencia de viajes designada. Requiere modelar la asociación entre el perfil del usuario, su agencia preferida y los puntos de contacto en el modelo de dominio.  
* **CRN-2 (Adopción de Spec-Driven Development y Gobernanza de Contratos):**  
  * *Justificación:* La consigna metodológica establece definir y formalizar contractualmente las interfaces antes de avanzar con la implementación del prototipo. En esta iteración, actúa como driver rector para exigir la especificación formal de contratos (OpenAPI para APIs síncronas de backend y esquemas JSON/AsyncAPI para eventos del message broker), sirviendo además como criterio delimitador para la asignación de tareas al equipo de desarrollo.

## **Paso 3: Elegir Uno o Más Elementos del Sistema a Refinar**

De acuerdo con el objetivo de la iteración (Primary Functionality) y los drivers seleccionados (US-01 a US-06, CRN-2), los elementos del sistema a refinar corresponden a los dos subsistemas de backend que concentran la lógica de negocio y la ingesta de datos:

1. **trotamundos.core\_travel (Servicio Transaccional de Viajes):**  
   - *Tipo de refinamiento:* Descomposición interna (vertical).  
   - *Propósito:* Descomponer el subsistema en capas lógicas de implementación (controladores de API, servicios de dominio, entidades y acceso a datos) para dar soporte al modelo de dominio de viajes e itinerarios (US-04, US-05, US-06) bajo contratos formales OpenAPI (CRN-2).  
2. **trotamundos.ingestion (Subsistema de Ingesta e Integración Externa):**  
   - *Tipo de refinamiento:* Descomposición interna (vertical).  
   - *Propósito:* Descomponer el worker de procesamiento en módulos de sondeo y escaneo de correos (US-01), componentes de filtrado y listas de remitentes seguros (US-02), y adaptadores de integración (Anti-Corruption Layer) hacia proveedores turísticos externos (US-03), definiendo el contrato del evento publicado hacia el Message Broker (CRN-2).

## **Paso 4: Elegir Uno o Más Conceptos de Diseño que Satisfagan los Drivers**

En este paso se formalizan los conceptos arquitectónicos (patrones de diseño de software, arquitecturas de referencia internas y patrones de integración) para dar soporte a la funcionalidad primaria del sistema y estructurar internamente los dos subsistemas refinados (trotamundos.core\_travel y trotamundos.ingestion).

#### **Identificación y Selección de Conceptos de Diseño**

| Concepto de Diseño / Patrón | Tipo de Concepto | Drivers que Ataca | Justificación de Selección (Rationale) |
| :---- | :---- | :---- | :---- |
| **Modelo de Dominio Enriquecido (*Domain Model Pattern*)** | Patrón de Diseño / Lógica de Negocio | US-04, US-05, US-06 | Modela las entidades centrales del negocio (Viaje, Reserva, AgenciaPreferida) encapsulando su comportamiento e invariantes. Resuelve las transiciones del ciclo de vida del viaje (*Activo* $\rightarrow$ *Completado* $\rightarrow$ *Archivado*) y la agregación de múltiples reservas en una única vista coherente. |
| **Arquitectura de Referencia de Aplicación de Servicio (*Service Application Pattern / N-Layered*)** | Arquitectura de Referencia Interna | CRN-2, US-04, US-05 | Descompone internamente el microservicio trotamundos.core\_travel en tres capas desacopladas: **Controladores de API** (exposición de endpoints), **Servicios de Aplicación / Dominio** (lógica de negocio) y **Repositorios** (acceso a datos). Permite implementar servicios sin interfaz gráfica y establecer unidades de implementación claras para los equipos de desarrollo. |
| **Capa Anticorrupción (*Anti-Corruption Layer \- ACL*)** | Patrón de Integración / Adaptación | US-01, US-03 | Aísla el dominio interno de trotamundos de los modelos heterogéneos y propietarios provenientes de APIs de aerolíneas, cadenas hoteleras y formatos de correo electrónico. Traduce estructuras dispares a un modelo canónico de evento (ReservationDetectedEvent) antes de publicarlo al broker de mensajería. |
| **Gobernanza de Interfaces Contract-First (*Spec-Driven Development*)** | Patrón Metodológico / Diseño de Interfaces | CRN-2, US-04, US-05 | Exige definir, versionar y validar formalmente los contratos antes de iniciar la codificación: especificaciones **OpenAPI 3.0** para los endpoints REST del Core Travel y esquemas **AsyncAPI / JSON Schema** para los eventos del Message Broker. Garantiza la trazabilidad del prototipo y desacopla el desarrollo en paralelo de los equipos. |
| **Externalización de la Configuración (*Externalize Configuration*)** | Táctica de Mantenibilidad / Despliegue | US-01, US-02, US-03 | Separa las credenciales, URLs de APIs externas y parámetros de sondeo de correo del código fuente de los microservicios, inyectándolos mediante variables de entorno o archivos de configuración según el ambiente operativo. |

#### 

#### **Conceptos de Diseño Descartados**

* **Modelo de Dominio Anémico (*Anemic Domain Model*) / *Transaction Script*:**  
  * *Motivo de descarte:* Tratar a las entidades como simples bolsas de datos (*getters/setters*) y dispersar la lógica de transición de viajes y validación de reservas en los controladores o scripts procedurales deriva en una arquitectura *ad-hoc*, con código duplicado y de muy difícil mantenimiento a medida que se incorporan nuevas reglas.  
* **Integración Directa sin Capa de Traducción (*Shared Kernel* / Modelo Compartido con Terceros):**  
  * *Motivo de descarte:* Adoptar directamente las estructuras de datos de las APIs externas de turismo o los esquemas crudos de los correos dentro del modelo de datos interno de Trotamundos generaría un acoplamiento severo. Cualquier cambio de contrato en un tercero propagaría modificaciones directas sobre el código transaccional de viajes.  
* **Enfoque *Code-First* para la Definición de APIs:**  
  * *Motivo de descarte:* Generar la documentación o contratos a partir de la implementación del código viola frontalmente el driver CRN-2 y la consigna rectora del trabajo práctico. Impide que los desarrolladores de frontend (BFFs) y backend trabajen simultáneamente y elimina la gobernanza previa de interfaces.

## **Paso 5: Instanciar Elementos Arquitectónicos y Asignar Responsabilidades**

**Catálogo de Elementos Instanciados y Asignación de Responsabilidades**

| Subsistema / Ámbito | Elemento Instanciado | Tipo de Elemento | Responsabilidades Asignadas (Responsibility Allocation) |
| :---- | :---- | :---- | :---- |
| **trotamundos.core\_travel** | **TripController / ReservationController** | Controlador de API / Capa Web | • Exponer los endpoints REST síncronos del servicio central de viajes. • Validar sintácticamente los payloads de entrada conforme al contrato OpenAPI. • Atender las peticiones de creación, modificación y cancelación manual de reservas (US-04) y consulta de itinerarios consolidados (US-05) provenientes de los BFFs. |
| **trotamundos.core\_travel** | **TripService / ItineraryManager** | Servicio de Aplicación / Dominio | • Coordinar la lógica transaccional de negocio sobre viajes e itinerarios. • Ejecutar las reglas del ciclo de vida de los viajes (agrupación temporal, transiciones de estado y archivado automático de viajes finalizados \- US-05). • Gestionar la vinculación entre reservas y la agencia de viajes preferida del usuario para la resolución de contingencias (US-06). |
| **trotamundos.core\_travel** | **Entidades de Dominio (Trip, Reservation, subtipos)** | Clases de Dominio (*Domain Model*) | • Trip (*Aggregate Root*): Encapsula la colección de reservas y el estado del itinerario (ACTIVE, COMPLETED, ARCHIVED). • Reservation (Clase base / Abstracción): Modela atributos comunes (código de confirmación, fechas, estado y agencia preferida). • Subtipos polimórficos (FlightReservation, HotelReservation, VehicleRentalReservation): Proveen extensibilidad para futuros transportes y alojamientos. |
| **trotamundos.core\_travel** | **TripRepository / ReservationRepository** | Capa de Acceso a Datos / Repositorio | • Abstracción de persistencia transaccional (vía ORM / Spring Data) sobre el almacén relacional. • Mapeo objeto-relacional de las entidades de dominio y ejecución de consultas de lectura optimizadas para el itinerario del usuario. |
| **trotamundos.core\_travel** | **ReservationEventConsumer** | Consumidor Asíncrono / Listener | • Escuchar el tópico/cola de eventos del Message Broker. • Deserializar el evento ReservationDetectedEvent y delegar a TripService para la incorporación automática de la reserva al viaje correspondiente (US-01, US-03). |
| **trotamundos.core\_travel** | **Base de Datos Operativa (PostgreSQL / RDBMS)** | Almacén de Datos Transaccional (OLTP) | • Garantizar persistencia con propiedades ACID para el registro de usuarios, viajes y reservas • Mantener índices sobre claves de usuario y fechas para soportar respuestas eficientes en el panel de viajes. |
| **trotamundos.ingestion** | **EmailPollingWorker** | Tarea en Segundo Plano (*Scheduled Worker*) | • Ejecutar tareas periódicas de sondeo sobre las bandejas de entrada de correo de los usuarios mediante protocolos estándar (IMAP / APIs de correo con OAuth) (US-01). • Operar de forma autónoma según los intervalos definidos en la configuración externalizada. |
| **trotamundos.ingestion** | **SenderWhitelistFilter** | Componente de Validación / Filtro | • Comprobar remitentes y dominios frente a las listas de remitentes seguros (*whitelists*) configuradas por el viajero (US-02). • Descartar mensajes no autorizados o irrelevantes antes de procesar su contenido. |
| **trotamundos.ingestion** | **EmailParser** | Extractor Sintáctico / Procesador | • Analizar la estructura sintáctica de correos electrónicos aprobados. • Extraer códigos de reserva, fechas, nombres de pasajeros y tramos de transporte u hospedaje (US-01). |
| **trotamundos.ingestion** | **ProviderAdapters (ACL Clients)** | Adaptadores de Integración / ACL | • Conectores especializados por proveedor (AirlineAdapter, HotelAdapter, RentalAdapter). • Consumir APIs externas de terceros y absorber las particularidades de sus formatos heterogéneos (US-03). |
| **trotamundos.ingestion** | **CanonicalDataTranslator** | Mapeador / Traductor ACL | • Transformar los modelos heterogéneos de terceros y datos extraídos de correos a la estructura del evento canónico interno de Trotamundos, aislando el dominio de variaciones externasl. |
| **trotamundos.ingestion** | **ReservationEventPublisher** | Productor de Mensajería | • Publicar de forma no bloqueante el evento normalizado ReservationDetectedEvent hacia el Message Broker para su procesamiento desacoplado. |

#### 

Definición Formal de Interfaces y Gobernanza Contractual (*Spec-Driven Development* \- CRN-2)

En cumplimiento de la consigna contractual, se formalizan las especificaciones de interfaz que rigen la interacción entre módulos previo a su implementación:

1. Contrato REST Síncrono: TravelCoreAPI.yaml (Especificación OpenAPI 3.0)  
2. Contrato Asíncrono de Eventos: ReservationDetectedEvent (JSON Schema / AsyncAPI)  
3. Configuración Externalizada: EnvironmentConfigProfile

[Spec-Driven Development](?tab=t.xvp9bf6elzv9)

## **Paso 6: Vistas y ADR Registrados**

* De los ADR-05 al ADR-09 corresponden a la iteración 2

[ADR-05](../adr/ADRS.md#adr-05-adopción-del-patrón-modelo-de-dominio-enriquecido-rich-domain-model)
[ADR-06](../adr/ADRS.md#adr-06-estructuración-interna-del-servicio-de-viajes-mediante-service-application-pattern-arquitectura-en-capas)
[ADR-07](../adr/ADRS.md#adr-07-aislamiento-de-fuentes-externas-mediante-capa-anticorrupción-anti-corruption-layer---acl)
[ADR-08](../adr/ADRS.md#adr-08-adopción-del-enfoque-contract-first-mediante-openapi-y-asyncapi-spec-driven-development)
[ADR-09](../adr/ADRS.md#adr-09-selección-de-base-de-datos-relacional-postgresql-para-el-modelo-transaccional-de-viajes)

* Vistas  
  * Vista 1 Clases Core\_travel  
  * Vista 2 Modulos Core\_travel-Ingestion  
  * Vista 3 Secuencia

## **Paso 7: Analizar el Diseño, Evaluar Resultados y Actualizar el Design Backlog**

### **1\. Evaluación del Objetivo de la Iteración y Revisión del Diseño**

* **Objetivo de la Iteración:** Diseñar las estructuras arquitectónicas internas y los contratos de interfaz que dan soporte a la funcionalidad primaria del sistema (*Primary Functionality*), definiendo el modelo de dominio (*Domain Model*), la descomposición interna en capas de los servicios principales y la gobernanza de contratos bajo el enfoque *Spec-Driven Development*.  
* **Evaluación General:** El objetivo de la iteración ha sido **alcanzado satisfactoriamente**. Se descompusieron verticalmente los subsistemas centrales (trotamundos.core\_travel y trotamundos.ingestion), se formalizó el modelo de dominio enriquecido (*Rich Domain Model* con invariantes y polimorfismo para reservas), se aislaron las fuentes externas mediante una Capa Anticorrupción (ACL) y se congelaron los contratos formales síncronos (OpenAPI 3.0) y asíncronos (AsyncAPI / JSON Schema) previo a la programación. Todas las decisiones quedaron respaldadas formalmente en los registros ADR-05 a ADR-09 y documentadas en sus respectivas vistas estáticas y dinámicas.

### **2\. Justificación del Cumplimiento de Drivers y Definición de Abordaje (*Definition of Addressed*)**

| Driver | Estado al Finalizar It. 2 | Evidencia Arquitectónica y Justificación (Definition of Addressed) |
| :---- | :---- | :---- |
| **US-04** *(Gestión Manual de Reservas)* | **Abordado Completamente** | Se diseñó la descomposición en 3 capas del servicio (Controllers $\rightarrow$ Services $\rightarrow$ Repositories), los contratos síncronos OpenAPI (POST, PUT, DELETE /reservations) y la persistencia relacional con PostgreSQL respaldada en el ADR-06 y ADR-09. La vista dinámica síncrona valida el flujo transaccional completo. |
| **US-05** *(Ciclo de Vida del Itinerario y Viajes)* | **Abordado Completamente** | Quedó resuelto mediante la entidad raíz Trip (*Aggregate Root*) que encapsula las transiciones de estado (ACTIVE, COMPLETED, ARCHIVED), invariantes de fechas y cálculo del itinerario unificado (ADR-05). Se especificó el endpoint GET /api/v1/trips/active en el contrato TravelCoreAPI.yaml. |
| **US-06** *(Resolución Rápida con Agencia Preferida)* | **Abordado Completamente** | Se modeló la asociación entre User, TravelAgency y Reservation en el diagrama de dominio. Quedó especificado el endpoint contractual POST /reservations/{id}/agency-support y la delegación en la capa de negocio para canalizar reclamos a la agencia asignada. |
| **CRN-2** *(Spec-Driven Development)* | **Abordado Completamente** *(Solo para el Core funcional e Ingesta)* | Se formalizaron y versionaron los contratos TravelCoreAPI.yaml (OpenAPI 3.0) y el esquema canónico ReservationDetectedEvent (JSON Schema / AsyncAPI) con anterioridad a la etapa de desarrollo (ADR-08), delimitando las unidades de implementación para los equipos de trabajo. |
| **US-02** *(Filtros y Whitelists de Remitentes)* | **Abordado Completamente** | Se instanció el componente determinístico SenderWhitelistFilter dentro de trotamundos.ingestion, interceptando correos antes del parsing y validando contra la lista de remitentes seguros del usuario, tal como se refleja en la vista de módulos y de secuencia asíncrona. |
| **US-01** *(Escaneo Automático de Correos)* | **Parcialmente Abordado** | Se instanciaron los componentes EmailPollingWorker, EmailParser, y el publicador de eventos al broker. Permanece **parcial** dado que la extracción y adaptabilidad a plantillas de correo heterogéneas (heurísticas de parseo e integración concreta con OAuth/IMAP de proveedores) requiere validación empírica en el prototipo ejecutable. |
| **US-03** *(Integración con Proveedores Externos)* | **Parcialmente Abordado** | Se definió la estructura de la Capa Anticorrupción (ProviderAdapters, CanonicalDataTranslator) formalizada en el ADR-07. Permanece **parcial** (manteniendo en el mismo estado a la restricción CON-2) porque resta implementar y verificar los contratos particulares y adaptadores específicos para cada aerolínea, hotel o arrendadora en el código |

**Actualizacion del Design Backlog**

| Categoría | Especificación | No abordado (Hasta el momento) | Parcialmente abordado | Abordado Completamente |
| :---- | :---- | :---- | :---- | :---- |
| Incumbencias Arquitectónicas | CNR 1 |  |   |  x |
|   | CNR 2 |  |  x |   |
|   | CNR 3 |  |   |  x |
|   | CNR 4 | x |   |   |
|   | CNR 5 | x |   |   |
| Restricciones | CON-1 |  |   |  x |
|   | CON-2 |  |  x |   |
| Atributos de Calidad | QA-1 | x |   |   |
|   | QA-2 | x |   |   |
|   | QA-3 |  |  x |   |
|   | QA-4 | x |   |   |
|   | QA-5 | x |   |   |
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
