# Iteración 5 \- Método ADD Sistema "Trotamundos"

## **Paso 2: Selección y justificación de Drivers Seleccionados**

#### **Objetivo de la Iteración**

* Diseñar las estructuras arquitectónicas de la **Capa de Presentación Móvil Nativa, Integración Sensorial (GPS) e Internacionalización Multirregión (i18n / l10n)**, resolviendo el atributo de calidad **QA-7** mediante el desacoplamiento entre el hardware del dispositivo móvil, los motores de localización/mapeo y la adaptación de contratos en el Mobile BFF, garantizando que el viajero obtenga rutas interactivas hacia su alojamiento en menos de 2 segundos y con un esfuerzo cognitivo mínimo (\<= 3 clics).

**Clasificación y Justificación de Drivers Seleccionados**

* **QA-7 (Usabilidad e Internacionalización \- Adaptabilidad y Localización Nativa):**  
  - *Justificación:* Es el último atributo de calidad del Árbol de Utilidad pendiente de diseño arquitectónico. Exige que un usuario en el extranjero (por ejemplo, viajando de EE. UU. a Japón) abra la aplicación móvil y obtenga de manera transparente su ubicación vía GPS nativo, la visualización de mapas locales con la ruta a su hotel, y la conversión contextualizada de moneda y zonas horarias en menos de 2 segundos y en no más de 3 clics de interacción. Requiere definir la interacción entre los sensores del dispositivo, las APIs de mapas y los servicios de contextualización regional.


* **CON-1 (Soporte de Clientes Web y Móvil Nativo):**  
  - *Justificación:* Restricción rectora que condiciona a no resolver la usabilidad como una simple página web embebida lenta (Webview pura), sino mediante una arquitectura que aproveche las capacidades de hardware del cliente móvil nativo (sensores GPS, acelerómetro, almacenamiento local) sin duplicar la lógica de negocio del backend.


* **QA-1 \[A1R-E3\] (Rendimiento \- Experiencia de Carga y Renderizado Móvil FCP \< 1.4 s):**  
  - *Justificación:* Actúa como driver de contención y rendimiento para la interfaz móvil. La incorporación de librerías de internacionalización (i18n), paquetes de mapas y lectura de sensores de geolocalización no debe penalizar el despliegue del contenido visual inicial ni sobrecargar el consumo de datos móviles del viajero.

## **Paso 3: Elegir Uno o Más Elementos del Sistema a Refinar**

De acuerdo con el objetivo de la iteración y los drivers seleccionados (QA-7, CON-1 y QA-1), los elementos del sistema a descomponer y refinar corresponden a la frontera de usuario móvil y su adaptador perimetral dedicado:

### 1\. trotamundos.mobile\_client (Aplicación Móvil Nativa)

* **Tipo de refinamiento:** Descomposición interna (vertical) y diseño modular de cliente.  
* **Propósito:** Refinar la arquitectura interna de la aplicación cliente (bajo un patrón de presentación desacoplado como MVVM / Clean Client Architecture) para dar soporte directo a **QA-7**:  
  - Un módulo de abstracción de hardware (location\_hardware\_adapter) que interactúe con los servicios nativos de geolocalización (LocationManager en Android / CoreLocation en iOS) mediante permisos transparentes y bajo consumo de batería.  
  - Un motor de internacionalización del lado del cliente (client\_i18n\_engine) que gestione paquetes de traducción descargables, formatos de fecha/hora según el huso horario local detectado y formateo de divisas.  
  - Un componente de renderizado de rutas y mapas vectoriales (map\_navigation\_controller) que cargue los puntos de interés del hotel del viaje activo y calcule la trayectoria visual en menos de 2 segundos.

### 2\. trotamundos.perimeter.mobile\_bff (Backend for Frontend Móvil)

* **Tipo de refinamiento:** Refinamiento de contratos de interfaz y agregación de contexto.  
* **Propósito:** Adaptar los endpoints del Mobile BFF (instanciado en la Iteración 1\) para suministrar la información regional necesaria sin generar múltiples viajes de red (*chattiness*):  
  - Exponer contratos optimizados (ej. GET /mobile/v1/trips/{id}/navigation-context) que unifiquen en un único payload ligero las coordenadas geográficas normalizadas del alojamiento, las tasas de conversión cambiaria oficiales del destino y las reglas de zonificación horaria.  
  - Incorporar almacenamiento en caché perimetral para datos estáticos de localización y catálogos de divisas, resguardando la latencia móvil de renderizado.

