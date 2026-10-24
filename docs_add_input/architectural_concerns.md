# ADD Input Drivers: Architectural Concerns (CRN)

- **Proyecto:** Trotamundos — Plataforma Integral de Gestión de Viajes
- **Fase:** ADD 3.0 — Step 1: Review Inputs
- **Contexto del Sistema:** Greenfield System (Diseño desde cero)
- **Estado:** Baseline / En Backlog

---

## 1. Definición del Propósito de Diseño (Design Purpose)

Diseñar desde cero (*greenfield*) la arquitectura de software escalable y distribuida para la plataforma **Trotamundos**, capaz de soportar hasta 15 millones de usuarios con picos estacionales de demanda, garantizando interoperabilidad con proveedores turísticos externos, aislamiento de cargas analíticas y cumplimiento metodológico mediante *Spec-Driven Development* y registros de decisiones MADR.

---

## 2. Design Backlog: Architectural Concerns

| ID | Architectural Concern | Descripción Técnica | Impacto Arquitectónico y Estrategia | Estado |
| :--- | :--- | :--- | :--- | :--- |
| **CRN-1** | Establecimiento de la estructura macro inicial (*Greenfield base*) | Definir la organización estructural global, particionamiento de responsabilidades y estilos base para un desarrollo que parte desde cero. | Selección inicial de arquitectura de referencia y estilo arquitectónico (ej. microservicios / servicios distribuidos desacoplados) en la Iteración 1 de ADD. | `Not Yet Addressed` |
| **CRN-2** | Adopción de *Spec-Driven Development* y gobernanza de interfaces | Estandarización previa y estricta de contratos de interfaz (APIs) antes de iniciar la implementación y construcción de prototipos. | Establecimiento de especificaciones formales (OpenAPI / AsyncAPI), desacoplamiento entre frontend/backend mediante dobles de prueba (*mocks*) y verificación de cumplimiento de contratos con sistemas externos. | `Not Yet Addressed` |
| **CRN-3** | Segregación de procesamiento analítico y transaccional | Desacoplamiento entre la persistencia operativa de reservas en tiempo real y la ejecución de consultas analíticas pesadas para monetización. | Adopción de patrones de segregación de responsabilidades de consulta y comando (ej. réplicas de solo lectura / CQRS / pipelines de streaming asincrónicos) para evitar degradación de latencia en la experiencia del usuario. | `Not Yet Addressed` |
| **CRN-4** | Modularización para distribución de trabajo y gobernanza ADR | Organización de subsistemas y módulos alineados a la estructura del equipo, permitiendo desarrollo concurrente y trazabilidad formal de decisiones. | Definición de límites de módulos con dependencias explícitas (alineado a la Ley de Conway) y obligatoriedad de captura de decisiones mediante plantillas MADR (*Markdown Architectural Decision Records*). | `Not Yet Addressed` |
| **CRN-5** | Extensibilidad para integración de módulos avanzados (Entrega 2) | Provisión de puntos de extensión desacoplados para incorporar capacidades avanzadas (Sistema de recomendación personalizada o Asistente virtual / Chatbot) sin refactorizaciones del núcleo transaccional. | Diseño de interfaces de extensión y comunicación orientada a eventos para permitir el *plugging* independiente de componentes inteligentes futuros. | `Not Yet Addressed` |

---

## 3. Mapeo a Iteraciones Iniciales de ADD

| Driver ID | Iteración Prevista | Foco Principal de Decisión |
| :--- | :--- | :--- |
| **CRN-1** | Iteración 1 | Estructura general, arquitectura de referencia, conectores y particionamiento macro. |
| **CRN-2** | Iteración 1 / 2 | Definición de capas de integración, estándares OpenAPI y contratos de API Gateway. |
| **CRN-3** | Iteración 2 / 3 | Estrategia de bases de datos operativas vs. almacén analítico, mecanismos de sincronización asincrónica. |
| **CRN-4** | Transversal (1 a 4) | Estructuración modular en repositorios, asignación por componente y generación de ADRs. |
| **CRN-5** | Iteración 3 / 4 | Puntos de desacople e interfaces de integración para servicios de recomendación/asistente. |
