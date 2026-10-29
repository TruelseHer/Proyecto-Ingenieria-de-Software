# Iteración 1 \- Método ADD Sistema "Trotamundos"

## **Paso 1: Review Inputs y Completado del Design Backlog**

| Categoria | Especificación | No abordado todavía | Parcialmente abordado | Abordado Completamente |
| :---- | :---- | :---- | :---- | :---- |
| Incumbencias Arquitectónicas | CNR 1 | x |   |   |
|   | CNR 2 | x |   |   |
|   | CNR 3 | x |   |   |
|   | CNR 4 | x |   |   |
|   | CNR 5 | x |   |   |
| Restricciones | CON-1 | x |   |   |
|   | CON-2 | x |   |   |
| Atributos de Calidad | QA-1 | x |   |   |
|   | QA-2 | x |   |   |
|   | QA-3 | x |   |   |
|   | QA-4 | x |   |   |
|   | QA-5 | x |   |   |
|   | QA-6 | x |   |   |
|   | QA-7 | x |   |   |
| Requerimientos Funcionales Primarios | US-01 | x |   |   |
|   | US-02 | x |   |   |
|   | US-03 | x |   |   |
|   | US-04 | x |   |   |
|   | US-05 | x |   |   |
|   | US-06 | x |   |   |
|   | US-07 | x |   |   |

## **Paso 2: Selección y justificación de Drivers Seleccionados**

#### **Objetivo de la Iteración**

* Definir la **estructura macro inicial del sistema** (*Greenfield base*), delimitando las fronteras entre subsistemas, los estilos arquitectónicos rectores y las capas de comunicación, garantizando el desacoplamiento temprano para clientes heterogéneos y servicios externos.

#### **Clasificación y Justificación de Drivers seleccionados**

* **CRN-1 (Establecimiento de la estructura macro inicial \- *Greenfield base*):**  
  * *Justificación:* Al ser un desarrollo desde cero, es imperativo establecer el particionamiento de responsabilidades de primer orden y los estilos arquitectónicos globales antes de profundizar en la lógica transaccional o componentes individuales.  
* **CON-1 (Soporte de clientes Web y Móvil nativo):**  
  * *Justificación:* Exige desacoplar completamente la lógica de negocio de las interfaces de usuario desde la capa perimetral (evitando la duplicación de código), lo que direcciona la arquitectura hacia patrones perimetrales como API Gateway y Backends for Frontends (BFF).  
* **CON-2 (Integración obligatoria con proveedores turísticos externos):**  
  * *Justificación:* El consumo de fuentes heterogéneas con contratos y tiempos variables demanda definir a nivel macro subsistemas de integración y capas de adaptación normalizadas (Anti-Corruption Layer) para no contaminar el dominio transaccional central.  
* **A3E-E1 (Escalabilidad \- Soporte de picos de \+500% ante eventos disruptivos):**  
  * *Justificación:* La división macro debe prever escalamiento horizontal independiente y desacoplamiento asíncrono para que eventos masivos de tráfico no colapsen la disponibilidad del sistema.  
* **CRN-3 (Segregación de procesamiento analítico y transaccional):**  
  * *Justificación:* Influye en no concebir un único almacén monolítico; desde el diseño inicial macro se debe prever la separación entre el subsistema operativo (OLTP) y el repositorio analítico (OLAP).

## **Paso 3: Elegir Uno o Más Elementos del Sistema a Refinar**

Al tratarse de la **Iteración 1 en un escenario Greenfield**, el elemento a descomponer y refinar es el **Sistema Completo ("Trotamundos")**

El refinamiento consistirá en:

1. Particionar el sistema en sus subsistemas macro principales (descomposición horizontal inicial).  
2. Definir los límites del sistema respecto a los actores externos (clientes web/móvil, proveedores turísticos, plataformas de correo y redes sociales)

## **Paso 4: Elegir Uno o Más Conceptos de Diseño que Satisfagan los Drivers**

En este paso se formalizan los conceptos arquitectónicos (patrones, estilos y tácticas) seleccionados para resolver el objetivo de la iteración (CRN-1), descartando aquellos que comprometen los requerimientos y traduciendo la estructura a un lenguaje técnico claro y directo para el equipo de desarrollo

