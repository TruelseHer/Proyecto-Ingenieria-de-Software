graph LR
    U["<b>Utilidad</b><br>(Trotamundos)"]

    %% Atributos
    U --> Perf["<b>Rendimiento</b>"]
    U --> Avail["<b>Disponibilidad</b>"]
    U --> Scal["<b>Escalabilidad</b>"]
    U --> Interop["<b>Interoperabilidad</b>"]
    U --> Sec["<b>Seguridad</b>"]
    U --> Evol["<b>Evolucionabilidad</b>"]
    U --> Usab["<b>Usabilidad</b>"]

    %% Sub-refinamientos
    Perf --> P_Lat["Latencia de Integración"]
    Perf --> P_Web["Tiempo de Respuesta Web"]
    Perf --> P_Mob["Experiencia Móvil (FCP)"]

    Avail --> A_Fail["Tolerancia a Fallas"]
    Scal --> S_Peak["Capacidad ante Picos"]
    Interop --> I_Sync["Ingesta de Reservas"]
    Sec --> S_Access["Control de Acceso"]
    Evol --> E_Ext["Extensibilidad Transportes"]
    Usab --> U_Loc["Localización y Rutas"]

    %% Hojas y Prioridades
    P_Lat --> L1["<b>[A1R-E1] (M, M)</b><br>Notificar alojamiento externo en &le; 5 min (registro &le; 2 min)"]
    P_Web --> L2["<b>[A1R-E2] (A, B)</b><br>Desplegar sugerencias web en &le; 800 ms"]
    P_Mob --> L3["<b>[A1R-E3] (M, M)</b><br>Primer renderizado crítico (FCP) móvil en &lt; 1.4 s"]

    A_Fail --> L4["<b>[A2D-E1] (A, A)</b><br>Failover automático sin perder sesión; downtime &le; 5 min/mes"]

    S_Peak --> L5["<b>[A3E-E1] (A, A)</b><br>Soportar pico de +500% tráfico; carga &lt; 3 s sin pérdida de datos"]

    I_Sync --> L6["<b>[A4I-E1] (A, M)</b><br>Sincronizar reservas por correo vía APIs al 100% sin intervención manual"]

    S_Access --> L7["<b>[A5S-E1] (A, M)</b><br>Bloquear 100% accesos indebidos a itinerarios/correos y auditar"]

    E_Ext --> L8["<b>[A6E-E1] (M, M)</b><br>Incorporar trenes/buses en &lt; 2 sprints sin tocar código core"]

    U_Loc --> L9["<b>[A7U-E1] (M, B)</b><br>Detectar GPS, adaptar mapa/moneda en &lt; 2 s y &le; 3 clics"]

    %% Estilos
    style U fill:#2b3a42,stroke:#1e272c,stroke-width:2px,color:#fff
    style L2 fill:#d4edda,stroke:#28a745,stroke-width:2px
    style L4 fill:#f8d7da,stroke:#dc3545,stroke-width:2px
    style L5 fill:#f8d7da,stroke:#dc3545,stroke-width:2px
    style L6 fill:#fff3cd,stroke:#ffc107,stroke-width:1.5px
    style L7 fill:#fff3cd,stroke:#ffc107,stroke-width:1.5px