Aquí tienes la redacción formal completa del **Paso 4** y el **Paso 5** para la **Iteración 5** del sistema **Trotamundos**, manteniendo consistencia metodológica con el formato de las iteraciones anteriores y enfocado estrictamente en **QA-7**, **CON-1** y **QA-1 \[A1R-E3\]**.

## **Paso 4: Elegir Uno o Más Conceptos de Diseño que Satisfagan los Drivers**

En este paso se formalizan los conceptos arquitectónicos (patrones de presentación, abstracciones de hardware, tácticas de renderizado y diseño de contratos móviles) seleccionados para resolver la experiencia de usuario móvil, la geolocalización nativa y la internacionalización multirregión (**QA-7**), resguardando el renderizado rápido (**QA-1**) bajo la restricción de cliente móvil nativo (**CON-1**).

### **1\. Identificación y Selección de Conceptos de Diseño**

| Concepto de Diseño / Patrón | Tipo de Concepto | Drivers que Ataca | Justificación de Selección (Rationale) |
| :---- | :---- | :---- | :---- |
| **Arquitectura de Presentación Desacoplada (MVVM \- Model-View-ViewModel con StateFlow)** | Patrón Arquitectónico de Cliente / UI | CON-1, QA-7 | Separa el ciclo de vida de la vista gráfica de la lógica de presentación y sincronización regional. Mantiene el estado de navegación de forma reactiva, evitando bloqueos en el hilo principal de renderizado (UI Thread) y permitiendo calcular la ruta al hotel con el mínimo esfuerzo de interacción ($\leq 3$ clics). |
| **Capa de Abstracción de Hardware de Sensores (Hardware Abstraction Layer / Adapter Pattern)** | Patrón de Diseño / Aislamiento de Plataforma | CON-1, QA-7 | Aísla el consumo del GPS nativo (FusedLocationProvider en Android y CoreLocation en iOS) detrás de una interfaz unificada. Permite obtener coordenadas con bajo consumo de batería, control transparente de permisos y fallback a triangulación por red celular sin acoplar la lógica de la app a un sistema operativo específico. |
| **Internacionalización y Localización Híbrida Descentralizada (Client-Driven i18n & Regional Formatting)** | Táctica de Usabilidad e Internacionalización | QA-7, QA-1 | Las traducciones estáticas se empaquetan localmente en el binario móvil, mientras que las reglas dinámicas de conversión de divisas y desfases de husos horarios se obtienen del BFF. La interfaz adapta números, símbolos monetarios y calendarios de forma instantánea en memoria sin necesidad de esperar llamadas de red bloqueantes. |
| **Carga Diferida y Renderizado Vectorial en Memoria (Lazy Loading & Vector Tiles Caching)** | Táctica de Rendimiento y Usabilidad | QA-7, QA-1 | Reduce el tiempo de visualización de mapas a menos de 2 segundos. Se descargan mosaicos vectoriales (*vector tiles*) ligeros y se reutiliza la caché del dispositivo para los mapas locales, garantizando que el First Contentful Paint se mantenga por debajo de 1.4 segundos sin saturar la red en conexiones móviles en el extranjero. |
| **Agregación de Contexto Orientada al Viajero Móvil (Mobile-Specific Endpoint Aggregation en BFF)** | Patrón de Integración / Perimetral | CON-1, QA-1 \[A1R-E3\], QA-7 | El Mobile BFF consolida en un único contrato ligero las coordenadas geográficas del hotel, la tasa de cambio de divisa y las credenciales efímeras para mapas vectoriales, eliminando el problema de *chattiness* (múltiples peticiones HTTP en roaming).  |

### **2\. Conceptos de Diseño Descartados**

#### **A. Aplicación Híbrida Basada en WebViews Embebidas (Pure Hybrid Webview Wrapper)**

* **Descripción:** Empaquetar una interfaz web responsiva dentro de un contenedor móvil básico y consultar mapas a través de un iframe web.  
 
* **Motivo de descarte:** Viola la restricción de cliente móvil nativo (**CON-1**). La sobrecarga de inicialización del motor web del dispositivo retrasa el FCP por encima de los 1.4 segundos y genera una experiencia lenta con latencias superiores a 2 segundos en el cálculo interactivo de rutas de mapas, degradando **QA-7**.  


