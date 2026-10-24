| ID | Architectural Concern | Descripción Técnica | Impacto Arquitectónico y Estrategia |
| :--- | :--- | :--- | :--- |
| :--- | :--- | :--- | :--- |
| **CRN-1** | Establecimiento de la estructura macro inicial (*Greenfield base*) | Definir la organización estructural global, particionamiento de responsabilidades y estilos base para un desarrollo que parte desde cero. | Selección inicial de arquitectura de referencia y estilo arquitectónico (ej. microservicios / servicios distribuidos desacoplados) en la Iteración 1 de ADD. 
| **CRN-2** | Adopción de *Spec-Driven Development* y gobernanza de interfaces | Estandarización previa y estricta de contratos de interfaz (APIs) antes de iniciar la implementación y construcción de prototipos. | 
| **CRN-3** | Segregación de procesamiento analítico y transaccional | Desacoplamiento entre la persistencia operativa de reservas en tiempo real y la ejecución de consultas analíticas pesadas para monetización. | 
| **CRN-4** | Modularización para distribución de trabajo y gobernanza ADR | Organización de subsistemas y módulos alineados a la estructura del equipo, permitiendo desarrollo concurrente y trazabilidad formal de decisiones. | 
| **CRN-5** | Extensibilidad para integración de módulos avanzados (Entrega 2) | Provisión de puntos de extensión desacoplados para incorporar capacidades avanzadas (Sistema de recomendación personalizada o Asistente virtual / Chatbot) sin refactorizaciones del núcleo transaccional. |