#### **1\. Identificación y Selección de Conceptos de Diseño**

| Concepto de Diseño / Patrón | Tipo de Concepto | Drivers que Ataca | Justificación de Selección (Rationale) |
| :---- | :---- | :---- | :---- |
| **API Gateway (con Backend for Frontend \- BFF)** | Patrón Arquitectónico / Perimetral | CON-1, CRN-1 | Centraliza el punto de entrada, autentica peticiones y adapta las respuestas según el tipo de cliente. Permite enviar respuestas ligeras para la app móvil (cumpliendo FCP $<$ 1.4 s) y estructuradas para el navegador web ($\leq$ 800 ms) sin duplicar código. |
| **Microservicios Agrupados por Carga / Rol** | Estilo Arquitectónico Macro | CRN-1, QA-3  | Divide el sistema en subsistemas autónomos según su comportamiento operativo (operación en vivo, procesamiento en segundo plano y análisis de datos). Evita el riesgo de saturar el sistema ante picos repentinos de tráfico del \+500%. |
| **Desacoplamiento Asíncrono mediante Message Broker (Cola / Bus de Eventos)** | Patrón de Comunicación e Integración / Conector | CON-2, QA-3   | Desacopla temporalmente el procesamiento de ingesta (escaneo de correos e interfaces de agencias externas) del dominio transaccional de viajes.  |
| **Separación de Base de Datos Operativa vs. Base de Datos Analítica** | Patrón de Segregación de Datos | CRN-3, CRN-1 | Garantiza que las consultas pesadas y reportes de monetización se ejecuten sobre una base de datos histórica dedicada, impidiendo que bloqueen las operaciones del día a día (consultar o guardar viajes). |

#### 

### **Conceptos de Diseño Descartados**

#### **A. Alternativa descartada al API Gateway / BFF: *Cliente Directo a Servicios (Direct Client-to-Service Communication)***

* **Descripción:** Cada cliente (Web y Móvil) invoca directamente los endpoints de los distintos servicios de backend sin una capa intermedia.  
* **Motivo de descarte:** Viola la restricción de clientes heterogéneos y latencia diferenciada. Obligaría a la app móvil a realizar múltiples viajes de red (*chattiness*) degradando el First-Contentful Paint ($<1.4$ s), duplicaría la lógica de agregación y expondría contratos internos hacia internet sin un perímetro de seguridad centralizado.

#### **B. Alternativa descartada al desacoplamiento asíncrono: *Coreografía / Orquestación Sincrónica Request-Response (REST/SOAP puro)***

* **Descripción:** Los servicios se encadenan mediante llamadas HTTP bloqueantes secuenciales para sincronizar datos con aerolíneas, hoteles y correos.  
* **Motivo de descarte:** Tal como ejemplifica la cátedra ("se abandonó el request-response por acoplamiento y fallas en cascada"), esta opción genera fragilidad ante caídas de proveedores turísticos externos y agota los hilos del servidor ante picos de \+500%, destruyendo la disponibilidad.

#### **C. Alternativa descartada a la segregación OLTP / OLAP: *Base de Datos Relacional Tradicional Única (Traditional Relational Architecture)***

* **Descripción:** Centralizar tanto las transacciones de reservas (CRUD de usuarios) como las consultas pesadas de minería de datos y monetización en un único motor relacional.  
* **Motivo de descarte:** Las consultas pesadas de agregación para monetización bloquearían tablas e índices transaccionales críticos, degradando el tiempo de respuesta de la web ($<800$ ms).

#### **D. Alternativa descartada a la distribución de carga: *Rich Client tradicional vs. RIA (Rich Internet Application / Single Page Application)***

* **Descripción:** Descargar toda la lógica y agregación en clientes pesados de escritorio o clientes monolíticos server-rendered tradicionales.  
* **Motivo de descarte:** No obedece las mejores practicas actuales, ya que el modelo que se selecciono en su lugar, asegura que las aplicaciones móviles nativas y las aplicaciones web modernas consuman contratos livianos estandarizados bajo *Spec-Driven Development*.

