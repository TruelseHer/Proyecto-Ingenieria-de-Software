# Indice General - Sistema "Trotamundos"

1.  Inputs:
    * [arbol de utilidad](input/arbol-de-utilidad.md)
    * [atributos de calidad](input/atributos-de-calidad.md)
    * [restricciones](input/restricciones.md)
    * [proposito de diseño](input/proposito-de-dise%C3%B1o.md)
    * [requerimientos funcionales primarios](input/requerimientos-funcionales-primarios.md)
    * [incumbencias arquitectonicas](input/incumbencias-arquitectonicas.md)

2. Iteraciones
    * [Iteracion 1](iteraciones/Iteracion_1.md)
    * [Iteracion 2](iteraciones/Iteracion_2.md)
    * [Iteracion 3](iteraciones/Iteracion_3.md)
    * [Iteracion 4](iteraciones/Iteracion_4.md)
    * [Iteracion 5](iteraciones/Iteracion_5.md)

3. Specs
    * [Iteracion 2](spec/Spec-Driven_Devolpement.md#correspondiente-a-iteración-2)
        * [A. Contrato REST Síncrono: TravelCoreAPI.yaml](spec/Spec-Driven_Devolpement.md#a-contrato-rest-síncrono-travelcoreapiyaml-especificación-openapi-30)
        * [B. Contrato Asíncrono de Eventos: ReservationDetectedEvent](spec/Spec-Driven_Devolpement.md#b-contrato-asíncrono-de-eventos-reservationdetectedevent-json-schema--asyncapi)
        * [C. Configuración Externalizada: EnvironmentConfigProfile](spec/Spec-Driven_Devolpement.md#c-configuración-externalizada-environmentconfigprofile)
    * [Iteracion 4](spec/Spec-Driven_Devolpement.md#correspondiente-a-iteración-4)
        * [A. Contrato Asíncrono de Ciclo de Vida: TripLifecycleEvent](spec/Spec-Driven_Devolpement.md#a-contrato-asíncrono-de-ciclo-de-vida-triplifecycleevent-json-schema--asyncapi)
        * [B. Contrato OpenAPI del Servicio Analítico de Monetización: AnalyticsAPI.yaml](spec/Spec-Driven_Devolpement.md#b-contrato-openapi-del-servicio-analítico-de-monetización-analyticsapiyaml)
        * [C. Interfaz de Software para Plugins de Recomendación](spec/Spec-Driven_Devolpement.md#c-interfaz-de-software-para-plugins-de-recomendación-recommendationstrategy)
        * [D. Especificación Táctica del Circuit Breaker en Ingesta](spec/Spec-Driven_Devolpement.md#d-especificación-táctica-del-circuit-breaker-en-ingesta-configuración-de-resiliencia)
    * [Iteracion 5](spec/Spec-Driven_Devolpement.md#correspondiente-a-iteración-5)
        * [A. Contrato REST: MobileNavigationAPI.yaml](spec/Spec-Driven_Devolpement.md#a-contrato-rest-mobilenavigationapiyaml-openapi-30)
        * [B. Interfaz de Software Nativa: ILocationService](spec/Spec-Driven_Devolpement.md#b-interfaz-de-software-nativa-ilocationservice-abstracción-de-hardware-móvil)
        * [C. Interfaz de Cálculo de Ruta: IRouteNavigationEngine](spec/Spec-Driven_Devolpement.md#c-interfaz-de-cálculo-de-ruta-iroutenavigationengine)

4. Vistas
    * Iteracion 1
        * [Vista de C&C](vistas/Iteracion%201%20Vista%201%20C&C.png)
        * [Vista de Modulos General](vistas/Iteracion%201%20Vista%202%20Modulos.png)
    * Iteracion 2
        * [Vista de Clases Core_Travel](vistas/Iteracion%202%20Vista%201%20Clases%20Core_travel.png)
        * [Vista de Modulos Core_travel e Ingesta](vistas/Iteracion%202%20Vista%202%20Modulos%20Core_travel%20-%20Ingestion.png)
        * [Vista de Secuencia Creacion Manual de una reserva](vistas/Iteracion%202%20Vista%203%20Secuencia.png)
    * Iteracion 3
        * [Vista de Deployment](vistas/Iteracion%203%20Vista%201%20Deployment.png)
        * [Vista de C&C Incluyendo bases de datos](vistas/Iteracion%203%20Vista%202%20C&C.png)
    * Iteracion 4
        * [Vista de C&C con Recomendador e Ingesta](vistas/Iteracion%204%20Vista%201%20C&C.png)
        * [Vista de Modulos Core_travel y Recomendador](vistas/Iteracion%204%20Vista%202%20Modulos.png)
        * [Vista de Secuencia de Ingesta ](vistas/vista3)
    * Iteracion 5
        * [Vista de Modulos del Cliente App Movil](vistas/Iteracion%205%20Vista%201%20Modulos.png)
        * [Vista de C&C Interaccion Movil](vistas/Iteracion%205%20Vista%202%20C&C.png)
        * [Vista de Secuencia de Navegacion e Interaccion](vistas/Iteracion%205%20Vista%203%20Secuencia.png)

5. ADRs
    * [Iteracion 1](adr/ADRS.md#adr-01-selección-del-estilo-macro-arquitectónico-basado-en-microservicios-agrupados-por-rolcarga)
        * [ADR-01: Selección del estilo macro arquitectónico basado en Microservicios Agrupados por Rol/Carga](adr/ADRS.md#adr-01-selección-del-estilo-macro-arquitectónico-basado-en-microservicios-agrupados-por-rolcarga)
        * [ADR-02: Adopción del patrón Edge API Gateway con Backend for Frontend (BFF) diferenciado](adr/ADRS.md#adr-02-adopción-del-patrón-edge-api-gateway-con-backend-for-frontend-bff-diferenciado)
        * [ADR-03: Desacoplamiento de la Ingesta Externa mediante Conector Message Broker (Asynchronous Messaging)](adr/ADRS.md#adr-03-desacoplamiento-de-la-ingesta-externa-mediante-conector-message-broker-asynchronous-messaging)
        * [ADR-04: Segregación física de Almacenes de Datos Operativo (OLTP) y Analítico (OLAP)](adr/ADRS.md#adr-04-segregación-física-de-almacenes-de-datos-operativo-oltp-y-analítico-olap)
    * [Iteracion 2](adr/ADRS.md#adr-05-adopción-del-patrón-modelo-de-dominio-enriquecido-rich-domain-model)
        * [ADR-05: Adopción del Patrón Modelo de Dominio Enriquecido (*Rich Domain Model*)](adr/ADRS.md#adr-05-adopción-del-patrón-modelo-de-dominio-enriquecido-rich-domain-model)
        * [ADR-06: Estructuración Interna del Servicio de Viajes mediante *Service Application Pattern* (Arquitectura en Capas)](adr/ADRS.md#adr-06-estructuración-interna-del-servicio-de-viajes-mediante-service-application-pattern-arquitectura-en-capas)
        * [ADR-07: Aislamiento de Fuentes Externas mediante Capa Anticorrupción (*Anti-Corruption Layer \- ACL*)](adr/ADRS.md#adr-07-aislamiento-de-fuentes-externas-mediante-capa-anticorrupción-anti-corruption-layer---acl)
        * [ADR-08: Adopción del Enfoque *Contract-First* mediante OpenAPI y AsyncAPI (*Spec-Driven Development*)](adr/ADRS.md#adr-08-adopción-del-enfoque-contract-first-mediante-openapi-y-asyncapi-spec-driven-development)
        * [ADR-09: Selección de Base de Datos Relacional (PostgreSQL) para el Modelo Transaccional de Viajes](adr/ADRS.md#adr-09-selección-de-base-de-datos-relacional-postgresql-para-el-modelo-transaccional-de-viajes)
    * [Iteracion 3](adr/ADRS.md#adr-10-topología-de-base-de-datos-primario-réplica-con-failover-multi-az-y-réplicas-de-lectura)
        * [ADR-10: Topología de Base de Datos Primario-Réplica con Failover Multi-AZ y Réplicas de Lectura](adr/ADRS.md#adr-10-topología-de-base-de-datos-primario-réplica-con-failover-multi-az-y-réplicas-de-lectura)
        * [ADR-11: Priorización de Seguridad y Validación Criptográfica en Memoria en el API Gateway](adr/ADRS.md#adr-11-priorización-de-seguridad-y-validación-criptográfica-en-memoria-en-el-api-gateway)
        * [ADR-12: Autoescalado Horizontal Elástico (HPA) y Amortiguación de Ingesta mediante Message Broker](adr/ADRS.md#adr-12-autoescalado-horizontal-elástico-hpa-y-amortiguación-de-ingesta-mediante-message-broker)
    * [Iteracion 4](adr/ADRS.md#adr-13-ingesta-analítica-asíncrona-guiada-por-eventos-para-consultas-de-monetización)
        * [ADR-13: Ingesta Analítica Asíncrona Guiada por Eventos para Consultas de Monetización](adr/ADRS.md#adr-13-ingesta-analítica-asíncrona-guiada-por-eventos-para-consultas-de-monetización)
        * [ADR-14: Anonimización y Desidentificación de Datos en Tránsito para el Repositorio Analítico](adr/ADRS.md#adr-14-anonimización-y-desidentificación-de-datos-en-tránsito-para-el-repositorio-analítico)
        * [ADR-15: Desacoplamiento del Motor de Recomendaciones mediante Arquitectura de Plugins y Puntos de Extensión](adr/ADRS.md#adr-15-desacoplamiento-del-motor-de-recomendaciones-mediante-arquitectura-de-plugins-y-puntos-de-extensión)
        * [ADR-16: Resiliencia de Ingesta Externa mediante Circuit Breaker y Dead-Letter Queue (DLQ)](adr/ADRS.md#adr-16-resiliencia-de-ingesta-externa-mediante-circuit-breaker-y-dead-letter-queue-dlq)
        * [ADR-17: Extensibilidad de Transportes mediante Modelo Abierto/Cerrado y Contratos Modulares](adr/ADRS.md#adr-17-extensibilidad-de-transportes-mediante-modelo-abiertocerrado-y-contratos-modulares)
    * [Iteracion 5](adr/ADRS.md#adr-18-adopción-del-patrón-mvvm-y-abstracción-de-hardware-de-localización-para-la-aplicación-móvil)
        * [ADR-18: Adopción del Patrón MVVM y Abstracción de Hardware de Localización para la Aplicación Móvil](adr/ADRS.md#adr-18-adopción-del-patrón-mvvm-y-abstracción-de-hardware-de-localización-para-la-aplicación-móvil)
        * [ADR-19: Localización e Internacionalización (i18n) Híbrida en Memoria con Agregación Contextual en Mobile BFF](adr/ADRS.md#adr-19-localización-e-internacionalización-i18n-híbrida-en-memoria-con-agregación-contextual-en-mobile-bff)