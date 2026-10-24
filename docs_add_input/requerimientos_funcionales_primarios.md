### US-01: Procesamiento y Escaneo Automático de Correos

*   **Como** viajero registrado en Trotamundos,
*   **Quiero** que el sistema sondee y analice automáticamente mi bandeja de correo vinculada en busca de confirmaciones de viajes,
*   **Para** que mis nuevas reservas (vuelos, hoteles, autos, trenes, buses) se agreguen al itinerario sin necesidad de cargarlas manualmente.

#### Criterios de Aceptación

*    El sistema solicita autorización explícita (OAuth 2.0 / tokens seguros) para acceder a la bandeja del proveedor de correo, garantizando la privacidad y seguridad de la información.
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

*    Las integraciones se estructuran mediante contratos formales (OpenAPI / JSON Schema) siguiendo la metodología _Spec-driven_.
*    Cualquier actualización o cancelación informada por la fuente externa impacta en la vista del usuario en un tiempo menor o igual a **5 minutos**.
*    El subsistema de integración implementa políticas de reintentos (_backoff_ exponencial) y tolerancia a fallos ante caídas o lentitud de proveedores externos.
*    Las reservas con estados especiales (reprogramado, demorado, cancelado) se visualizan con indicadores claros de alerta en el panel.