## **Paso 5: Instanciar Elementos Arquitectónicos y Asignar Responsabilidades**

| Elemento Arquitectónico Instanciado | Tipo de Elemento | Responsabilidades Asignadas (Responsibility Allocation) |
| :---- | :---- | :---- |
| **API Gateway** | Componente Perimetral / Gateway | • Punto único de entrada seguro al sistema para el tráfico perimetral. • Terminación TLS/HTTPS y autenticación/validación preliminar de tokens JWT. • Enrutamiento y balanceo de peticiones entrantes hacia los adaptadores BFF correspondientes. |
| **Mobile BFF (Backend for Frontend)** | Módulo de Aplicación / Perimetral | • Exposición de contratos ligeros adaptados para la aplicación móvil nativa. • Agregación y poda de campos para optimizar el consumo de ancho de banda y cumplir la meta de First-Contentful Paint. • Transformación de datos adaptados al ciclo de vida móvil y caché perimetral liviana. |
| **Web BFF (Backend for Frontend)** | Módulo de Aplicación / Perimetral | • Exposición de contratos y composición de payloads para la interfaz web (Single Page Application). • Agregación de vistas completas de itinerarios para garantizar tiempos de respuesta rápidos. • Manejo de sesión y cabeceras de navegación de navegador. |
| **Servicio Transaccional de Viajes (Core Travel Service)** | Subsistema / Microservicio Transaccional | • Gestión centralizada de las operaciones de negocio sobre viajes, reservas e itinerarios del usuario. • Persistencia y actualización transaccional en la base de datos operativa. • Publicación de eventos de cambio de estado hacia el bus de mensajería (itinerario creado, actualizado, cancelado). |
| **Subsistema de Ingesta e Integración Externa (External Ingestion Service)** | Subsistema / Worker de Procesamiento | • Sondeo y escaneo continuo de bandejas de correo para detectar confirmaciones de viaje. • Conexión e integración con APIs de agencias externas (aerolíneas, hoteles, transportes). • Normalización de formatos externos heterogéneos y emisión asíncrona de eventos con las reservas detectadas. |
| **Message Broker (Event Bus / Colas)** | Conector / Intermediario Asíncrono (Referencia al estilo arquitectonico Publisher/Subscriber) | • Enrutamiento, retención y entrega asíncrona desacoplada de eventos y mensajes entre subsistemas. • Absorción de picos de carga repentinos (buffer de peticiones). • Soporte de reintentos automáticos y desacoplamiento temporal ante indisponibilidad de proveedores externos. |
| **Base de Datos Operativa (OLTP Store)** | Repositorio de Datos Transaccional | • Persistencia relacional/documental de alta concurrencia para el modelo transaccional de usuarios y viajes. • Garantía de integridad de datos y baja latencia de lectura/escritura en las consultas cotidianas del panel de viajes. |
| **Subsistema Analítico (Analytics Service)** | Subsistema / Microservicio de Procesamiento | • Consumo y procesamiento de eventos de viajes históricos y consolidados. • Ejecución de algoritmos de agregación, reportes de tendencias y analítica pesada para fines de monetización. |
| **Base de Datos Analítica (OLAP Store / Data Warehouse)** | Repositorio de Datos Analítico | • Almacenamiento histórico optimizado para lectura y procesamiento analítico masivo. • Aislamiento de cargas de trabajo de monetización para evitar la contención de recursos en el repositorio operativo. |

## **Paso 6: Vistas y ADR Registrados**

* De los ADR-01 al ADR-04 corresponden a la Iteración 1

