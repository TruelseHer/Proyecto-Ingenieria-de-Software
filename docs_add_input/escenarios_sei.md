

* **Rendimiento**  
  * **Escenario Propuesto 1:** Un servicio externo como Booking publica un alojamiento disponible en la fecha y ubicación que el cliente desea viajar 

| Fuente del Estímulo | Servicio externo de reserva (Ej: Booking) |
| :---- | :---- |
| **Estímulo** | Se publica un alojamiento como disponible en la ubicación y fecha de un posible viaje |
| **Ambiente (Contexto operativo)** | Operación normal del sistema en el chequeo de servicios externos |
| **Artefacto (Porción del sistema que recibe el estímulo)** | Módulo de gestión de servicios externos |
| **Respuesta** | El sistema gestiona eficientemente los elementos ingresados en las apps externas para sugerir al usuario |
| **Medida de la respuesta** | El tiempo transcurrido entre el evento y la respuesta es como máximo 5 minutos El tiempo entre el estímulo externo y registrar dicho evento es como máximo 2 minutos |


  * **Escenario Propuesto 2:** Un usuario desea comprar un paquete de viaje a través web

| Fuente del Estímulo | Usuario desde un navegador web |
| :---- | :---- |
| **Estímulo** | El usuario navega y quiere explorar opciones de paquetes de viaje |
| **Ambiente (Contexto operativo)** | Operación normal (solicitudes de ingreso al sistema) |
| **Artefacto (Porción del sistema que recibe el estímulo)** | Módulo de gestión de solicitudes de ingreso al sitio |
| **Respuesta** | El sistema despliega contenido sugerido según la identidad del usuario |
| **Medida de la respuesta** | El tiempo de respuesta desde el navegador web del usuario debe ser de 800 ms o menos |

  * **Escenario Propuesto 3:** 

| Fuente del Estímulo | Usuario final operando desde un dispositivo móvil. *(Supuesto)* |
| :---- | :---- |
| **Estímulo** | Solicitud de acceso a la plataforma o navegación hacia una vista/pantalla principal en la aplicación móvil. |
| **Ambiente (Contexto operativo)** | Aplicación móvil (cliente nativo/vista web) del sistema "Trotamundos”  Tiempo de ejecución (*runtime*) bajo condiciones normales de red móvil (ej. 4G/5G) |
| **Artefacto (Porción del sistema que recibe el estímulo)** | Aplicación móvil (cliente nativo/vista web) del sistema "Trotamundos".  |
| **Respuesta** | El cliente móvil procesa el renderizado crítico, aplicando tácticas de optimización de recursos (minimización de *assets*, carga asíncrona y procesamiento progresivo del DOM) para desplegar los primeros elementos visuales en pantalla |
| **Medida de la respuesta** | El primer despliegue de contenido (*First Contentful Paint* / FCP) se completa en un tiempo menor a 1.4 segundos  |

