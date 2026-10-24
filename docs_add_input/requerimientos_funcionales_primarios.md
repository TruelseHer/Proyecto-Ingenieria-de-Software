### US-01: Procesamiento y Escaneo Automático de Correos

*   **Como** viajero registrado en Trotamundos,
*   **Quiero** que el sistema sondee y analice automáticamente mi bandeja de correo vinculada en busca de confirmaciones de viajes,
*   **Para** que mis nuevas reservas (vuelos, hoteles, autos, trenes, buses) se agreguen al itinerario sin necesidad de cargarlas manualmente.

#### Criterios de Aceptación

*    El sistema solicita autorización explícita  para acceder a la bandeja del proveedor de correo, garantizando la privacidad y seguridad de la información.
*    El motor de extracción identifica y parsea datos clave: localizador/PNR, fechas, horas, origen, destino, tipo de servicio y proveedor.
*    Toda reserva detectada debe reflejarse en el itinerario de la aplicación dentro de los **5 minutos** posteriores a la recepción del correo.
*    Si un correo no corresponde a una reserva válida o estructurada, se descarta sin almacenar ni exponer el contenido del mensaje.

---

### US-02: Filtros y Whitelist de Remitentes de Correo

*   **Como** usuario consciente de la privacidad de mis comunicaciones,
*   **Quiero** configurar filtros y definir una lista blanca (_whitelist_) de remitentes o dominios autorizados,
*   **Para** garantizar que el proceso de escaneo de Trotamundos sólo inspeccione correos provenientes de mis agencias y proveedores de confianza.

#### Criterios de Aceptación

*    La interfaz de configuración permite agregar, listar, editar y eliminar remitentes o dominios específicos (ej. `@booking.com`, `@latam.com`).
*    El proceso de sondeo ignora y omite de forma estricta cualquier mensaje cuyo remitente no coincida con los dominios o correos habilitados en la lista blanca activa.
*    Se proporciona un listado base de proveedores reconocidos que el usuario puede habilitar o inhabilitar con un solo clic.
*    Las modificaciones en la lista blanca se sincronizan de inmediato para el siguiente ciclo de sondeo programado.

---

### US-03: Integración con Sistemas de Viajes de Terceros vía API

*   **Como** viajero en tránsito,
*   **Quiero** que el sistema se comunique de forma fluida con sistemas externos (aerolíneas, rentadoras, hoteles) mediante APIs estándar,
*   **Para** recibir actualizaciones en tiempo real sobre cambios de itinerario, puertas de embarque, retrasos y cancelaciones.

#### Criterios de Aceptación

*    Las integraciones se estructuran mediante contratos formales  siguiendo la metodología _Spec-driven_.
*    Cualquier actualización o cancelación informada por la fuente externa impacta en la vista del usuario en un tiempo menor o igual a **5 minutos**.
*    El subsistema de integración implementa políticas de reintentos y tolerancia a fallos ante caídas o lentitud de proveedores externos.
*    Las reservas con estados especiales (reprogramado, demorado, cancelado) se visualizan con indicadores claros de alerta en el panel.

### US-04: Gestión Manual de Reservas

* **Como** usuario del panel de Trotamundos,  
* **Quiero** crear, modificar o eliminar reservas de manera manual,  
* **Para** gestionar arreglos de viaje informales o de proveedores que no ofrecen integraciones por API ni confirmaciones por correo escaneables.  

#### Criterios de Aceptación

-  La interfaz (web y móvil) provee un formulario estructurado para registrar tipo de transporte/alojamiento, fechas, locaciones y notas.
-  En la aplicación web, el tiempo de respuesta al guardar o actualizar una reserva es de **800 ms o menor**.
-  La eliminación de un ítem solicita confirmación previa y actualiza de inmediato el cálculo temporal del itinerario.
-  Los campos modificados de forma manual quedan marcados para evitar ser sobrescritos por sincronizaciones automáticas posteriores.

---

### US-05: Agrupación y Ciclo de Vida del Viaje

* **Como** viajero recurrente,  
* **Quiero** que mis reservas se organicen por viaje y que los viajes completados se archiven automáticamente,  
* **Para** mantener una interfaz limpia y centrada exclusivamente en los desplazamientos actuales y próximos.  

#### Criterios de Aceptación

-  Las reservas con continuidad cronológica y de ubicación se asocian de forma automática bajo un mismo "Viaje" (permitiendo ajustes manuales de agrupación).
-  Al cumplirse la fecha y hora del último evento del itinerario, el viaje se retira de la pantalla principal y pasa a la vista de "Archivados".
-  En plataformas móviles, el despliegue del contenido inicial (*First Contentful Paint*) del panel de viajes no supera los **1.4 segundos**.
-  El usuario puede navegar, consultar y restaurar viajes archivados en cualquier momento desde una sección dedicada.

  
### US-06: Asistencia y Resolución de Problemas con Agencia Preferida

* **Como** viajero afectado por una disrupción o cancelación,  
* **Quiero** conectar directamente una reserva problemática con mi agencia de viajes preferida desde el panel,  
* **Para** agilizar la resolución de incidentes, reprogramaciones y reclamos con un operador de asistencia.  

#### Criterios de Aceptación

-  El usuario puede vincular y configurar su agencia de cabecera en el perfil de usuario.
-  Frente a una contingencia (vuelo cancelado, hotel sobrevendido), la interfaz presenta un acceso directo ("Solicitar resolución").
- Al iniciar la solicitud, el sistema transfiere de forma segura el contexto del viaje y códigos de reserva a la API de la agencia, evitando la reintroducción de datos.
-  El usuario recibe y visualiza un identificador de ticket o estado del trámite directamente sobre la tarjeta de la reserva.

---

### US-07: Recopilación y Consulta de Datos Analíticos (Monetización)

* **Como** analista del equipo de datos de Trotamundos,  
* **Quiero** que el sistema registre y consolide métricas de viajes (destinos, tendencias, cancelaciones) de forma anonimizada,  
* **Para** ejecutar consultas analíticas pesadas orientadas a monetización sin degradar la experiencia de los usuarios activos.  

#### Criterios de Aceptación

-  La captura y envío de eventos analíticos se realiza de forma asíncrona para no penalizar la latencia operativa ni el rendimiento bajo picos de carga (2 millones de usuarios activos semanales).
-  Todos los datos recopilados se anonimizan, desasociando información personal identificable (PII) según los requisitos de seguridad y privacidad.
-  Las consultas complejas y pesadas se dirigen a un repositorio de datos analítico segregado (CQRS / Data Warehouse), preservando la disponibilidad del sistema transaccional.
-  Se habilitan vistas y reportes por rangos temporales, destinos con mayor demanda y tasas de contingencias por aerolínea/proveedor.