[ADR-01](./adr/ADRS.md#https://github.com/TruelseHer/Proyecto-Ingenieria-de-Software/blob/main/adr/ADRS.md#adr-01-selecci%C3%B3n-del-estilo-macro-arquitect%C3%B3nico-basado-en-microservicios-agrupados-por-rolcarga)

* Vistas  
  * Vista 1 C\&C  
  * Vista 2 Módulos

## **Paso 7: Analizar el Diseño, Evaluar Resultados y Actualizar el Design Backlog**

### **1\. Evaluación del Objetivo de la Iteración y Análisis de Decisiones**

* **Objetivo de la Iteración:** Definir la estructura macro inicial del sistema (*Greenfield base*), delimitando las fronteras entre subsistemas, los estilos arquitectónicos rectores y las capas de comunicación, garantizando el desacoplamiento temprano para clientes heterogéneos y servicios externos.  
* **Evaluación General:** El objetivo de la iteración **ha sido alcanzado satisfactoriamente**. Se definió la macroestructura del sistema mediante la selección de estilos rectores (Microservicios agrupados por rol/carga operativa, API Gateway con BFF y Segregación Dual OLTP/OLAP), soportados formalmente en los ADR-01 a ADR-04 y representados en las vistas de Componentes y Conectores (C\&C) y de Descomposición de Módulos.

### **2\. Justificación del Cumplimiento de Drivers Foco Seleccionados**

| Driver | Estado al Finalizar It. 1 | Evidencia Arquitectónica y Justificación (Definition of Addressed) |
| :---- | :---- | :---- |
| **CRN-1** *(Estructura macro inicial \- Greenfield base)* | **Abordado Completamente** | Se resolvió el particionamiento de primer orden dividiendo el sistema en subsistemas autónomos (*Perímetro*, *Core Travel*, *Ingesta Externa*, *Analítica*). La relación estática quedó documentada en la Vista de Módulos y la dinámica en la Vista C\&C, formalizada en el **ADR-01**. Al tratarse de un concern de estructura global, queda cerrado en este nivel macro. |
| **CON-1** *(Soporte de clientes Web y Móvil nativo)* | **Abordado Completamente** | Se implementó el patrón arquitectónico Edge API Gateway complementado con adaptadores especializados *Mobile BFF* y *Web BFF* (**ADR-02**). La frontera física y lógica desacopla completamente las interfaces del backend transaccional, adaptando los payloads según el tipo de cliente. Las restricciones perimetrales quedan selladas. |
| **CON-2** *(Integración obligatoria con proveedores turísticos)* | **Parcialmente Abordado** | Se definió la frontera macro mediante el *Subsistema de Ingesta e Integración Externa* y su desacoplamiento asíncrono con el Message Broker (**ADR-03**). Sin embargo, el driver permanece **parcial** dado que resta especificar en la Iteración 2 los contratos de interfaz (OpenAPI), los mecanismos de anti-corrupción (ACL) y los adaptadores específicos para aerolíneas, hoteles y proveedores de correo. |
| **QA-3**  *(Escalabilidad: Soportar picos de \+500% ante eventos disruptivos)* | **Parcialmente Abordado** | Se incorporaron los pilares estructurales: división de servicios con escalamiento horizontal independiente (**ADR-01**) y uso del Message Broker como búfer elástico amortiguador (**ADR-03**). Permanece **parcial** según el marco de la cátedra porque la satisfacción total de la medida de respuesta requiere definir tácticas de asignación de réplicas, políticas de elasticidad e infraestructura que se abordarán en la **Iteración 3**. |
| **CRN-3** *(Segregación de procesamiento analítico y transaccional)* | **Abordado Completamente** | Quedó resuelto estructuralmente al separar de forma tajante el repositorio transaccional (*OLTP Store*) del almacén analítico (*OLAP Data Warehouse*) y delimitar el *Analytics Service* (**ADR-04**). Las operaciones masivas de monetización no compiten ni bloquean las tablas de los viajes de los usuarios. |

**Actualizacion del Design Backlog**

| Categoría | Especificación | No abordado (Hasta el momento) | Parcialmente abordado | Abordado Completamente |
| :---- | :---- | :---- | :---- | :---- |
| Incumbencias Arquitectónicas | CNR 1 |  |   |  x |
|   | CNR 2 | x |   |   |
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
| Requerimientos Funcionales Primarios | US-01 | x |   |   |
|   | US-02 | x |   |   |
|   | US-03 | x |   |   |
|   | US-04 | x |   |   |
|   | US-05 | x |   |   |
|   | US-06 | x |   |   |
|   | US-07 | x |   |   |

### 