#### **B. Resolución Centralizada de Traducciones y Conversión de Monedas por Cada Interacción (Server-Side UI Formatting)**

* **Descripción:** Enviar al servidor backend cada solicitud de pantalla para que devuelva los textos traducidos y los montos ya formateados con la moneda extranjera.  

* **Motivo de descarte:** Cada clic requeriría un viaje de red internacional de ida y vuelta (*round-trip latency*), volviendo imposible asegurar indicaciones de ruta y precios en menos de 2 segundos en redes móviles extranjeras, además de sobrecargar los BFFs con responsabilidades puras de presentación.  


#### **C. Sondeo Continuo de Alta Precisión GPS sin Suspensión (Continuous High-Accuracy GPS Polling)**

* **Descripción:** Mantener el chip GPS del dispositivo activo continuamente en segundo plano para registrar cada cambio de posición del usuario.  

* **Motivo de descarte:** Provoca un consumo acelerado de la batería del dispositivo del viajero y genera bloqueos si la señal satelital se pierde al llegar a aeropuertos o estaciones subterráneas. Se descarta a favor de lecturas bajo demanda (*On-Demand Location Request*) con límites de precisión adaptativos y geocodificación inversa cacheada.  


## **Paso 5: Instanciar Elementos Arquitectónicos y Asignar Responsabilidades**

En este paso se instancian los módulos de software del cliente móvil (trotamundos.mobile\_client) y las extensiones perimetrales en trotamundos.perimeter.mobile\_bff, definiendo sus contratos e interfaces de interacción.

### **1\. Catálogo de Elementos Instanciados y Asignación de Responsabilidades**

| Subsistema / Ámbito | Elemento Instanciado | Tipo de Elemento | Responsabilidades Asignadas (Responsibility Allocation) |
| :---- | :---- | :---- | :---- |
| **trotamundos.mobile\_client** | **NavigationViewModel** | ViewModel de Presentación (MVVM) | • Coordinar el estado reactivo de la pantalla de desplazamiento hacia el hotel. • Manejar el flujo de interacción de usuario para garantizar que con un máximo de 3 clics ("Abrir App" $\rightarrow$ "Ver Viaje Activo" $\rightarrow$ "Cómo Llegar") se inicie la navegación visual (QA-7). • Orquestar la solicitud concurrente de posición GPS y la obtención del contexto del alojamiento. |
| **trotamundos.mobile\_client** | **LocationHardwareAdapter** | Módulo de Abstracción de Hardware | • Proveer una interfaz agnóstica para interactuar con los sensores de ubicación del sistema operativo (Android/iOS) (CON-1). • Gestionar permisos de usuario en tiempo de ejecución. • Consultar la posición actual (latitud, longitud, precisión) de manera rápida en menos de 500 ms usando la última ubicación conocida (*Last Known Location*) como punto de partida inmediato. |
| **trotamundos.mobile\_client** | **ClientI18nEngine** | Motor de Internacionalización y Divisas | • Aplicar localmente las reglas de formato cultural (símbolos de moneda, separadores decimales, formatos horarios de 12/24 h). • Traducir etiquetas visuales y descripciones en destino utilizando recursos locales empaquetados en el binario • Realizar la conversión entre la divisa local del destino y la divisa de referencia del viajero con las tasas suministradas por el BFF. |
| **trotamundos.mobile\_client** | **MapVectorRenderer** | Componente UI / Mapeo Vectorial | • Renderizar la capa base de mapas utilizando renderizado vectorial nativo acelerado por GPU (QA-1). • Trazar la polilínea de la ruta calculada desde la posición GPS del usuario hacia el hotel en menos de 2 segundos (QA-7). • Cachear en el almacenamiento local del dispositivo los mosaicos de mapas (*vector tiles*) para acceso offline o bajo conectividad intermitente. |
| **trotamundos.perimeter** | **MobileNavigationBFFController** | Controlador Perimetral / Capa BFF | • Exponer el endpoint especializado GET /mobile/v1/trips/{tripId}/navigation-context. • Agregar datos de la reserva de alojamiento activa (dirección, coordenadas lat/lon normalizadas), datos de conversión de divisa local y zona horaria en un único payload JSON liviano ($<15$ KB) para cumplir FCP $<1.4$ s (QA-1). |
| **trotamundos.perimeter** | **ExchangeRateGatewayAdapter** | Adaptador de Integración Perimetral | • Consultar periódicamente servicios financieros externos para actualizar tasas de cambio oficiales. • Guardar en memoria caché (Redis) las tablas de equivalencia de monedas para no incurrir en demoras durante las peticiones móviles. |

