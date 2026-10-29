# Contratos de interfaces Spec-Driven

# **Correspondiente a Iteración 2:**

### A. Contrato REST Síncrono: TravelCoreAPI.yaml (Especificación OpenAPI 3.0)

Gobierna la interacción síncrona entre los BFFs perimetrales y el servicio trotamundos.core\_travel:

* **GET /api/v1/trips/active**: Retorna el itinerario consolidado activo del usuario autenticado (US-05).  
* **POST /api/v1/trips**: Creación manual de un nuevo viaje agrupador (US-05).  
* **POST /api/v1/trips/{tripId}/reservations**: Registro manual de una nueva reserva asociada a un viaje (US-04).  
* **PUT /api/v1/reservations/{reservationId}**: Modificación o actualización manual de datos de una reserva (US-04).  
* **DELETE /api/v1/reservations/{reservationId}**: Cancelación de una reserva existente (US-04).  
* **POST /api/v1/reservations/{reservationId}/agency-support**: Apertura de ticket de contingencia o consulta con la agencia preferida asignada (US-06).

###  B. Contrato Asíncrono de Eventos: ReservationDetectedEvent (JSON Schema / AsyncAPI)

Gobierna el payload emitido por trotamundos.ingestion y consumido por trotamundos.core\_travel a través del Message Broker (US-01, US-03):

JSON  
{  
  "\$schema": "http\://json-schema.org/draft-07/schema\#",  
  "title": "ReservationDetectedEvent",  
  "type": "object",  
  "properties": {  
    "eventId": { "type": "string", "format": "uuid" },  
    "userId": { "type": "string" },  
    "source": { "type": "string", "enum": \["EMAIL\_SCAN", "PROVIDER\_API"\] },  
    "providerName": { "type": "string" },  
    "reservationType": { "type": "string", "enum": \["FLIGHT", "HOTEL", "RENTAL", "TRAIN", "BUS"\] },  
    "confirmationCode": { "type": "string" },  
    "startDate": { "type": "string", "format": "date-time" },  
    "endDate": { "type": "string", "format": "date-time" },  
    "details": { "type": "object" },  
    "timestamp": { "type": "string", "format": "date-time" }  
  },  
  "required": \["eventId", "userId", "source", "reservationType", "confirmationCode", "startDate", "timestamp"\]  
}

### C. Configuración Externalizada: EnvironmentConfigProfile

Conjunto de parámetros desacoplados del código fuente para permitir portabilidad entre entornos de desarrollo, integración y despliegue del prototipo:

* POLLING\_EMAIL\_INTERVAL\_MS: Frecuencia de escaneo de bandejas de entrada.  
* EXTERNAL\_PROVIDER\_\*\_BASE\_URL: Direcciones base de las APIs de aerolíneas, hoteles y arrendadoras.  
* EXTERNAL\_PROVIDER\_\*\_API\_KEY: Credenciales de acceso seguras para servicios de terceros.  
* SPRING\_DATASOURCE\_URL / BROKER\_URL: Parámetros de conexión a la base de datos PostgreSQL y al Message Broker.

# **Correspondiente a Iteración 4:**

### A. Contrato Asíncrono de Ciclo de Vida: TripLifecycleEvent (JSON Schema / AsyncAPI)

Gobierna los eventos emitidos por trotamundos.core\_travel y consumidos por trotamundos.analytics para la recopilación de datos de monetización (US-07):

JSON  
{  
  "\$schema": "http\://json-schema.org/draft-07/schema\#",  
  "title": "TripLifecycleEvent",  
  "type": "object",  
  "properties": {  
    "eventId": { "type": "string", "format": "uuid" },  
    "tripId": { "type": "string", "format": "uuid" },  
    "userId": { "type": "string" },  
    "eventType": {   
      "type": "string",   
      "enum": \["TRIP\_CREATED", "RESERVATION\_ADDED", "TRIP\_MODIFIED", "TRIP\_COMPLETED", "TRIP\_ARCHIVED", "TRIP\_CANCELLED"\]   
    },  
    "destinationSummary": {  
      "type": "object",  
      "properties": {  
        "city": { "type": "string" },  
        "countryCode": { "type": "string" },  
        "startDate": { "type": "string", "format": "date-time" },  
        "endDate": { "type": "string", "format": "date-time" }  
      },  
      "required": \["city", "countryCode", "startDate", "endDate"\]  
    },  
    "transportTypes": {  
      "type": "array",  
      "items": { "type": "string", "enum": \["FLIGHT", "HOTEL", "RENTAL", "TRAIN", "BUS"\] }  
    },  
    "timestamp": { "type": "string", "format": "date-time" }  
  },  
  "required": \["eventId", "tripId", "userId", "eventType", "destinationSummary", "timestamp"\]  
}

#### 

### B. Contrato OpenAPI del Servicio Analítico de Monetización: AnalyticsAPI.yaml

Gobierna las consultas pesadas y reportes expuestos por trotamundos.analytics para los usuarios analistas de negocio (US-07):

* GET /api/v1/analytics/trends/destinations: Retorna el ranking de destinos de viaje consolidados y segmentados por temporalidad (mes/trimestre) y tasa de crecimiento.  
    
* GET /api/v1/analytics/metrics/disruptions: Provee agregaciones de demoras, reprogramaciones y cancelaciones desglosadas por tipo de transporte (vuelo, tren, bus) y proveedor turístico.  
    
* GET /api/v1/analytics/monetization/summary: Genera el consolidado histórico anonimizado de volumen de itinerarios completados para conciliación con aerolíneas y agencias preferidas.

#### 

