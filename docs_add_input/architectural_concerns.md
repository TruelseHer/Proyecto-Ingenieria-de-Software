# Architectural Concerns (CRN) - Proyecto Trotamundos

Los **Architectural Concerns** representan inquietudes, necesidades técnicas, organizacionales y de ingeniería que guían el diseño de la arquitectura desde el inicio del método ADD, sin provenir necesariamente de una función directa de usuario.

---

## Tabla Resumen de Drivers (Concerns)

| ID | Concern (Inquietud Arquitectónica) | Justificación  |
| :--- | :--- | :--- |
| **CRN-1** | Establecer la estructura macro del sistema (*Greenfield*). | Inicio del diseño de la plataforma "Trotamundos" desde cero. |
| **CRN-2** | Habilitar diseño por contratos (*Spec-driven development*). | Exigencia metodológica del TPE para trazabilidad y APIs. |
| **CRN-3** | Separar las consultas pesadas de números/estadísticas del uso diario de la app. | Consultas pesadas de monetización sin degradar rendimiento. |
| **CRN-4** | Modularizar para asignación de trabajo y documentación ADR. | Coordinación grupal y captura formal con MADR. |
| **CRN-5** | Extensibilidad para módulos avanzados (Recomendación / Chatbot). | Alcance obligatorio especificado en el enunciado del tpe. |

---

## Descripción Detallada de cada Concern

### CRN-1: Establecer la estructura macro del sistema (Greenfield)
* **¿Qué es?**  
  Como el proyecto arranca desde cero (*greenfield*), la primera preocupación técnica es definir el esqueleto general de la aplicación antes de programar nada.
* **Impacto en el diseño:**  
  Obliga a elegir en la primera iteración el estilo arquitectónico base (por ejemplo, microservicios frente a un monolito modular) para que el equipo sepa cómo se organizan los componentes, los servicios y la base de datos central.

---

### CRN-2: Habilitar diseño por contratos (Spec-Driven Development)
* **¿Qué es?**  
  Consiste en escribir y dejar fijas las especificaciones de las APIs (por ejemplo con OpenAPI / Swagger) antes de tirar una sola línea de código real.
* **Impacto en el diseño:**  
  Permite que el equipo frontend pueda trabajar armando pantallas con datos simulados (*mocks*) mientras el equipo backend programa la lógica real sin trabarse. Además, asegura que cuando conectemos las APIs de aerolíneas y hoteles, todos hablen exactamente el mismo idioma y formato.

---

### CRN-3: Separar las consultas pesadas de estadísticas del uso diario de la app
* **Explicación simple:**  
  Imaginate una caja de supermercado. Si una persona va a pagar un alfajor, tarda 5 segundos. Pero si atrás cae el dueño del súper a contar todas las ventas del mes en esa misma caja, la traba y nadie más puede pagar su alfajor.  
  Acá pasa lo mismo: Trotamundos necesita sacar estadísticas gigantes para ganar plata con publicidad o datos (monetización). Esas cuentas tardan un montón y consumen muchísima máquina. Si hacemos esas cuentas pesadas en la misma base de datos donde la gente está buscando sus vuelos o viendo sus pasajes, la app se le va a congelar a todo el mundo.  
* **Impacto en el diseño:**  
  Creamos "dos caminos separados": por un lado viaja la gente normal reservando viajes rápido, y por otro lado (en otra base o réplica separada) corremos los cálculos pesados sin molestar a los usuarios.

---

### CRN-4: Modularizar para asignación de trabajo y documentación ADR
* **¿Qué es?**  
  Separar el sistema en partes claras e independientes para que los miembros del grupo puedan repartirse las tareas sin pisarse el código, dejando asentadas las decisiones importantes.
* **Impacto en el diseño:**  
  Permite que cada integrante se haga responsable de un módulo o servicio en GitHub, facilitando además justificar cada elección técnica mediante plantillas MADR (*Architectural Decision Records*) como exige la cátedra.

---

### CRN-5: Extensibilidad para módulos avanzados (Recomendación / Chatbot)
* **¿Qué es?**  
  Dejar la arquitectura preparada desde el día uno para que sumar funciones complejas en el futuro no obligue a romper o rehacer todo lo que ya funciona.
* **Impacto en el diseño:**  
  Para la segunda entrega del TPE es obligatorio agregar un *Sistema de Recomendación* o un *Chatbot/Asistente virtual*. Este concern asegura que el núcleo de itinerarios y reservas exponga puntos de conexión limpios donde "enchufar" estos componentes inteligentes de manera aislada.