### 

### **2\. Definición Formal de Interfaces y Gobernanza Contractual (*Spec-Driven Development*)** 

[A. Contrato REST: MobileNavigationAPI.yaml (OpenAPI 3.0)](../spec/Spec-Driven_Devolpement.md#a-contrato-rest-mobilenavigationapiyaml-openapi-30)  
[B. Interfaz de Software Nativa: ILocationService (Abstracción de Hardware Móvil)](../spec/Spec-Driven_Devolpement.md#b-interfaz-de-software-nativa-ilocationservice-abstracción-de-hardware-móvil)  
[C. Interfaz de Cálculo de Ruta: IRouteNavigationEngine](../spec/Spec-Driven_Devolpement.md#c-interfaz-de-cálculo-de-ruta-iroutenavigationengine)

## **Paso 6: Vistas y ADR Registrados**

* De los ADR-18 a ADR-19 corresponden a la Iteración 5

[ADR-18](../adr/ADRS.md#adr-18-adopción-del-patrón-mvvm-y-abstracción-de-hardware-de-localización-para-la-aplicación-móvil)
[ADR-19](../adr/ADRS.md#adr-19-localización-e-internacionalización-i18n-híbrida-en-memoria-con-agregación-contextual-en-mobile-bff)

* Vistas  
  * Vista 1 Módulos  
  * Vista 2 C\&C  
  * Vista 3 Secuencia

## **Paso 7: Analizar el Diseño, Evaluar Resultados y Actualizar el Design Backlog**

### **1\. Evaluación del Objetivo de la Iteración y Revisión del Diseño**

* **Objetivo de la Iteración:** Diseñar las estructuras arquitectónicas de la **Capa de Presentación Móvil Nativa, Integración Sensorial (GPS) e Internacionalización Multiregión (i18n / l10n)**, resolviendo el atributo de calidad **QA-7** mediante el desacoplamiento entre el hardware del dispositivo móvil, los motores de localización/mapeo y la adaptación de contratos en el Mobile BFF, garantizando que el viajero obtenga rutas interactivas hacia su alojamiento en menos de 2 segundos y con un esfuerzo cognitivo mínimo ($\leq$ 3 clics).  
* **Evaluación General:** El objetivo de la iteración ha sido **cumplido satisfactoriamente**. A través de las vistas arquitectónicas modeladas en el Paso 6 y los registros **ADR-18** a **ADR-20**, se resolvió la integración eficiente con los sensores nativos del sistema operativo móvil (Android/iOS) sin comprometer la autonomía de batería ni el hilo de renderizado gráfico de la interfaz. La adopción del patrón **MVVM** complementado con una **Capa de Abstracción de Hardware (ILocationService)** y la estrategia de posicionamiento inicial por *Last Known Location* garantiza la obtención de coordenadas en $<500$ ms. Asimismo, el patrón de **agregación contextual en el Mobile BFF** entrega en un único payload ligero ($<15$ KB) las coordenadas del hotel, la tasa cambiaria oficial y la zonificación temporal, permitiendo que el motor local ClientI18nEngine y el renderizador de mapas vectoriales sobre caché local desplieguen la ruta activa en menos de 2 segundos y en 3 clics de interacción, satisfaciendo la medida de respuesta de **QA-7** sin degradar el rendimiento móvil (**QA-1** \[A1R-E3\]).

### **2\. Justificación del Cumplimiento de Drivers y Definición de Abordaje (*Definition of Addressed*)**

| Driver / Escenario | Estado al Finalizar It. 5 | Evidencia Arquitectónica y Justificación (Definition of Addressed) |
| :---- | :---- | :---- |
| **QA-7** *(Usabilidad e Internacionalización \- Localización en $<2$ s y $\leq 3$ clics)* | **Abordado Completamente** | Se diseñó la arquitectura de presentación reactiva MVVM con abstracción de hardware GPS bajo demanda y *Last Known Location* inicial (**ADR-18**). Se implementó la agregación de contexto de destino en el Mobile BFF y el motor local ClientI18nEngine para conversiones numéricas y horarias en memoria (**ADR-19**). Junto con el renderizado de mosaicos vectoriales cacheados en disco, se asegura el despliegue de la ruta al hotel en $<2$ segundos y dentro del flujo guiado de 3 clics sin pasos manuales. |
| **CON-1** *(Soporte de Clientes Web y Móvil Nativo)* | **Abordado Completamente** | Quedó ratificada la separación estricta: se evitaron contenedores WebView lentos, dotando a la aplicación móvil de soporte nativo para sensores, almacenamiento de mosaicos y renderizado acelerado por GPU, alimentada de forma desacoplada por el Mobile BFF (**ADR-02**, **ADR-18**). |
| **QA-1 \[A1R-E3\]** *(Rendimiento Móvil: FCP $<1.4$ s)* | **Abordado Completamente** | La agregación de datos en un único payload comprimido ($<15$ KB) en MobileNavigationBFFController y el almacenamiento en caché local evitan llamadas redundantes en itinerancia (*roaming*), asegurando que el renderizado inicial y las pantallas de navegación cumplan con el umbral de FCP $<1.4$ s. |
| **QA-1 \[A1R-E1, A1R-E2\]** *(Rendimiento Web $\leq 800$ ms e Integración $\leq 5$ min)* | **Parcialmente Abordado** | Las tácticas de diseño estructurales (Redis distribuido, BFFs especializados, segregación transaccional OLTP/OLAP y desacoplamiento con Kafka) están formalizadas. Conforme al estándar de evaluación ADD, permanece en estado parcial hasta su medición cuantitativa final sobre el entorno de pruebas de carga. |
| **QA-3** *(Escalabilidad: Picos del \+500% ante eventos disruptivos)* | **Parcialmente Abordado** | Se mantienen implementadas las tácticas de autoescalado elástico HPA, réplicas de lectura en PostgreSQL y colas elásticas en el broker. Requiere validación empírica mediante pruebas de estrés sobre la infraestructura desplegada. |
| **CRN-1, CRN-2, CRN-3, CRN-4, CRN-5** | **Abordados Completamente** | Cerrados formalmente en iteraciones previas: estructura macro base (It. 1), gobierno *Spec-Driven* (It. 2), segregación analítica (It. 1), modularización y puntos de extensión (It. 4). |
| **CON-2, QA-2, QA-4, QA-5, QA-6** | **Abordados Completamente** | Cerrados formalmente en iteraciones previas: adaptadores resilientes con Circuit Breaker y DLQ (It. 4), alta disponibilidad Multi-AZ con SLA $\leq 5$ min/mes (It. 3), seguridad y control de acceso estricto al 100% (It. 3), y extensibilidad de transporte polimórfico en $<2$ sprints (It. 4). |
| **US-01 a US-07** | **Abordados Completamente** | Las siete historias de usuario funcionales del sistema cuentan con modelos de dominio, controladores, adaptadores o pipelines de datos analíticos completamente formalizados |

**Actualización del Design Backlog:**

| Categoría | Especificación | No abordado (Hasta el momento) | Parcialmente abordado | Abordado Completamente |
| :---- | :---- | :---- | :---- | :---- |
| Incumbencias Arquitectónicas | CNR 1 |  |   |  x |
|   | CNR 2 |  |   |  x |
|   | CNR 3 |  |   |  x |
|   | CNR 4 |  |   |  x |
|   | CNR 5 |  |   |  x |
| Restricciones | CON-1 |  |   |  x |
|   | CON-2 |  |   |  x |
| Atributos de Calidad | QA-1 |  |  x |   |
|   | QA-2 |  |   |  x |
|   | QA-3 |  |  x |   |
|   | QA-4 |  |   |  x |
|   | QA-5 |  |   |  x |
|   | QA-6 |  |   |  x |
|   | QA-7 |  |   |  x |
| Requerimientos Funcionales Primarios | US-01 |  |   |  x |
|   | US-02 |  |   |  x |
|   | US-03 |  |  |  x |
|   | US-04 |  |   |  x |
|   | US-05 |  |   |  x |
|   | US-06 |  |   |  x |
|   | US-07 |  |   |  x |

### 