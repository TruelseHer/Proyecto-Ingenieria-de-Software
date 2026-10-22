
### **Rendimiento (Performance)**

* **Escenario 1: Notificación de Alojamiento Disponible (Integración Externa)**
  * **Fuente del estímulo:** Servicio externo de reserva (ej. Booking).
  * **Estímulo:** Se publica un alojamiento como disponible en la ubicación y fecha de un posible viaje.
  * **Ambiente (Contexto operativo):** Operación normal del sistema en el chequeo de servicios externos.
  * **Artefacto:** Módulo de gestión de servicios externos.
  * **Respuesta:** El sistema gestiona eficientemente los elementos ingresados en las apps externas para sugerir al usuario.
  * **Medida de la respuesta:** El tiempo transcurrido entre el evento y la respuesta es como máximo de **5 minutos**. El tiempo entre el estímulo externo y el registro de dicho evento es como máximo de **2 minutos**.
* **Escenario 2: Navegación de Paquetes de Viaje en la Web**
  * **Fuente del estímulo:** Usuario navegando desde un navegador web.
  * **Estímulo:** El usuario navega y quiere explorar opciones de paquetes de viaje.
  * **Ambiente (Contexto operativo):** Operación normal (solicitudes de ingreso al sistema).
  * **Artefacto:** Módulo de gestión de solicitudes de ingreso al sitio.
  * **Respuesta:** El sistema despliega contenido sugerido según la identidad del usuario.
  * **Medida de la respuesta:** El tiempo de respuesta desde el navegador web del usuario debe ser de **800 ms o menos**.
* **Escenario 3: Acceso y Primer Renderizado en la Aplicación Móvil**
  * **Fuente del estímulo:** Usuario final operando desde un dispositivo móvil.
  * **Estímulo:** Solicitud de acceso a la plataforma o navegación hacia una vista/pantalla principal en la aplicación móvil.
  * **Ambiente (Contexto operativo):** Aplicación móvil del sistema Trotamundos en tiempo de ejecución (*runtime*) bajo condiciones normales de red móvil (ej. 4G/5G).
  * **Artefacto:** Aplicación móvil (cliente nativo/vista web) del sistema Trotamundos.
  * **Respuesta:** El cliente móvil procesa el renderizado crítico, aplicando tácticas de optimización de recursos (minimización de *assets*, carga asíncrona y procesamiento progresivo del DOM) para desplegar los primeros elementos visuales en pantalla.
  * **Medida de la respuesta:** El primer despliegue de contenido (*First Contentful Paint* / FCP) se completa en un tiempo menor a **1.4 segundos**.

### **Disponibilidad (Availability)**

* **Escenario 1: Tolerancia a Fallas en Infraestructura Crítica**
  * **Fuente del estímulo:** Falla interna del sistema o caída de la infraestructura (ej. nodo de servidor o base de datos).
  * **Estímulo:** Un componente crítico del sistema se cae inesperadamente, interrumpiendo el servicio normal.
  * **Ambiente (Contexto operativo):** Operación normal con carga global de usuarios concurrentes.
  * **Artefacto:** Infraestructura de servidores y bases de datos del sistema.
  * **Respuesta:** El sistema detecta la caída, enruta automáticamente el tráfico hacia un servidor de respaldo (*failover*) y notifica al equipo de operaciones (DevOps), manteniendo la sesión de los usuarios activa.
  * **Medida de la respuesta:** El servicio se restablece completamente asegurando que el tiempo total de inactividad no planificada **no supere los 5 minutos acumulados por mes**.