### C. Interfaz de Software para Plugins de Recomendación (RecommendationStrategy)

Gobierna la integración desacoplada entre el host del módulo y los algoritmos predictivos (CRN-5, QA-6):

Java  
package trotamundos.recommendations.engine;

import java.util.List;  
import java.util.UUID;

public interface RecommendationStrategy {  
    /\*\*  
     \* Genera sugerencias personalizadas de destinos o servicios complementarios  
     \* sin acoplamiento síncrono con la base de datos operativa.  
     \*   
     \* @param context Contexto de viaje del usuario (destino, fechas y tipo de movilidad).  
     \* @return Lista de ítems recomendados ordenados por relevancia.  
     \*/  
    List\<RecommendationItem\> generateRecommendations(UserTripContext context);  
      
    /\*\*  
     \* Identificador único de la estrategia de recomendación para gobernanza y métricas A/B.  
     \*/  
    String getStrategyIdentifier();  
}

#### 

### D. Especificación Táctica del Circuit Breaker en Ingesta (Configuración de Resiliencia)

Parámetros declarativos formalizados para proteger las llamadas salientes hacia APIs externas en trotamundos.ingestion (QA-4, CON-2):

* slidingWindowSize: 20 llamadas (ventana rodante de muestreo).  
*   
* failureRateThreshold: 50% (porcentaje de respuestas HTTP 5xx o *timeouts* para abrir el circuito).  
*   
* slowCallRateThreshold: 50% (llamadas cuya latencia supere los 3000 ms).  
*   
* waitDurationInOpenState: 30 segundos (tiempo de aislamiento antes de pasar a estado de prueba *Half-Open*).  
*   
* retryPolicy: Máximo 3 reintentos inmediatos con retroceso exponencial (initialInterval \= 500ms, multiplier \= 2.0, jitter \= 0.2) antes de disparar el desvío a la DLQ.

# **Correspondiente a Iteración 5:**

### A. Contrato REST: MobileNavigationAPI.yaml (OpenAPI 3.0)

Gobierna la interacción síncrona optimizada entre la aplicación móvil nativa y el Mobile BFF:

YAML  
openapi: 3.0.3  
info:  
  title: Trotamundos Mobile Navigation Context API  
  version: 1.0.0  
  description: Endpoint especializado del Mobile BFF para navegación, geolocalización e i18n (QA-7).  
paths:  
  /mobile/v1/trips/{tripId}/navigation-context:  
    get:  
      summary: Obtiene el contexto consolidado de navegación hacia el alojamiento activo  
      parameters:  
        \- name: tripId  
          in: path  
          required: true  
          schema:  
            type: string  
            format: uuid  
        \- name: targetCurrency  
          in: query  
          required: false  
          description: Moneda configurada por el usuario (ej. USD, EUR)  
          schema:  
            type: string  
            default: USD  
      responses:  
        '200':  
          description: Contexto de navegación retornado exitosamente  
          content:  
            application/json:  
              schema:  
                \$ref: '\#/components/schemas/NavigationContextResponse'  
components:  
  schemas:  
    NavigationContextResponse:  
      type: object  
      properties:  
        hotelDetails:  
          type: object  
          properties:  
            hotelName:  
              type: string  
            address:  
              type: string  
            latitude:  
              type: number  
              format: double  
            longitude:  
              type: number  
              format: double  
          required: \[hotelName, address, latitude, longitude\]  
        localization:  
          type: object  
          properties:  
            destinationTimeZone:  
              type: string  
              example: "Asia/Tokyo"  
            localCurrency:  
              type: string  
              example: "JPY"  
            exchangeRateToTarget:  
              type: number  
              format: float  
              example: 154.20  
          required: \[destinationTimeZone, localCurrency, exchangeRateToTarget\]  
        mapRoutingToken:  
          type: string  
          description: Token efímero para consumo de mosaicos vectoriales  
      required: \[hotelDetails, localization\]

#### 

### B. Interfaz de Software Nativa: ILocationService (Abstracción de Hardware Móvil)

Contrato interno desacoplado dentro de la arquitectura de la app móvil nativa:

Java  
package trotamundos.mobile.location;

public interface ILocationService {  
    /\*\*  
     \* Obtiene la posición actual del dispositivo usando GPS nativo o red celular.  
     \* Implementa resolución rápida mediante Last Known Location para garantizar respuesta en \< 500ms.  
     \*   
     \* @param highAccuracy Indica si se fuerza el encendido del satélite GPS.  
     \* @return Coordenadas actuales del viajero (latitud, longitud, precisión).  
     \*/  
    GeoPosition getCurrentPosition(boolean highAccuracy);

    /\*\*  
     \* Comprueba y solicita de forma asíncrona los permisos del sistema operativo.  
     \*/  
    boolean hasLocationPermission();  
}

#### 

### C. Interfaz de Cálculo de Ruta: IRouteNavigationEngine

Gobierna la interacción entre el ViewModel y el motor de dibujo de mapas vectoriales en pantalla:  
Java  
package trotamundos.mobile.navigation;

public interface IRouteNavigationEngine {  
    /\*\*  
     \* Calcula y proyecta la trayectoria sobre el mapa vectorial hacia el hotel destino.  
     \* Cumple la medida de despliegue interactivo en menos de 2 segundos (QA-7).  
     \*   
     \* @param origin Posición actual obtenida por GPS.  
     \* @param destination Coordenadas del hotel provistas por el BFF.  
     \* @return Objeto de ruta con distancia, tiempo estimado de llegada y polilínea gráfica.  
     \*/  
    RouteCalculationResult calculateAndDrawRoute(GeoPosition origin, GeoPosition destination);  
}

