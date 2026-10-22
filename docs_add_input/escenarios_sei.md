# Escenarios de atributos de calidad con la plantilla SEI
## 1. **Rendimiento (Performance)**

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

## 2. **Disponibilidad (Availability)**

* **Escenario 1: Tolerancia a Fallas en Infraestructura Crítica**
  * **Fuente del estímulo:** Falla interna del sistema o caída de la infraestructura (ej. nodo de servidor o base de datos).
  * **Estímulo:** Un componente crítico del sistema se cae inesperadamente, interrumpiendo el servicio normal.
  * **Ambiente (Contexto operativo):** Operación normal con carga global de usuarios concurrentes.
  * **Artefacto:** Infraestructura de servidores y bases de datos del sistema.
  * **Respuesta:** El sistema detecta la caída, enruta automáticamente el tráfico hacia un servidor de respaldo (*failover*) y notifica al equipo de operaciones (DevOps), manteniendo la sesión de los usuarios activa.
  * **Medida de la respuesta:** El servicio se restablece completamente asegurando que el tiempo total de inactividad no planificada **no supere los 5 minutos acumulados por mes**.

## 3. Escalabilidad (Scalability)
*La plataforma debe dimensionarse para soportar una base de 15 millones de cuentas, operando con 2 millones de usuarios activos semanales y soportando dramáticos picos de tráfico ante eventos disruptivos masivos.*

*   **Fuente del estímulo:** Usuarios del sistema (viajeros).
*   **Estímulo:** Ocurre un evento disruptivo masivo (ej. tormenta global que cancela vuelos), generando un pico dramático repentino de usuarios ingresando a la plataforma para reprogramar sus itinerarios.
*   **Artefacto:** Capa de servicios y base de datos de la plataforma.
*   **Entorno:** Operación bajo alta demanda, con 2 millones de usuarios activos semanales y un incremento súbito del 500% en el tráfico en pocos minutos.
*   **Respuesta:** La plataforma escala horizontalmente aprovisionando de manera automática nuevas instancias de servidores y equilibrando la carga para procesar las peticiones simultáneas sin degradar el servicio.
*   **Medida de respuesta:** El sistema maneja el pico de tráfico exitosamente, manteniendo el tiempo de carga de las páginas por debajo de 3 segundos y sin rechazar conexiones válidas ni perder datos.

## 4. Interoperabilidad (Interoperability)
*Integración fluida con intercambios de aerolíneas, hoteles y arrendadoras a través de APIs estándar, además de interoperar con proveedores de correo y redes sociales.*

*   **Fuente del estímulo:** Sistemas externos (APIs de Aerolíneas/Hoteles, Proveedores de Correo Electrónico).
*   **Estímulo:** El usuario recibe un correo electrónico con una confirmación de reserva que el sistema debe procesar.
*   **Artefacto:** Módulo de integraciones, escaneo de correos y adaptadores de API.
*   **Entorno:** Operación normal.
*   **Respuesta:** El sistema se autentica con el proveedor de correo, escanea el mensaje utilizando las APIs estándar, extrae los datos relevantes, cruza la información con los sistemas de aerolíneas/hoteles y actualiza el itinerario del usuario.
*   **Medida de respuesta:** El itinerario se actualiza y sincroniza en el sistema del usuario con un 100% de precisión en el mapeo de datos, intercambiando la información en un formato estándar (ej. JSON/REST) sin requerir intervención manual.
