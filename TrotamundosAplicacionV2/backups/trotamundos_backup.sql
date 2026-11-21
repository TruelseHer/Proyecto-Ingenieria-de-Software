--
-- PostgreSQL database dump
--

\restrict ugLdUeSmgfXwZbCGEdl2i6wyo8qmgo7Ya1muWzhkraU6etAktc4bPIaxkZbPWyO

-- Dumped from database version 16.16
-- Dumped by pg_dump version 16.16

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

ALTER TABLE IF EXISTS ONLY public.trips DROP CONSTRAINT IF EXISTS trips_owner_id_fkey;
ALTER TABLE IF EXISTS ONLY public.support_tickets DROP CONSTRAINT IF EXISTS support_tickets_reservation_id_fkey;
ALTER TABLE IF EXISTS ONLY public.sender_rules DROP CONSTRAINT IF EXISTS sender_rules_user_id_fkey;
ALTER TABLE IF EXISTS ONLY public.reservations DROP CONSTRAINT IF EXISTS reservations_trip_id_fkey;
ALTER TABLE IF EXISTS ONLY public.email_confirmations DROP CONSTRAINT IF EXISTS email_confirmations_user_id_fkey;
DROP INDEX IF EXISTS public.idx_trips_owner;
DROP INDEX IF EXISTS public.idx_sender_rules_user;
DROP INDEX IF EXISTS public.idx_reservations_trip;
DROP INDEX IF EXISTS public.idx_email_confirmations_user;
DROP INDEX IF EXISTS public.idx_catalog_offers_type;
DROP INDEX IF EXISTS public.idx_catalog_offers_pnr;
ALTER TABLE IF EXISTS ONLY public.users DROP CONSTRAINT IF EXISTS users_pkey;
ALTER TABLE IF EXISTS ONLY public.users DROP CONSTRAINT IF EXISTS users_email_key;
ALTER TABLE IF EXISTS ONLY public.trips DROP CONSTRAINT IF EXISTS trips_pkey;
ALTER TABLE IF EXISTS ONLY public.travel_agencies DROP CONSTRAINT IF EXISTS travel_agencies_pkey;
ALTER TABLE IF EXISTS ONLY public.support_tickets DROP CONSTRAINT IF EXISTS support_tickets_reservation_id_key;
ALTER TABLE IF EXISTS ONLY public.support_tickets DROP CONSTRAINT IF EXISTS support_tickets_pkey;
ALTER TABLE IF EXISTS ONLY public.sender_rules DROP CONSTRAINT IF EXISTS sender_rules_pkey;
ALTER TABLE IF EXISTS ONLY public.reservations DROP CONSTRAINT IF EXISTS reservations_pkey;
ALTER TABLE IF EXISTS ONLY public.pending_syncs DROP CONSTRAINT IF EXISTS pending_syncs_pkey;
ALTER TABLE IF EXISTS ONLY public.email_confirmations DROP CONSTRAINT IF EXISTS email_confirmations_pkey;
ALTER TABLE IF EXISTS ONLY public.catalog_packages DROP CONSTRAINT IF EXISTS catalog_packages_pkey;
ALTER TABLE IF EXISTS ONLY public.catalog_offers DROP CONSTRAINT IF EXISTS catalog_offers_pkey;
ALTER TABLE IF EXISTS ONLY public.audit_logs DROP CONSTRAINT IF EXISTS audit_logs_pkey;
ALTER TABLE IF EXISTS ONLY public.analytics_facts DROP CONSTRAINT IF EXISTS analytics_facts_pkey;
DROP TABLE IF EXISTS public.users;
DROP TABLE IF EXISTS public.trips;
DROP TABLE IF EXISTS public.travel_agencies;
DROP TABLE IF EXISTS public.support_tickets;
DROP TABLE IF EXISTS public.sender_rules;
DROP TABLE IF EXISTS public.reservations;
DROP TABLE IF EXISTS public.pending_syncs;
DROP TABLE IF EXISTS public.email_confirmations;
DROP TABLE IF EXISTS public.catalog_packages;
DROP TABLE IF EXISTS public.catalog_offers;
DROP TABLE IF EXISTS public.audit_logs;
DROP TABLE IF EXISTS public.analytics_facts;
SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: analytics_facts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.analytics_facts (
    id text NOT NULL,
    pseudonym_id text NOT NULL,
    event_type text NOT NULL,
    destination_city text NOT NULL,
    country_code text NOT NULL,
    transport_types jsonb DEFAULT '[]'::jsonb NOT NULL,
    disrupted boolean DEFAULT false NOT NULL,
    provider_name text DEFAULT ''::text NOT NULL,
    recorded_at timestamp with time zone NOT NULL
);


--
-- Name: audit_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.audit_logs (
    id text NOT NULL,
    actor_user_id text NOT NULL,
    action text NOT NULL,
    resource_type text NOT NULL,
    resource_id text NOT NULL,
    outcome text NOT NULL,
    reason text NOT NULL,
    ip_address text NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: catalog_offers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.catalog_offers (
    id text NOT NULL,
    type text NOT NULL,
    provider_name text NOT NULL,
    provider_email text NOT NULL,
    title text NOT NULL,
    subtitle text NOT NULL,
    origin text NOT NULL,
    destination text NOT NULL,
    destination_city text NOT NULL,
    country_code text NOT NULL,
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone NOT NULL,
    price double precision NOT NULL,
    currency text NOT NULL,
    seats_available integer NOT NULL,
    rating double precision NOT NULL,
    image_url text NOT NULL,
    highlight_tag text NOT NULL,
    confirmation_prefix text NOT NULL,
    sample_pnr text NOT NULL,
    details jsonb DEFAULT '{}'::jsonb NOT NULL
);


--
-- Name: catalog_packages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.catalog_packages (
    id text NOT NULL,
    slug text NOT NULL,
    title text NOT NULL,
    subtitle text NOT NULL,
    destination_city text NOT NULL,
    country_code text NOT NULL,
    region text NOT NULL,
    summary text NOT NULL,
    image_url text NOT NULL,
    badge text NOT NULL,
    duration_days integer NOT NULL,
    price_per_person double precision NOT NULL,
    currency text NOT NULL,
    rating double precision NOT NULL,
    reviews_count integer NOT NULL,
    includes_summary jsonb DEFAULT '[]'::jsonb NOT NULL,
    available_dates jsonb DEFAULT '[]'::jsonb NOT NULL,
    included_segments jsonb DEFAULT '[]'::jsonb NOT NULL
);


--
-- Name: email_confirmations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.email_confirmations (
    id text NOT NULL,
    user_id text NOT NULL,
    from_email text NOT NULL,
    provider_name text NOT NULL,
    subject text NOT NULL,
    body text NOT NULL,
    pnr text NOT NULL,
    type text NOT NULL,
    origin text NOT NULL,
    destination text NOT NULL,
    destination_city text NOT NULL,
    country_code text NOT NULL,
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone NOT NULL,
    price double precision NOT NULL,
    currency text NOT NULL,
    details jsonb DEFAULT '{}'::jsonb NOT NULL,
    status text NOT NULL,
    status_reason text DEFAULT ''::text NOT NULL,
    imported_trip_id text DEFAULT ''::text NOT NULL,
    received_at timestamp with time zone NOT NULL
);


--
-- Name: pending_syncs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.pending_syncs (
    id text NOT NULL,
    user_id text NOT NULL,
    source text NOT NULL,
    provider_name text NOT NULL,
    summary text NOT NULL,
    raw_payload text NOT NULL,
    error_reason text NOT NULL,
    attempts integer NOT NULL,
    status text NOT NULL,
    next_retry_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: reservations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reservations (
    id text NOT NULL,
    trip_id text NOT NULL,
    type text NOT NULL,
    status text NOT NULL,
    provider_name text NOT NULL,
    confirmation_code text NOT NULL,
    origin text NOT NULL,
    destination text NOT NULL,
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone NOT NULL,
    price double precision NOT NULL,
    currency text NOT NULL,
    notes text DEFAULT ''::text NOT NULL,
    source text NOT NULL,
    manually_modified boolean DEFAULT false NOT NULL,
    modified_fields jsonb DEFAULT '[]'::jsonb NOT NULL,
    details jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: sender_rules; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sender_rules (
    id text NOT NULL,
    user_id text NOT NULL,
    pattern text NOT NULL,
    provider_label text NOT NULL,
    category text NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    is_default boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: support_tickets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.support_tickets (
    id text NOT NULL,
    reference_code text NOT NULL,
    reservation_id text NOT NULL,
    trip_id text NOT NULL,
    user_id text NOT NULL,
    agency_id text NOT NULL,
    agency_name text NOT NULL,
    issue_summary text NOT NULL,
    status text NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: travel_agencies; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.travel_agencies (
    id text NOT NULL,
    name text NOT NULL,
    support_email text NOT NULL,
    phone text NOT NULL,
    sla_hours integer NOT NULL,
    specialty text NOT NULL
);


--
-- Name: trips; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.trips (
    id text NOT NULL,
    owner_id text NOT NULL,
    shared_with jsonb DEFAULT '[]'::jsonb NOT NULL,
    shared_with_emails jsonb DEFAULT '[]'::jsonb NOT NULL,
    title text NOT NULL,
    destination_city text NOT NULL,
    country_code text NOT NULL,
    cover_image_url text DEFAULT ''::text NOT NULL,
    status text NOT NULL,
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id text NOT NULL,
    name text NOT NULL,
    email text NOT NULL,
    password_hash text DEFAULT ''::text NOT NULL,
    role text NOT NULL,
    home_currency text NOT NULL,
    preferred_agency_id text NOT NULL,
    mailbox_authorized boolean DEFAULT false NOT NULL,
    connected_email text NOT NULL,
    avatar_initials text NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Data for Name: analytics_facts; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.analytics_facts (id, pseudonym_id, event_type, destination_city, country_code, transport_types, disrupted, provider_name, recorded_at) FROM stdin;
fact-seed-01	anon_8f2c41e9a30b71d2	TRIP_CREATED	Tokio	JP	["FLIGHT", "HOTEL", "TRAIN"]	f	All Nippon Airways	2026-11-13 23:00:00+00
fact-seed-02	anon_19a47b03d8e251c4	TRIP_COMPLETED	Tokio	JP	["FLIGHT", "TRAIN", "HOTEL"]	f	JR Central Shinkansen	2026-11-14 11:00:00+00
fact-seed-03	anon_3d91b277e04f118a	TRIP_COMPLETED	Madrid	ES	["FLIGHT", "HOTEL", "TRAIN", "BUS"]	f	Iberia	2026-11-14 23:00:00+00
fact-seed-04	anon_77c109fa42e881b0	TRIP_CREATED	Madrid	ES	["TRAIN", "HOTEL"]	f	Renfe AVE	2026-11-15 07:00:00+00
fact-seed-05	anon_51a98c02d7e4491b	TRIP_COMPLETED	Bariloche	AR	["FLIGHT", "HOTEL", "RENTAL"]	f	Aerolíneas Argentinas	2026-11-15 19:00:00+00
fact-seed-06	anon_62e014bc99d3187a	TRIP_CREATED	Bariloche	AR	["RENTAL", "BUS", "HOTEL"]	f	Hertz Rent a Car	2026-11-16 03:00:00+00
fact-seed-07	anon_c074a55b91e2340f	TRIP_COMPLETED	Roma	IT	["FLIGHT", "HOTEL", "TRAIN"]	f	ITA Airways	2026-11-16 11:00:00+00
fact-seed-08	anon_44f912ca08b772e1	TRIP_CREATED	Cusco	PE	["FLIGHT", "HOTEL", "TRAIN"]	f	LATAM Airlines	2026-11-16 19:00:00+00
fact-seed-09	anon_99e301ab64d8f221	TRIP_COMPLETED	París	FR	["FLIGHT", "HOTEL", "TRAIN"]	f	Eurostar International	2026-11-17 03:00:00+00
fact-seed-10	anon_b81204fe77a913c5	TRIP_MODIFIED	Tokio	JP	["FLIGHT"]	t	Air France	2026-11-17 09:00:00+00
fact-seed-11	anon_e10492dc55b608a2	TRIP_COMPLETED	Cusco	PE	["TRAIN", "HOTEL"]	f	PeruRail Vistadome	2026-11-17 15:00:00+00
fact-seed-12	anon_d30819ab41c552e9	TRIP_COMPLETED	Roma	IT	["TRAIN", "BUS", "HOTEL"]	f	Booking.com	2026-11-17 19:00:00+00
fact-e0abe182a84bd9f2	anon_7b0b69aaef82d519	TRIP_CREATED	Terminal San Martín de los Andes	INT	null	f		2026-11-17 23:56:58.49432+00
fact-02fe64cb92b4a5a4	anon_7b0b69aaef82d519	RESERVATION_ADDED	Terminal San Martín de los Andes	INT	["BUS"]	f	Vía Bariloche	2026-11-17 23:56:58.498295+00
fact-6814f23833a706f2	anon_f824be8bfb1a8cbd	TRIP_CREATED	Cusco	PE	["FLIGHT", "HOTEL", "TRAIN"]	f	Cusco	2026-11-17 23:58:51.368135+00
fact-570e3cd88331f2af	anon_12975dbc0b4c8c48	TRIP_CREATED	Roma	IT	null	f		2026-11-21 18:35:21.01233+00
fact-b23ff3e9ef54db3c	anon_12975dbc0b4c8c48	TRIP_CREATED	Tandil	AR	null	f		2026-11-21 18:35:50.265647+00
fact-dbd14d8819f1b3e5	anon_12975dbc0b4c8c48	TRIP_ARCHIVED	Roma	IT	null	f		2026-11-21 18:36:06.353553+00
fact-a26d1e46fe347e1e	anon_12975dbc0b4c8c48	TRIP_MODIFIED	Roma	IT	null	f		2026-11-21 18:36:18.339554+00
fact-1bced82283703b55	anon_12975dbc0b4c8c48	TRIP_CREATED	Bariloche	AR	null	f		2026-11-21 18:39:31.482574+00
fact-74a521b97af5ca06	anon_12975dbc0b4c8c48	RESERVATION_ADDED	Bariloche	AR	["FLIGHT"]	f	Aerolíneas Argentinas	2026-11-21 18:39:31.512736+00
fact-15c91db5e1d48e0d	anon_12975dbc0b4c8c48	TRIP_ARCHIVED	Tandil	AR	null	f		2026-11-21 19:01:13.216764+00
fact-e8faa02f8e63c84b	anon_12975dbc0b4c8c48	TRIP_ARCHIVED	Roma	IT	null	f		2026-11-21 19:01:14.07735+00
fact-435529b9dc105dee	anon_12975dbc0b4c8c48	TRIP_ARCHIVED	Bariloche	AR	["FLIGHT"]	f		2026-11-21 19:01:15.907553+00
fact-a48f2cff3e2a5985	anon_12975dbc0b4c8c48	TRIP_CREATED	París	FR	null	f		2026-11-21 19:09:40.529811+00
fact-8eeb82e5a4bbdf11	anon_12975dbc0b4c8c48	RESERVATION_ADDED	París	FR	["HOTEL"]	f	Booking.com	2026-11-21 19:09:40.560277+00
fact-941bfbf0a004a588	anon_12975dbc0b4c8c48	RESERVATION_ADDED	París	FR	["HOTEL"]	f	Booking.com	2026-11-21 19:10:16.609256+00
fact-266b93044a20b20c	anon_12975dbc0b4c8c48	TRIP_CREATED	Bariloche	AR	null	f		2026-11-21 19:10:46.063449+00
fact-f76bb0b44295c377	anon_12975dbc0b4c8c48	RESERVATION_ADDED	París	FR	["HOTEL"]	f	Booking.com	2026-11-21 19:10:53.410023+00
fact-d825f330ae8201a2	anon_58e7a3bb6f73f407	TRIP_CREATED	Bariloche	AR	null	f		2026-11-21 19:25:35.864583+00
fact-b565abccbb295e5b	anon_58e7a3bb6f73f407	RESERVATION_ADDED	Bariloche	AR	["FLIGHT"]	f	Aerolíneas Argentinas	2026-11-21 19:25:35.867056+00
fact-9de737bf9e9a3457	anon_58e7a3bb6f73f407	TRIP_CREATED	Tokio	JP	null	f		2026-11-21 19:25:36.444467+00
fact-95e7631593c9f1c2	anon_58e7a3bb6f73f407	RESERVATION_ADDED	Tokio	JP	["FLIGHT"]	f	All Nippon Airways	2026-11-21 19:25:36.446857+00
fact-c91bdd0b7c691081	anon_58e7a3bb6f73f407	RESERVATION_ADDED	Tokio	JP	["FLIGHT", "HOTEL"]	f	Booking.com	2026-11-21 19:25:36.451872+00
fact-a3ec113d0f3d4022	anon_58e7a3bb6f73f407	RESERVATION_ADDED	Tokio	JP	["FLIGHT", "HOTEL", "TRAIN"]	f	JR Central Shinkansen	2026-11-21 19:25:36.456915+00
fact-ceb8411b9dd6442c	anon_58e7a3bb6f73f407	RESERVATION_ADDED	Bariloche	AR	["FLIGHT", "HOTEL"]	f	Booking.com	2026-11-21 19:25:36.461595+00
fact-845a9990e6d674c5	anon_12975dbc0b4c8c48	RESERVATION_ADDED	Bariloche	AR	["HOTEL"]	f	Booking.com	2026-11-21 19:28:22.779851+00
\.


--
-- Data for Name: audit_logs; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.audit_logs (id, actor_user_id, action, resource_type, resource_id, outcome, reason, ip_address, created_at) FROM stdin;
aud-ae2cea6096066d3b	usr-28802ee52e6b868c	SHARE_TRIP	TRIP	trip-aeaf1492eceae331	ALLOWED	Itinerario compartido por correo con hermes@hotmail.com	127.0.0.1	2026-11-17 23:57:40.308891+00
aud-1f35537c9afe95b5	usr-ed97b9ed908373ef	SHARE_TRIP	TRIP	trip-714723e7ff2894ef	ALLOWED	Itinerario compartido por correo con changalardicd@gmail.com	127.0.0.1	2026-11-17 23:59:09.393422+00
\.


--
-- Data for Name: catalog_offers; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.catalog_offers (id, type, provider_name, provider_email, title, subtitle, origin, destination, destination_city, country_code, start_date, end_date, price, currency, seats_available, rating, image_url, highlight_tag, confirmation_prefix, sample_pnr, details) FROM stdin;
off-flight-ana-8821	FLIGHT	All Nippon Airways	confirmations@ana.co.jp	Vuelo Directo a Tokio Haneda (NH 105)	Boeing 787-9 Dreamliner • Premium Economy • 2 maletas incluidas	Los Ángeles (LAX)	Tokio Haneda (HND)	Tokio	JP	2026-11-25 09:00:00+00	2026-11-25 21:00:00+00	1280	USD	9	4.95	https://images.unsplash.com/photo-1503899036084-c55cdd92da26?auto=format&fit=crop&w=900&q=80	Vuelo Directo 5★ Skytrax	ANA	ANA-8821	{"gate": "114", "seat": "14A", "terminal": "Terminal 3", "cabinClass": "Premium Economy", "flightNumber": "NH 105", "arrivalAirport": "HND", "departureAirport": "LAX"}
off-flight-ib-3409	FLIGHT	Iberia	reservas@iberia.com	Vuelo Transatlántico a Madrid Barajas (IB 6844)	Airbus A350-900 • Turista Premium • Sala VIP en conexión	Buenos Aires (EZE)	Madrid Barajas (MAD)	Madrid	ES	2026-11-27 12:00:00+00	2026-11-28 00:00:00+00	1120	EUR	11	4.87	https://images.unsplash.com/photo-1543783207-ec64e4d95325?auto=format&fit=crop&w=900&q=80	Sin Escalas • Airbus A350	IB	IB-3409	{"gate": "S32", "seat": "09K", "terminal": "T4S", "cabinClass": "Turista Premium", "flightNumber": "IB 6844", "arrivalAirport": "MAD", "departureAirport": "EZE"}
off-flight-ar-1140	FLIGHT	Aerolíneas Argentinas	tickets@aerolineas.com.ar	Vuelo a San Carlos de Bariloche (AR 1680)	Boeing 737 MAX • Club Economy • Equipaje de nieve/trekking sin cargo	Buenos Aires Aeroparque (AEP)	San Carlos de Bariloche (BRC)	Bariloche	AR	2026-11-23 07:00:00+00	2026-11-23 09:25:00+00	265	USD	14	4.84	https://images.unsplash.com/photo-1578894381163-e72c17f2d45f?auto=format&fit=crop&w=900&q=80	Equipaje Deportivo Incluido	AR	AR-1140	{"gate": "07", "seat": "05F", "terminal": "Nacional", "cabinClass": "Club Economy", "flightNumber": "AR 1680", "arrivalAirport": "BRC", "departureAirport": "AEP"}
off-flight-lat-5520	FLIGHT	LATAM Airlines	confirmaciones@latam.com	Vuelo Andino a Cusco Imperial (LA 2027)	Airbus A319 • Cabina Premium Economy • Embarque Preferente	Lima Jorge Chávez (LIM)	Cusco Velasco Astete (CUZ)	Cusco	PE	2026-11-24 07:00:00+00	2026-11-24 08:30:00+00	210	USD	8	4.9	https://images.unsplash.com/photo-1587595431973-160d0d94add1?auto=format&fit=crop&w=900&q=80	Vista Cordillera de los Andes	LAT	LAT-5520	{"gate": "12", "seat": "02A", "terminal": "Nacional", "cabinClass": "Premium Economy", "flightNumber": "LA 2027", "arrivalAirport": "CUZ", "departureAirport": "LIM"}
off-flight-ita-6290	FLIGHT	ITA Airways	booking@ita-airways.com	Vuelo Directo a Roma Fiumicino (AZ 061)	Airbus A321neo • Clase Superior • Gastronomía italiana a bordo	Madrid Barajas (MAD)	Roma Fiumicino (FCO)	Roma	IT	2026-11-29 08:00:00+00	2026-11-29 10:35:00+00	290	EUR	12	4.86	https://images.unsplash.com/photo-1529260830199-42c24126f198?auto=format&fit=crop&w=900&q=80	Nuevo Airbus A321neo	ITA	ITA-6290	{"gate": "B19", "seat": "06A", "terminal": "Terminal 1", "cabinClass": "Superior Classic", "flightNumber": "AZ 061", "arrivalAirport": "FCO", "departureAirport": "MAD"}
off-flight-af-7104	FLIGHT	Air France	confirmation@airfrance.fr	Vuelo Directo a París Charles de Gaulle (AF 007)	Boeing 777-300ER • Premium Economy • Champagne de bienvenida	Nueva York (JFK)	París Charles de Gaulle (CDG)	París	FR	2026-11-26 17:00:00+00	2026-11-27 00:20:00+00	980	EUR	10	4.91	https://images.unsplash.com/photo-1499856871958-5b9627545d1a?auto=format&fit=crop&w=900&q=80	SkyPriority Incluido	AF	AF-7104	{"gate": "K41", "seat": "11A", "terminal": "Terminal 2E", "cabinClass": "Premium Economy", "flightNumber": "AF 007", "arrivalAirport": "CDG", "departureAirport": "JFK"}
off-hotel-bkg-7741	HOTEL	Booking.com	reservations@booking.com	Hotel Gracery Shinjuku Tokio (5 Noches)	Pisos altos panorámicos en Shinjuku • A 4 min de JR Shinjuku Station	Shinjuku, Tokio	Hotel Gracery Shinjuku	Tokio	JP	2026-11-25 14:00:00+00	2026-11-30 10:00:00+00	940	USD	5	4.92	https://images.unsplash.com/photo-1542314831-068cd1dbfeeb?auto=format&fit=crop&w=900&q=80	Vista Skyline de Tokio	BKG	BKG-7741	{"address": "1-19-1 Kabukicho, Shinjuku-ku, Tokio 160-8466", "latitude": 35.6954, "roomType": "Superior Double Panorama", "hotelName": "Hotel Gracery Shinjuku", "longitude": 139.7021, "checkInTime": "15:00", "checkOutTime": "11:00"}
off-hotel-bkg-9012	HOTEL	Booking.com	reservations@booking.com	Hotel Riu Plaza España Madrid (4 Noches)	Edificio histórico en Gran Vía 84 • Piscina exterior y Rooftop 360°	Gran Vía, Madrid	Hotel Riu Plaza España	Madrid	ES	2026-11-27 14:00:00+00	2026-12-01 11:00:00+00	760	EUR	6	4.93	https://images.unsplash.com/photo-1566073771259-6a8506099945?auto=format&fit=crop&w=900&q=80	Desayuno Buffet & Sky Bar	BKG	BKG-9012	{"address": "Calle Gran Vía 84, Centro, 28013 Madrid", "latitude": 40.4239, "roomType": "Deluxe King Gran Vía", "hotelName": "Hotel Riu Plaza España", "longitude": -3.7112, "checkInTime": "15:00", "checkOutTime": "12:00"}
off-hotel-bkg-5508	HOTEL	Booking.com	reservations@booking.com	Llao Llao Resort, Golf & Spa Bariloche (4 Noches)	Península Llao Llao • Vista al Lago Nahuel Huapi y Cerro López	Bustillo Km 25, Bariloche	Llao Llao Resort, Golf & Spa	Bariloche	AR	2026-11-23 14:00:00+00	2026-11-27 10:00:00+00	1180	USD	4	4.97	https://images.unsplash.com/photo-1582719508461-905c673771fd?auto=format&fit=crop&w=900&q=80	Resort 5★ Patagonia	BKG	BKG-5508	{"address": "Av. Ezequiel Bustillo Km 25, R8401 San Carlos de Bariloche", "latitude": -41.0583, "roomType": "Suite Lago Moreno", "hotelName": "Llao Llao Resort, Golf & Spa", "longitude": -71.5332, "checkInTime": "15:00", "checkOutTime": "11:00"}
off-hotel-bkg-6315	HOTEL	Booking.com	reservations@booking.com	Pullman Paris Tour Eiffel (4 Noches)	A 300 metros de la Torre Eiffel y el Campo de Marte • Balcón Privado	Distrito 15, París	Pullman Paris Tour Eiffel	París	FR	2026-11-26 14:00:00+00	2026-11-30 11:00:00+00	1040	EUR	5	4.91	https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?auto=format&fit=crop&w=900&q=80	Vista Directa Torre Eiffel	BKG	BKG-6315	{"address": "18 Avenue de Suffren, 75015 París", "latitude": 48.8556, "roomType": "Deluxe Eiffel View Balcony", "hotelName": "Pullman Paris Tour Eiffel", "longitude": 2.2928, "checkInTime": "15:00", "checkOutTime": "12:00"}
off-train-jr-9914	TRAIN	JR Central Shinkansen	tickets@jrpass.com	Tren Bala Shinkansen Nozomi 21 (Tokio → Kioto)	2h 13m de alta velocidad • Asiento reservado ventanilla Monte Fuji	Tokio Station	Kioto Station	Tokio	JP	2026-11-28 08:00:00+00	2026-11-28 10:15:00+00	145	USD	18	4.98	https://images.unsplash.com/photo-1524413840807-0c3cb6fa808d?auto=format&fit=crop&w=900&q=80	Puntualidad Japonesa 99.9%	JR	JR-9914	{"seat": "12E", "wagon": "Coche 7", "trainLine": "Tokaido Shinkansen", "trainNumber": "Nozomi 21", "arrivalStation": "Kioto Central (Andén 12)", "departureStation": "Tokio Central (Andén 16)"}
off-train-ave-6102	TRAIN	Renfe AVE	billetes@renfe.com	Tren Renfe AVE Directo (Madrid → Barcelona)	2h 30m sin paradas • Coche Confort • Sala Club en Puerta de Atocha	Madrid Puerta de Atocha	Barcelona Sants	Madrid	ES	2026-11-30 08:30:00+00	2026-11-30 11:00:00+00	118	EUR	16	4.91	https://images.unsplash.com/photo-1474487548417-781cb71495f3?auto=format&fit=crop&w=900&q=80	Alta Velocidad 310 km/h	AVE	AVE-6102	{"seat": "08A", "wagon": "Coche 3", "trainLine": "AVE Alta Velocidad", "trainNumber": "AVE 03102", "arrivalStation": "Barcelona Sants", "departureStation": "Madrid Puerta de Atocha"}
off-train-eur-4410	TRAIN	Eurostar International	bookings@eurostar.com	Eurostar Bajo el Canal (París → Londres)	Cruce por el Eurotúnel en 2h 15m • Clase Eurostar Plus con comida ligera	París Gare du Nord	London St Pancras International	París	FR	2026-11-30 09:00:00+00	2026-11-30 11:15:00+00	185	EUR	14	4.93	https://images.unsplash.com/photo-1513635269975-59663e0ac1ad?auto=format&fit=crop&w=900&q=80	Centro a Centro en 2h 15m	EUR	EUR-4410	{"seat": "61", "wagon": "Coach 09", "trainLine": "Eurostar Channel Line", "trainNumber": "ES 9019", "arrivalStation": "London St Pancras Int.", "departureStation": "París Gare du Nord"}
off-train-pru-8045	TRAIN	PeruRail Vistadome	reservas@perurail.com	Tren Panorámico PeruRail Vistadome (Valle Sagrado → Machu Picchu)	Vagón acristalado 360° • Música y gastronomía andina a bordo	Ollantaytambo (Cusco)	Machu Picchu Pueblo	Cusco	PE	2026-11-27 06:30:00+00	2026-11-27 08:15:00+00	132	USD	12	4.96	https://images.unsplash.com/photo-1580619305218-8423a7ef79b4?auto=format&fit=crop&w=900&q=80	Cúpula Panorámica de Cristal	PRU	PRU-8045	{"seat": "15A", "wagon": "Coche B", "trainLine": "PeruRail Vistadome Observatory", "trainNumber": "VSD 303", "arrivalStation": "Estación Machu Picchu Pueblo", "departureStation": "Estación Ollantaytambo"}
off-rental-htz-3319	RENTAL	Hertz Rent a Car	confirmations@hertz.com	Alquiler SUV 4x4 Automática en Bariloche (5 Días)	Tracción integral AWD • Kilometraje ilimitado • Seguro total incluido	Aeropuerto Bariloche (BRC)	Aeropuerto Bariloche (BRC)	Bariloche	AR	2026-11-23 10:00:00+00	2026-11-28 10:00:00+00	340	USD	6	4.89	https://images.unsplash.com/photo-1533473359331-0135ef1b58bf?auto=format&fit=crop&w=900&q=80	Ideal Ruta de los 7 Lagos	HTZ	HTZ-3319	{"pickupLocation": "Mostrador Hertz Aeropuerto Bariloche (BRC)", "dropoffLocation": "Mostrador Hertz Aeropuerto Bariloche (BRC)", "vehicleCategory": "SUV 4x4 Automática AWD", "insuranceIncluded": true}
off-rental-avs-4190	RENTAL	Avis Prestige	reservations@avis.com	Híbrido Ejecutivo en Roma & Toscana (4 Días)	Acceso autorizado zonas ZTL • Navegador GPS europeo y telepeaje	Estación Roma Termini	Aeropuerto Roma Fiumicino (FCO)	Roma	IT	2026-11-29 09:00:00+00	2026-12-03 17:00:00+00	295	EUR	7	4.85	https://images.unsplash.com/photo-1549399542-7e3f8b79c341?auto=format&fit=crop&w=900&q=80	Cobertura Total Cero Franquicia	AVS	AVS-4190	{"pickupLocation": "Oficina Avis Roma Termini", "dropoffLocation": "Aeropuerto Roma Fiumicino (FCO)", "vehicleCategory": "Compacto Ejecutivo Híbrido", "insuranceIncluded": true}
off-bus-als-2088	BUS	ALSA Supra	billetes@alsa.es	Autobús ALSA Supra VIP (Madrid → Granada)	Butacas individuales de cuero • Wi-Fi de alta velocidad y pantallas táctiles	Madrid Estación Sur	Granada Estación Central	Madrid	ES	2026-11-29 07:30:00+00	2026-11-29 12:00:00+00	48	EUR	15	4.88	https://images.unsplash.com/photo-1544620347-c4fd4a3d5957?auto=format&fit=crop&w=900&q=80	Servicio Supra VIP	ALS	ALS-2088	{"seat": "03A", "platform": "Dársena 24", "busCompany": "ALSA Supra Economy", "lineNumber": "SUP-501", "arrivalTerminal": "Estación de Autobuses de Granada", "departureTerminal": "Madrid Estación Sur (Méndez Álvaro)"}
off-bus-vbr-1920	BUS	Vía Bariloche	pasajes@viabariloche.com.ar	Bus Panorámico Cama Suite (Bariloche → San Martín de los Andes)	Recorrido escénico por Villa La Angostura y Lagos • Butaca 180°	Terminal Bariloche	Terminal San Martín de los Andes	Bariloche	AR	2026-11-25 08:00:00+00	2026-11-25 12:15:00+00	42	USD	16	4.87	https://images.unsplash.com/photo-1469854523086-cc02fe5d8800?auto=format&fit=crop&w=900&q=80	Cama Suite Panorámico	VBR	VBR-1920	{"seat": "01C", "platform": "Plataforma 5", "busCompany": "Vía Bariloche Cama Suite", "lineNumber": "7L-102", "arrivalTerminal": "Terminal San Martín de los Andes (Lago Lácar)", "departureTerminal": "Terminal de Ómnibus de Bariloche"}
\.


--
-- Data for Name: catalog_packages; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.catalog_packages (id, slug, title, subtitle, destination_city, country_code, region, summary, image_url, badge, duration_days, price_per_person, currency, rating, reviews_count, includes_summary, available_dates, included_segments) FROM stdin;
pkg-japon-imperial	japon-imperial-tokio-hakone-kioto	Japón Imperial: Tokio, Hakone & Kioto	Vuelo directo ANA • Hotel en Shinjuku • Tren Bala Shinkansen Nozomi	Tokio	JP	Asia	Recorre la vanguardia de Tokio, contempla el Monte Fuji rumbo a Hakone y descubre los templos milenarios de Kioto a bordo del Shinkansen Nozomi con alojamiento céntrico.	https://images.unsplash.com/photo-1540959733332-eab4deabeeaf?auto=format&fit=crop&w=1200&q=85	Más elegido	10	2890	USD	4.95	184	["Vuelo internacional All Nippon Airways (Equipaje 23kg)", "Hotel Gracery Shinjuku 4★ con desayuno buffet", "Tren Bala Tokaido Shinkansen Nozomi con reserva de asiento"]	[{"id": "date-jp-1", "label": "Salida Próxima • Temporada Cerezos & Jardines", "endDate": "2026-12-04T23:00:00Z", "startDate": "2026-11-24T23:00:00Z", "priceModifier": 1, "spotsAvailable": 6}, {"id": "date-jp-2", "label": "Salida Especial • Festivales de Kioto", "endDate": "2026-12-18T23:00:00Z", "startDate": "2026-12-08T23:00:00Z", "priceModifier": 1.05, "spotsAvailable": 9}, {"id": "date-jp-3", "label": "Salida Flexible • Ruta Cultural Completa", "endDate": "2027-01-06T23:00:00Z", "startDate": "2026-12-27T23:00:00Z", "priceModifier": 0.96, "spotsAvailable": 12}]	[{"type": "FLIGHT", "notes": "Vuelo transpacífico directo • Equipaje facturado y selección de asiento incluidos", "origin": "Los Ángeles (LAX)", "details": {"gate": "114", "seat": "14A", "terminal": "Terminal 3", "cabinClass": "Premium Economy", "flightNumber": "NH 105", "arrivalAirport": "HND", "departureAirport": "LAX"}, "hourStart": 10, "priceShare": 0.48, "destination": "Tokio Haneda (HND)", "providerName": "All Nippon Airways", "durationHours": 12, "dayOffsetStart": 0, "confirmationPrefix": "ANA"}, {"type": "HOTEL", "notes": "Habitación Superior panorámica a pasos de la estación Shinjuku • Desayuno incluido", "origin": "Kabukicho, Shinjuku, Tokio", "details": {"address": "1-19-1 Kabukicho, Shinjuku-ku, Tokio 160-8466", "latitude": 35.6954, "roomType": "Superior Double Panorama", "hotelName": "Hotel Gracery Shinjuku", "longitude": 139.7021, "checkInTime": "15:00", "checkOutTime": "11:00"}, "hourStart": 15, "priceShare": 0.38, "destination": "Hotel Gracery Shinjuku", "providerName": "Booking.com", "durationHours": 216, "dayOffsetStart": 0, "confirmationPrefix": "BKG"}, {"type": "TRAIN", "notes": "Asiento reservado ventanilla E con vista al Monte Fuji en Hakone", "origin": "Tokio Station", "details": {"seat": "12E", "wagon": "Coche 7", "trainLine": "Tokaido Shinkansen", "trainNumber": "Nozomi 21", "arrivalStation": "Kioto Central (Andén 12)", "departureStation": "Tokio Central (Andén 16)"}, "hourStart": 9, "priceShare": 0.14, "destination": "Kioto Station", "providerName": "JR Central Shinkansen", "durationHours": 2, "dayOffsetStart": 4, "confirmationPrefix": "JR"}]
pkg-alta-velocidad-iberica	alta-velocidad-iberica-madrid-cordoba-barcelona	Alta Velocidad Ibérica: Madrid, Córdoba & Barcelona	Vuelo Iberia • Hotel Riu Plaza España • Tren Renfe AVE + Bus Costa Brava	Madrid	ES	Europa	Conecta la Gran Vía madrileña, el casco histórico de Córdoba y el modernismo de Barcelona viajando en tren de alta velocidad AVE y excursión costera por el Mediterráneo.	https://images.unsplash.com/photo-1539037116277-4db20889f2d4?auto=format&fit=crop&w=1200&q=85	Ruta en Tren AVE	8	1940	EUR	4.91	142	["Vuelo directo Iberia Clase Turista Confort", "Hotel Riu Plaza España 4★ en plena Gran Vía", "Tren Renfe AVE Madrid → Córdoba → Barcelona + Bus Costa Brava"]	[{"id": "date-es-1", "label": "Salida Próxima • Primavera Mediterránea", "endDate": "2026-12-04T23:00:00Z", "startDate": "2026-11-26T23:00:00Z", "priceModifier": 1, "spotsAvailable": 8}, {"id": "date-es-2", "label": "Salida Cultural • Arte & Gastronomía", "endDate": "2026-12-19T23:00:00Z", "startDate": "2026-12-11T23:00:00Z", "priceModifier": 1.03, "spotsAvailable": 10}, {"id": "date-es-3", "label": "Salida Especial • Sol y Terrazas", "endDate": "2027-01-08T23:00:00Z", "startDate": "2026-12-31T23:00:00Z", "priceModifier": 0.95, "spotsAvailable": 14}]	[{"type": "FLIGHT", "notes": "Vuelo directo transatlántico • Embarque grupo prioritario", "origin": "Buenos Aires (EZE)", "details": {"gate": "S28", "seat": "10K", "terminal": "T4S", "cabinClass": "Turista Premium", "flightNumber": "IB 6844", "arrivalAirport": "MAD", "departureAirport": "EZE"}, "hourStart": 13, "priceShare": 0.46, "destination": "Madrid Barajas (MAD)", "providerName": "Iberia", "durationHours": 12, "dayOffsetStart": 0, "confirmationPrefix": "IB"}, {"type": "HOTEL", "notes": "Acceso incluido al Sky Bar 360° • Desayuno buffet completo", "origin": "Gran Vía 84, Madrid", "details": {"address": "Calle Gran Vía 84, Centro, 28013 Madrid", "latitude": 40.4239, "roomType": "Deluxe King Gran Vía", "hotelName": "Hotel Riu Plaza España", "longitude": -3.7112, "checkInTime": "15:00", "checkOutTime": "12:00"}, "hourStart": 15, "priceShare": 0.36, "destination": "Hotel Riu Plaza España", "providerName": "Booking.com", "durationHours": 168, "dayOffsetStart": 0, "confirmationPrefix": "BKG"}, {"type": "TRAIN", "notes": "Alta velocidad 310 km/h • Coche Confort con Wi-Fi PlayRenfe", "origin": "Madrid Puerta de Atocha", "details": {"seat": "08A", "wagon": "Coche 3", "trainLine": "AVE Alta Velocidad", "trainNumber": "AVE 03102", "arrivalStation": "Barcelona Sants", "departureStation": "Madrid Puerta de Atocha"}, "hourStart": 9, "priceShare": 0.12, "destination": "Barcelona Sants (vía Córdoba)", "providerName": "Renfe AVE", "durationHours": 3, "dayOffsetStart": 3, "confirmationPrefix": "AVE"}, {"type": "BUS", "notes": "Autocar panorámico Supra con asientos de cuero reclinables", "origin": "Barcelona Estació del Nord", "details": {"seat": "04B", "platform": "Dársena 12", "busCompany": "ALSA Supra Mediterráneo", "lineNumber": "CBR-204", "arrivalTerminal": "Terminal Tossa de Mar", "departureTerminal": "Barcelona Estació del Nord"}, "hourStart": 10, "priceShare": 0.06, "destination": "Tossa de Mar / Costa Brava", "providerName": "ALSA Supra", "durationHours": 2, "dayOffsetStart": 5, "confirmationPrefix": "ALS"}]
pkg-patagonia-andina	patagonia-andina-bariloche-ruta-7-lagos	Patagonia Andina: Bariloche & Ruta de los 7 Lagos	Vuelo Aerolíneas Argentinas • Llao Llao Resort • Alquiler SUV Hertz 4x4	Bariloche	AR	Sudamérica	Vive la naturaleza majestuosa del Parque Nacional Nahuel Huapi hospedándote frente al lago y recorriendo a tu ritmo Villa La Angostura y San Martín de los Andes en SUV 4x4.	https://images.unsplash.com/photo-1518709268805-4e9042af9f23?auto=format&fit=crop&w=1200&q=85	Naturaleza & Fly-Drive	6	1480	USD	4.94	128	["Vuelo Aerolíneas Argentinas directo a Bariloche (BRC)", "5 noches en Llao Llao Resort, Golf & Spa 5★", "Alquiler SUV Hertz 4x4 con seguro contra todo riesgo"]	[{"id": "date-ar-1", "label": "Salida Próxima • Lagos Cristalina & Bosque", "endDate": "2026-11-28T23:00:00Z", "startDate": "2026-11-22T23:00:00Z", "priceModifier": 1, "spotsAvailable": 5}, {"id": "date-ar-2", "label": "Salida Aventura • Cumbres Patagónicas", "endDate": "2026-12-11T23:00:00Z", "startDate": "2026-12-05T23:00:00Z", "priceModifier": 1.04, "spotsAvailable": 8}, {"id": "date-ar-3", "label": "Salida Relax • Spa & Circuito Chico", "endDate": "2026-12-29T23:00:00Z", "startDate": "2026-12-23T23:00:00Z", "priceModifier": 0.94, "spotsAvailable": 10}]	[{"type": "FLIGHT", "notes": "Cabina Club Economy • Equipaje de bodega incluido", "origin": "Buenos Aires Aeroparque (AEP)", "details": {"gate": "07", "seat": "04F", "terminal": "Partidas Nacionales", "cabinClass": "Club Economy", "flightNumber": "AR 1680", "arrivalAirport": "BRC", "departureAirport": "AEP"}, "hourStart": 8, "priceShare": 0.24, "destination": "San Carlos de Bariloche (BRC)", "providerName": "Aerolíneas Argentinas", "durationHours": 2, "dayOffsetStart": 0, "confirmationPrefix": "AR"}, {"type": "HOTEL", "notes": "Suite Lago Moreno con acceso ilimitado a piscina climatizada in-out y spa", "origin": "Av. Ezequiel Bustillo Km 25, Bariloche", "details": {"address": "Av. Ezequiel Bustillo Km 25, R8401 San Carlos de Bariloche", "latitude": -41.0583, "roomType": "Suite Lago Nahuel Huapi", "hotelName": "Llao Llao Resort, Golf & Spa", "longitude": -71.5332, "checkInTime": "15:00", "checkOutTime": "11:00"}, "hourStart": 15, "priceShare": 0.54, "destination": "Llao Llao Resort, Golf & Spa", "providerName": "Booking.com", "durationHours": 120, "dayOffsetStart": 0, "confirmationPrefix": "BKG"}, {"type": "RENTAL", "notes": "SUV 4x4 Automática con kilometraje libre para la Ruta de los 7 Lagos", "origin": "Aeropuerto Bariloche (BRC)", "details": {"pickupLocation": "Mostrador Hertz Aeropuerto Bariloche (BRC)", "dropoffLocation": "Mostrador Hertz Aeropuerto Bariloche (BRC)", "vehicleCategory": "SUV 4x4 Automática AWD", "insuranceIncluded": true}, "hourStart": 11, "priceShare": 0.22, "destination": "Aeropuerto Bariloche (BRC)", "providerName": "Hertz Rent a Car", "durationHours": 120, "dayOffsetStart": 0, "confirmationPrefix": "HTZ"}]
pkg-roma-costa-amalfitana	roma-clasica-florencia-costa-amalfitana	Roma Clásica, Florencia & Costa Amalfitana	Vuelo ITA Airways • Hotel Artemide Roma • Tren Frecciarossa + Bus Sorrento	Roma	IT	Europa	Descubre el Coliseo y el Vaticano, cruza la Toscana en tren de alta velocidad Frecciarossa hacia Florencia y contempla los acantilados de Sorrento y Positano.	https://images.unsplash.com/photo-1552832230-c0197dd311b5?auto=format&fit=crop&w=1200&q=85	Clásico Italiano	9	2290	EUR	4.92	165	["Vuelo directo ITA Airways a Roma Fiumicino (FCO)", "Hotel Artemide Roma 4★ Superior en Via Nazionale", "Tren Frecciarossa 1000 + Traslado panorámico a Sorrento"]	[{"id": "date-it-1", "label": "Salida Próxima • Dolce Vita & Toscana", "endDate": "2026-12-07T23:00:00Z", "startDate": "2026-11-28T23:00:00Z", "priceModifier": 1, "spotsAvailable": 7}, {"id": "date-it-2", "label": "Salida Especial • Costa Amalfitana Soleada", "endDate": "2026-12-22T23:00:00Z", "startDate": "2026-12-13T23:00:00Z", "priceModifier": 1.05, "spotsAvailable": 9}, {"id": "date-it-3", "label": "Salida Arte & Historia • Museos sin Filas", "endDate": "2027-01-11T23:00:00Z", "startDate": "2027-01-02T23:00:00Z", "priceModifier": 0.96, "spotsAvailable": 11}]	[{"type": "FLIGHT", "notes": "Vuelo directo • Franquicia de equipaje en cabina y bodega", "origin": "Madrid Barajas (MAD)", "details": {"gate": "B19", "seat": "07A", "terminal": "Terminal 1", "cabinClass": "Superior Classic", "flightNumber": "AZ 061", "arrivalAirport": "FCO", "departureAirport": "MAD"}, "hourStart": 9, "priceShare": 0.34, "destination": "Roma Fiumicino (FCO)", "providerName": "ITA Airways", "durationHours": 3, "dayOffsetStart": 0, "confirmationPrefix": "ITA"}, {"type": "HOTEL", "notes": "Minibar de cortesía diario • Terraza Ambrosia con vista a los tejados de Roma", "origin": "Via Nazionale 22, Roma", "details": {"address": "Via Nazionale 22, Repubblica, 00184 Roma RM", "latitude": 41.9016, "roomType": "Executive King Room", "hotelName": "Hotel Artemide Roma", "longitude": 12.4939, "checkInTime": "14:00", "checkOutTime": "11:00"}, "hourStart": 14, "priceShare": 0.48, "destination": "Hotel Artemide Roma", "providerName": "Booking.com", "durationHours": 192, "dayOffsetStart": 0, "confirmationPrefix": "BKG"}, {"type": "TRAIN", "notes": "Tren de alta velocidad Frecciarossa 1000 • Clase Business Silencio", "origin": "Roma Termini", "details": {"seat": "05D", "wagon": "Carrozza 2", "trainLine": "Frecciarossa 1000", "trainNumber": "FR 9514", "arrivalStation": "Firenze S.M. Novella", "departureStation": "Roma Termini"}, "hourStart": 8, "priceShare": 0.11, "destination": "Firenze Santa Maria Novella", "providerName": "Trenitalia Frecciarossa", "durationHours": 2, "dayOffsetStart": 3, "confirmationPrefix": "FRC"}, {"type": "BUS", "notes": "Recorrido escénico por el Golfo de Nápoles y miradores de la Costa Amalfitana", "origin": "Roma Tiburtina", "details": {"seat": "02A", "platform": "Andén 6", "busCompany": "Marozzi Costiera Gran Turismo", "lineNumber": "AMF-108", "arrivalTerminal": "Piazza Tasso, Sorrento", "departureTerminal": "Estación Roma Tiburtina"}, "hourStart": 7, "priceShare": 0.07, "destination": "Sorrento & Positano Mirador", "providerName": "Marozzi Costiera Express", "durationHours": 4, "dayOffsetStart": 6, "confirmationPrefix": "MRZ"}]
pkg-cusco-machu-picchu	cusco-valle-sagrado-machu-picchu	Cusco, Valle Sagrado & Machu Picchu	Vuelo LATAM • Hotel Palacio del Inka • Tren Panorámico PeruRail Vistadome	Cusco	PE	Sudamérica	Adéntrate en el corazón del Imperio Inca entre las calles empedradas de Cusco, las terrazas de Ollantaytambo y el tren con cúpula de cristal hacia la ciudadela de Machu Picchu.	https://images.unsplash.com/photo-1526392060635-9d6019884377?auto=format&fit=crop&w=1200&q=85	Maravilla del Mundo	7	1720	USD	4.96	210	["Vuelo LATAM Airlines directo a Cusco (CUZ)", "Palacio del Inka, a Luxury Collection Hotel 5★", "Tren panorámico PeruRail Vistadome a Aguas Calientes / Machu Picchu"]	[{"id": "date-pe-1", "label": "Salida Próxima • Cielos Despejados Andinos", "endDate": "2026-11-30T23:00:00Z", "startDate": "2026-11-23T23:00:00Z", "priceModifier": 1, "spotsAvailable": 6}, {"id": "date-pe-2", "label": "Salida Especial • Ruta del Sol & Valle Sagrado", "endDate": "2026-12-13T23:00:00Z", "startDate": "2026-12-06T23:00:00Z", "priceModifier": 1.04, "spotsAvailable": 9}, {"id": "date-pe-3", "label": "Salida Expedición • Ingreso Matutino Ciudadela", "endDate": "2027-01-01T23:00:00Z", "startDate": "2026-12-25T23:00:00Z", "priceModifier": 0.95, "spotsAvailable": 11}]	[{"type": "FLIGHT", "notes": "Cabina Premium Economy • Ventanilla lado cordillera", "origin": "Lima Jorge Chávez (LIM)", "details": {"gate": "12", "seat": "02A", "terminal": "Nacional", "cabinClass": "Premium Economy", "flightNumber": "LA 2027", "arrivalAirport": "CUZ", "departureAirport": "LIM"}, "hourStart": 8, "priceShare": 0.26, "destination": "Cusco Alejandro Velasco Astete (CUZ)", "providerName": "LATAM Airlines", "durationHours": 2, "dayOffsetStart": 0, "confirmationPrefix": "LAT"}, {"type": "HOTEL", "notes": "Casona histórica frente al Qorikancha • Mate de coca de bienvenida y oxigenación", "origin": "Plazoleta Santo Domingo 259, Cusco", "details": {"address": "Plazoleta Santo Domingo 259, Centro Histórico, Cusco 08002", "latitude": -13.5204, "roomType": "Colonial Deluxe Suite", "hotelName": "Palacio del Inka, a Luxury Collection Hotel", "longitude": -71.9754, "checkInTime": "14:00", "checkOutTime": "12:00"}, "hourStart": 14, "priceShare": 0.52, "destination": "Palacio del Inka Luxury Collection", "providerName": "Booking.com", "durationHours": 144, "dayOffsetStart": 0, "confirmationPrefix": "BKG"}, {"type": "TRAIN", "notes": "Vagón panorámico con ventanas trapezoidales en el techo y snack andino a bordo", "origin": "Ollantaytambo (Valle Sagrado)", "details": {"seat": "15A", "wagon": "Coche B", "trainLine": "PeruRail Vistadome Observatory", "trainNumber": "VSD 303", "arrivalStation": "Estación Machu Picchu Pueblo", "departureStation": "Estación Ollantaytambo"}, "hourStart": 7, "priceShare": 0.22, "destination": "Machu Picchu Pueblo (Aguas Calientes)", "providerName": "PeruRail Vistadome", "durationHours": 2, "dayOffsetStart": 3, "confirmationPrefix": "PRU"}]
pkg-paris-londres-eurostar	paris-londres-por-el-eurotunel	París & Londres por el Eurotúnel	Vuelo Air France • Pullman Paris Tour Eiffel • Tren Eurostar a London St Pancras	París	FR	Europa	Dos capitales icónicas en un solo itinerario: despierta frente a la Torre Eiffel, navega por el Sena y cruza el Canal de la Mancha en 2h 15m a bordo del tren Eurostar hasta el corazón de Londres.	https://images.unsplash.com/photo-1502602898657-3e91760cbb34?auto=format&fit=crop&w=1200&q=85	Dos Capitales	8	2460	EUR	4.93	156	["Vuelo internacional Air France a París Charles de Gaulle (CDG)", "Hotel Pullman Paris Tour Eiffel 4★ a los pies de la Torre", "Tren de alta velocidad Eurostar París Gare du Nord → London St Pancras"]	[{"id": "date-fr-1", "label": "Salida Próxima • Luces del Sena & Támesis", "endDate": "2026-12-03T23:00:00Z", "startDate": "2026-11-25T23:00:00Z", "priceModifier": 1, "spotsAvailable": 6}, {"id": "date-fr-2", "label": "Salida Especial • Museos & West End", "endDate": "2026-12-17T23:00:00Z", "startDate": "2026-12-09T23:00:00Z", "priceModifier": 1.04, "spotsAvailable": 8}, {"id": "date-fr-3", "label": "Salida Romántica • Bulevares & Jardines Reales", "endDate": "2027-01-06T23:00:00Z", "startDate": "2026-12-29T23:00:00Z", "priceModifier": 0.96, "spotsAvailable": 12}]	[{"type": "FLIGHT", "notes": "Servicio francés a bordo • Acceso prioritario SkyPriority", "origin": "Nueva York (JFK)", "details": {"gate": "K41", "seat": "11A", "terminal": "Terminal 2E", "cabinClass": "Premium Economy", "flightNumber": "AF 007", "arrivalAirport": "CDG", "departureAirport": "JFK"}, "hourStart": 18, "priceShare": 0.42, "destination": "París Charles de Gaulle (CDG)", "providerName": "Air France", "durationHours": 7, "dayOffsetStart": 0, "confirmationPrefix": "AF"}, {"type": "HOTEL", "notes": "Habitación con balcón y vista directa a la Torre Eiffel iluminada", "origin": "18 Avenue de Suffren, 75015 París", "details": {"address": "18 Avenue de Suffren, Entrée au 22 rue Jean Rey, 75015 París", "latitude": 48.8556, "roomType": "Deluxe Eiffel View Balcony", "hotelName": "Pullman Paris Tour Eiffel", "longitude": 2.2928, "checkInTime": "15:00", "checkOutTime": "12:00"}, "hourStart": 15, "priceShare": 0.44, "destination": "Pullman Paris Tour Eiffel", "providerName": "Booking.com", "durationHours": 168, "dayOffsetStart": 0, "confirmationPrefix": "BKG"}, {"type": "TRAIN", "notes": "Cruce de alta velocidad bajo el Canal de la Mancha en Eurostar Plus", "origin": "París Gare du Nord", "details": {"seat": "61", "wagon": "Coach 09", "trainLine": "Eurostar Channel Line", "trainNumber": "ES 9019", "arrivalStation": "London St Pancras Int.", "departureStation": "París Gare du Nord"}, "hourStart": 10, "priceShare": 0.14, "destination": "London St Pancras International", "providerName": "Eurostar International", "durationHours": 2, "dayOffsetStart": 4, "confirmationPrefix": "EUR"}]
\.


--
-- Data for Name: email_confirmations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.email_confirmations (id, user_id, from_email, provider_name, subject, body, pnr, type, origin, destination, destination_city, country_code, start_date, end_date, price, currency, details, status, status_reason, imported_trip_id, received_at) FROM stdin;
eml-135012d000fdce8e	usr-28802ee52e6b868c	pasajes@viabariloche.com.ar	Vía Bariloche	Confirmación Autobús — Bus Panorámico Cama Suite (Bariloche → San Martín de los Andes) (VBR-1920)	Reserva confirmada en el catálogo Trotamundos. Código PNR: VBR-1920. Terminal Bariloche → Terminal San Martín de los Andes.	VBR-1920	BUS	Terminal Bariloche	Terminal San Martín de los Andes	Bariloche	AR	2026-11-25 08:00:00+00	2026-11-25 12:15:00+00	42	USD	{"seat": "01C", "platform": "Plataforma 5", "busCompany": "Vía Bariloche Cama Suite", "lineNumber": "7L-102", "arrivalTerminal": "Terminal San Martín de los Andes (Lago Lácar)", "departureTerminal": "Terminal de Ómnibus de Bariloche"}	IMPORTED	Agregada al viaje "Viaje a Terminal San Martín de los Andes"	trip-aeaf1492eceae331	2026-11-17 23:56:58.499145+00
eml-8349ee99a3f71e2b	usr-ed97b9ed908373ef	no-reply@latam.com	LATAM Airlines	Confirmación de reserva Vuelo — LATAM Airlines (LAT-93101)	Tu reserva del paquete "Cusco, Valle Sagrado & Machu Picchu" está confirmada. Código PNR: LAT-93101. Trayecto/Servicio: Lima Jorge Chávez (LIM) → Cusco Alejandro Velasco Astete (CUZ).	LAT-93101	FLIGHT	Lima Jorge Chávez (LIM)	Cusco Alejandro Velasco Astete (CUZ)	Cusco	PE	2026-11-24 07:00:00+00	2026-11-24 09:00:00+00	447.2	USD	{"gate": "12", "seat": "02A", "terminal": "Nacional", "cabinClass": "Premium Economy", "flightNumber": "LA 2027", "arrivalAirport": "CUZ", "departureAirport": "LIM"}	IMPORTED	Incorporada al reservar el paquete "Cusco, Valle Sagrado & Machu Picchu"	trip-714723e7ff2894ef	2026-11-17 23:58:51.361531+00
eml-71d60d80ce21e72b	usr-ed97b9ed908373ef	noreply@booking.com	Booking.com	Confirmación de reserva Alojamiento — Booking.com (BKG-93102)	Tu reserva del paquete "Cusco, Valle Sagrado & Machu Picchu" está confirmada. Código PNR: BKG-93102. Trayecto/Servicio: Plazoleta Santo Domingo 259, Cusco → Palacio del Inka Luxury Collection.	BKG-93102	HOTEL	Plazoleta Santo Domingo 259, Cusco	Palacio del Inka Luxury Collection	Cusco	PE	2026-11-24 13:00:00+00	2026-11-30 13:00:00+00	894.4	USD	{"address": "Plazoleta Santo Domingo 259, Centro Histórico, Cusco 08002", "latitude": -13.5204, "roomType": "Colonial Deluxe Suite", "hotelName": "Palacio del Inka, a Luxury Collection Hotel", "longitude": -71.9754, "checkInTime": "14:00", "checkOutTime": "12:00"}	IMPORTED	Incorporada al reservar el paquete "Cusco, Valle Sagrado & Machu Picchu"	trip-714723e7ff2894ef	2026-11-17 23:58:51.361531+00
eml-4d68f341249f6c0e	usr-ed97b9ed908373ef	tickets@jrpass.com	PeruRail Vistadome	Confirmación de reserva Tren — PeruRail Vistadome (PRU-93103)	Tu reserva del paquete "Cusco, Valle Sagrado & Machu Picchu" está confirmada. Código PNR: PRU-93103. Trayecto/Servicio: Ollantaytambo (Valle Sagrado) → Machu Picchu Pueblo (Aguas Calientes).	PRU-93103	TRAIN	Ollantaytambo (Valle Sagrado)	Machu Picchu Pueblo (Aguas Calientes)	Cusco	PE	2026-11-27 06:00:00+00	2026-11-27 08:00:00+00	378.4	USD	{"seat": "15A", "wagon": "Coche B", "trainLine": "PeruRail Vistadome Observatory", "trainNumber": "VSD 303", "arrivalStation": "Estación Machu Picchu Pueblo", "departureStation": "Estación Ollantaytambo"}	IMPORTED	Incorporada al reservar el paquete "Cusco, Valle Sagrado & Machu Picchu"	trip-714723e7ff2894ef	2026-11-17 23:58:51.361531+00
eml-8ef20a24757aa569	usr-96b7e95209c554b6	tickets@aerolineas.com.ar	Aerolíneas Argentinas	Comprobante PNR AR-1140 — Aerolíneas Argentinas (Vuelo a San Carlos de Bariloche (AR 1680))	Reserva recuperada por código PNR AR-1140. Servicio: Buenos Aires Aeroparque (AEP) → San Carlos de Bariloche (BRC).	AR-1140	FLIGHT	Buenos Aires Aeroparque (AEP)	San Carlos de Bariloche (BRC)	Bariloche	AR	2026-11-23 07:00:00+00	2026-11-23 09:25:00+00	265	USD	{"gate": "07", "seat": "05F", "terminal": "Nacional", "cabinClass": "Club Economy", "flightNumber": "AR 1680", "arrivalAirport": "BRC", "departureAirport": "AEP"}	IMPORTED	Importada por código PNR en "Viaje a Bariloche"	trip-db99b55be3a82fe3	2026-11-21 18:39:31.514004+00
eml-845f5c9cc9509b5a	usr-96b7e95209c554b6	tickets@aerolineas.com.ar	Aerolíneas Argentinas	Comprobante PNR AR-1140 — Aerolíneas Argentinas (Vuelo a San Carlos de Bariloche (AR 1680))	Reserva recuperada por código PNR AR-1140. Servicio: Buenos Aires Aeroparque (AEP) → San Carlos de Bariloche (BRC).	AR-1140	FLIGHT	Buenos Aires Aeroparque (AEP)	San Carlos de Bariloche (BRC)	Bariloche	AR	2026-11-23 07:00:00+00	2026-11-23 09:25:00+00	265	USD	{"gate": "07", "seat": "05F", "terminal": "Nacional", "cabinClass": "Club Economy", "flightNumber": "AR 1680", "arrivalAirport": "BRC", "departureAirport": "AEP"}	IMPORTED	Importada por código PNR en "Viaje a Bariloche"	trip-db99b55be3a82fe3	2026-11-21 18:39:43.941441+00
eml-7c170c8df9e11358	usr-96b7e95209c554b6	reservations@booking.com	Booking.com	Confirmación Alojamiento — Pullman Paris Tour Eiffel (4 Noches) (BKG-6315)	Reserva confirmada en el catálogo Trotamundos. Código PNR: BKG-6315. Distrito 15, París → Pullman Paris Tour Eiffel.	BKG-6315	HOTEL	Distrito 15, París	Pullman Paris Tour Eiffel	París	FR	2026-11-26 14:00:00+00	2026-11-30 11:00:00+00	1040	EUR	{"address": "18 Avenue de Suffren, 75015 París", "latitude": 48.8556, "roomType": "Deluxe Eiffel View Balcony", "hotelName": "Pullman Paris Tour Eiffel", "longitude": 2.2928, "checkInTime": "15:00", "checkOutTime": "12:00"}	IMPORTED	Agregada al viaje "Viaje a París"	trip-16c90494f7187db4	2026-11-21 19:09:40.561102+00
eml-a3f4ed9e9d1979b8	usr-96b7e95209c554b6	reservations@booking.com	Booking.com	Confirmación Alojamiento — Llao Llao Resort, Golf & Spa Bariloche (4 Noches) (BKG-5508)	Reserva confirmada en el catálogo Trotamundos. Código PNR: BKG-5508. Bustillo Km 25, Bariloche → Llao Llao Resort, Golf & Spa.	BKG-5508	HOTEL	Bustillo Km 25, Bariloche	Llao Llao Resort, Golf & Spa	Bariloche	AR	2026-11-23 14:00:00+00	2026-11-27 10:00:00+00	1180	USD	{"address": "Av. Ezequiel Bustillo Km 25, R8401 San Carlos de Bariloche", "latitude": -41.0583, "roomType": "Suite Lago Moreno", "hotelName": "Llao Llao Resort, Golf & Spa", "longitude": -71.5332, "checkInTime": "15:00", "checkOutTime": "11:00"}	IMPORTED	Agregada al viaje "Viaje a París"	trip-16c90494f7187db4	2026-11-21 19:10:16.610088+00
eml-364fe761b6f4b373	usr-96b7e95209c554b6	reservations@booking.com	Booking.com	Confirmación Alojamiento — Hotel Riu Plaza España Madrid (4 Noches) (BKG-9012)	Reserva confirmada en el catálogo Trotamundos. Código PNR: BKG-9012. Gran Vía, Madrid → Hotel Riu Plaza España.	BKG-9012	HOTEL	Gran Vía, Madrid	Hotel Riu Plaza España	Madrid	ES	2026-11-27 14:00:00+00	2026-12-01 11:00:00+00	760	EUR	{"address": "Calle Gran Vía 84, Centro, 28013 Madrid", "latitude": 40.4239, "roomType": "Deluxe King Gran Vía", "hotelName": "Hotel Riu Plaza España", "longitude": -3.7112, "checkInTime": "15:00", "checkOutTime": "12:00"}	IMPORTED	Agregada al viaje "Viaje a París"	trip-16c90494f7187db4	2026-11-21 19:10:53.41093+00
eml-937c5639ef4e0aeb	usr-96b7e95209c554b6	reservations@booking.com	Booking.com	Confirmación Alojamiento — Llao Llao Resort, Golf & Spa Bariloche (4 Noches) (BKG-5508)	Reserva confirmada en el catálogo Trotamundos. Código PNR: BKG-5508. Bustillo Km 25, Bariloche → Llao Llao Resort, Golf & Spa.	BKG-5508	HOTEL	Bustillo Km 25, Bariloche	Llao Llao Resort, Golf & Spa	Bariloche	AR	2026-11-23 14:00:00+00	2026-11-27 10:00:00+00	1180	USD	{"address": "Av. Ezequiel Bustillo Km 25, R8401 San Carlos de Bariloche", "latitude": -41.0583, "roomType": "Suite Lago Moreno", "hotelName": "Llao Llao Resort, Golf & Spa", "longitude": -71.5332, "checkInTime": "15:00", "checkOutTime": "11:00"}	IMPORTED	Agregada al viaje "Viaje a París"	trip-16c90494f7187db4	2026-11-21 19:11:03.658173+00
eml-cd8f02cea97da397	usr-c7ae6c1824782002	tickets@aerolineas.com.ar	Aerolíneas Argentinas	Confirmación de Vuelo: Vuelo a San Carlos de Bariloche (AR 1680) (PNR: AR-1140)	Hemos emitido tu comprobante con Aerolíneas Argentinas. Localizador PNR: AR-1140. Servicio: Buenos Aires Aeroparque (AEP) → San Carlos de Bariloche (BRC).	AR-1140	FLIGHT	Buenos Aires Aeroparque (AEP)	San Carlos de Bariloche (BRC)	Bariloche	AR	2026-11-23 07:00:00+00	2026-11-23 09:25:00+00	265	USD	{"gate": "07", "seat": "05F", "terminal": "Nacional", "cabinClass": "Club Economy", "flightNumber": "AR 1680", "arrivalAirport": "BRC", "departureAirport": "AEP"}	IMPORTED	Sincronizada en "Viaje a Bariloche"	trip-9aeda99c6c53875b	2026-11-21 19:23:06.133515+00
eml-5b470480619b3adc	usr-c7ae6c1824782002	confirmaciones@ana.co.jp	All Nippon Airways	Comprobante de Vuelo — All Nippon Airways (PNR: ANA-62901)	Confirmación emitida para Japón Imperial: Tokio, Hakone & Kioto. Código PNR: ANA-62901. Origen: Los Ángeles (LAX) — Destino: Tokio Haneda (HND).	ANA-62901	FLIGHT	Los Ángeles (LAX)	Tokio Haneda (HND)	Tokio	JP	2026-11-25 09:00:00+00	2026-11-25 21:00:00+00	1387.2	USD	{"gate": "114", "seat": "14A", "terminal": "Terminal 3", "cabinClass": "Premium Economy", "flightNumber": "NH 105", "arrivalAirport": "HND", "departureAirport": "LAX"}	IMPORTED	Sincronizada en "Viaje a Tokio"	trip-857a1e7f07e1b2ae	2026-11-21 19:23:49.52879+00
eml-4261534476565586	usr-c7ae6c1824782002	noreply@booking.com	Booking.com	Comprobante de Alojamiento — Booking.com (PNR: BKG-62902)	Confirmación emitida para Japón Imperial: Tokio, Hakone & Kioto. Código PNR: BKG-62902. Origen: Kabukicho, Shinjuku, Tokio — Destino: Hotel Gracery Shinjuku.	BKG-62902	HOTEL	Kabukicho, Shinjuku, Tokio	Hotel Gracery Shinjuku	Tokio	JP	2026-11-25 14:00:00+00	2026-12-04 14:00:00+00	1098.2	USD	{"address": "1-19-1 Kabukicho, Shinjuku-ku, Tokio 160-8466", "latitude": 35.6954, "roomType": "Superior Double Panorama", "hotelName": "Hotel Gracery Shinjuku", "longitude": 139.7021, "checkInTime": "15:00", "checkOutTime": "11:00"}	IMPORTED	Sincronizada en "Viaje a Tokio"	trip-857a1e7f07e1b2ae	2026-11-21 19:23:49.52879+00
eml-644a530218cd58bb	usr-c7ae6c1824782002	reservas@jrpass.com	JR Central Shinkansen	Comprobante de Tren — JR Central Shinkansen (PNR: JR-62903)	Confirmación emitida para Japón Imperial: Tokio, Hakone & Kioto. Código PNR: JR-62903. Origen: Tokio Station — Destino: Kioto Station.	JR-62903	TRAIN	Tokio Station	Kioto Station	Tokio	JP	2026-11-29 08:00:00+00	2026-11-29 10:00:00+00	404.6	USD	{"seat": "12E", "wagon": "Coche 7", "trainLine": "Tokaido Shinkansen", "trainNumber": "Nozomi 21", "arrivalStation": "Kioto Central (Andén 12)", "departureStation": "Tokio Central (Andén 16)"}	IMPORTED	Sincronizada en "Viaje a Tokio"	trip-857a1e7f07e1b2ae	2026-11-21 19:23:49.52879+00
eml-7702c5c03fe92015	usr-c7ae6c1824782002	reservations@booking.com	Booking.com	Confirmación de Alojamiento: Llao Llao Resort, Golf & Spa Bariloche (4 Noches) (PNR: BKG-5508)	Hemos emitido tu comprobante con Booking.com. Localizador PNR: BKG-5508. Servicio: Bustillo Km 25, Bariloche → Llao Llao Resort, Golf & Spa.	BKG-5508	HOTEL	Bustillo Km 25, Bariloche	Llao Llao Resort, Golf & Spa	Bariloche	AR	2026-11-23 14:00:00+00	2026-11-27 10:00:00+00	1180	USD	{"address": "Av. Ezequiel Bustillo Km 25, R8401 San Carlos de Bariloche", "latitude": -41.0583, "roomType": "Suite Lago Moreno", "hotelName": "Llao Llao Resort, Golf & Spa", "longitude": -71.5332, "checkInTime": "15:00", "checkOutTime": "11:00"}	IMPORTED	Sincronizada en "Viaje a Bariloche"	trip-9aeda99c6c53875b	2026-11-21 19:23:24.372354+00
eml-25fb5574bf05658f	usr-96b7e95209c554b6	reservations@booking.com	Booking.com	Confirmación de Alojamiento: Llao Llao Resort, Golf & Spa Bariloche (4 Noches) (PNR: BKG-5508)	Hemos emitido tu comprobante con Booking.com. Localizador PNR: BKG-5508. Servicio: Bustillo Km 25, Bariloche → Llao Llao Resort, Golf & Spa.	BKG-5508	HOTEL	Bustillo Km 25, Bariloche	Llao Llao Resort, Golf & Spa	Bariloche	AR	2026-11-23 14:00:00+00	2026-11-27 10:00:00+00	1180	USD	{"address": "Av. Ezequiel Bustillo Km 25, R8401 San Carlos de Bariloche", "latitude": -41.0583, "roomType": "Suite Lago Moreno", "hotelName": "Llao Llao Resort, Golf & Spa", "longitude": -71.5332, "checkInTime": "15:00", "checkOutTime": "11:00"}	IMPORTED	Sincronizada en "Viaje America"	trip-247691b21ac4d5fa	2026-11-21 19:28:11.89706+00
\.


--
-- Data for Name: pending_syncs; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.pending_syncs (id, user_id, source, provider_name, summary, raw_payload, error_reason, attempts, status, next_retry_at, created_at) FROM stdin;
\.


--
-- Data for Name: reservations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.reservations (id, trip_id, type, status, provider_name, confirmation_code, origin, destination, start_date, end_date, price, currency, notes, source, manually_modified, modified_fields, details, created_at, updated_at) FROM stdin;
res-5eec4f7372438873	trip-aeaf1492eceae331	BUS	CONFIRMED	Vía Bariloche	VBR-1920	Terminal Bariloche	Terminal San Martín de los Andes	2026-11-25 08:00:00+00	2026-11-25 12:15:00+00	42	USD	Bus Panorámico Cama Suite (Bariloche → San Martín de los Andes)	CATALOG	f	null	{"seat": "01C", "platform": "Plataforma 5", "busCompany": "Vía Bariloche Cama Suite", "lineNumber": "7L-102", "arrivalTerminal": "Terminal San Martín de los Andes (Lago Lácar)", "departureTerminal": "Terminal de Ómnibus de Bariloche"}	2026-11-17 23:56:58.496834+00	2026-11-17 23:57:09.577639+00
res-3eb16658621ed71e	trip-714723e7ff2894ef	FLIGHT	CONFIRMED	LATAM Airlines	LAT-93101	Lima Jorge Chávez (LIM)	Cusco Alejandro Velasco Astete (CUZ)	2026-11-24 07:00:00+00	2026-11-24 09:00:00+00	447.2	USD	Cabina Premium Economy • Ventanilla lado cordillera	CATALOG	f	null	{"gate": "12", "seat": "02A", "terminal": "Nacional", "cabinClass": "Premium Economy", "flightNumber": "LA 2027", "arrivalAirport": "CUZ", "departureAirport": "LIM"}	2026-11-17 23:58:51.361531+00	2026-11-17 23:58:51.361531+00
res-cbf7dd4650e30474	trip-714723e7ff2894ef	HOTEL	CONFIRMED	Booking.com	BKG-93102	Plazoleta Santo Domingo 259, Cusco	Palacio del Inka Luxury Collection	2026-11-24 13:00:00+00	2026-11-30 13:00:00+00	894.4	USD	Casona histórica frente al Qorikancha • Mate de coca de bienvenida y oxigenación	CATALOG	f	null	{"address": "Plazoleta Santo Domingo 259, Centro Histórico, Cusco 08002", "latitude": -13.5204, "roomType": "Colonial Deluxe Suite", "hotelName": "Palacio del Inka, a Luxury Collection Hotel", "longitude": -71.9754, "checkInTime": "14:00", "checkOutTime": "12:00"}	2026-11-17 23:58:51.361531+00	2026-11-17 23:58:51.361531+00
res-53a2e3217526aaa3	trip-714723e7ff2894ef	TRAIN	CONFIRMED	PeruRail Vistadome	PRU-93103	Ollantaytambo (Valle Sagrado)	Machu Picchu Pueblo (Aguas Calientes)	2026-11-27 06:00:00+00	2026-11-27 08:00:00+00	378.4	USD	Vagón panorámico con ventanas trapezoidales en el techo y snack andino a bordo	CATALOG	f	null	{"seat": "15A", "wagon": "Coche B", "trainLine": "PeruRail Vistadome Observatory", "trainNumber": "VSD 303", "arrivalStation": "Estación Machu Picchu Pueblo", "departureStation": "Estación Ollantaytambo"}	2026-11-17 23:58:51.361531+00	2026-11-17 23:58:51.361531+00
res-906f34bd0675d5d6	trip-db99b55be3a82fe3	FLIGHT	CONFIRMED	Aerolíneas Argentinas	AR-1140	Buenos Aires Aeroparque (AEP)	San Carlos de Bariloche (BRC)	2026-11-23 07:00:00+00	2026-11-23 09:25:00+00	265	USD	Importada directamente por código PNR (Aerolíneas Argentinas)	PROVIDER_API	f	null	{"gate": "07", "seat": "05F", "terminal": "Nacional", "cabinClass": "Club Economy", "flightNumber": "AR 1680", "arrivalAirport": "BRC", "departureAirport": "AEP"}	2026-11-21 18:39:31.508512+00	2026-11-21 18:39:43.939081+00
res-fc4f5915f45d1635	trip-9aeda99c6c53875b	FLIGHT	CONFIRMED	Aerolíneas Argentinas	AR-1140	Buenos Aires Aeroparque (AEP)	San Carlos de Bariloche (BRC)	2026-11-23 07:00:00+00	2026-11-23 09:25:00+00	265	USD	Sincronizada desde tu bandeja (@aerolineas.com.ar)	EMAIL_SCAN	f	null	{"gate": "07", "seat": "05F", "terminal": "Nacional", "cabinClass": "Club Economy", "flightNumber": "AR 1680", "arrivalAirport": "BRC", "departureAirport": "AEP"}	2026-11-21 19:25:35.865825+00	2026-11-21 19:25:35.865825+00
res-91370323c300f9c0	trip-16c90494f7187db4	HOTEL	CONFIRMED	Booking.com	BKG-5508	Bustillo Km 25, Bariloche	Llao Llao Resort, Golf & Spa	2026-11-23 14:00:00+00	2026-11-27 10:00:00+00	1180	USD	Llao Llao Resort, Golf & Spa Bariloche (4 Noches)	CATALOG	f	null	{"address": "Av. Ezequiel Bustillo Km 25, R8401 San Carlos de Bariloche", "latitude": -41.0583, "roomType": "Suite Lago Moreno", "hotelName": "Llao Llao Resort, Golf & Spa", "longitude": -71.5332, "checkInTime": "15:00", "checkOutTime": "11:00"}	2026-11-21 19:10:16.606956+00	2026-11-21 19:11:03.655215+00
res-fadf30eed0c99183	trip-16c90494f7187db4	HOTEL	CONFIRMED	Booking.com	BKG-6315	Distrito 15, París	Pullman Paris Tour Eiffel	2026-11-26 14:00:00+00	2026-11-30 11:00:00+00	1040	EUR	Pullman Paris Tour Eiffel (4 Noches)	CATALOG	f	null	{"address": "18 Avenue de Suffren, 75015 París", "latitude": 48.8556, "roomType": "Deluxe Eiffel View Balcony", "hotelName": "Pullman Paris Tour Eiffel", "longitude": 2.2928, "checkInTime": "15:00", "checkOutTime": "12:00"}	2026-11-21 19:09:40.558874+00	2026-11-21 19:09:40.558874+00
res-a2b95876625730d9	trip-16c90494f7187db4	HOTEL	CONFIRMED	Booking.com	BKG-9012	Gran Vía, Madrid	Hotel Riu Plaza España	2026-11-27 14:00:00+00	2026-12-01 11:00:00+00	760	EUR	Hotel Riu Plaza España Madrid (4 Noches)	CATALOG	f	null	{"address": "Calle Gran Vía 84, Centro, 28013 Madrid", "latitude": 40.4239, "roomType": "Deluxe King Gran Vía", "hotelName": "Hotel Riu Plaza España", "longitude": -3.7112, "checkInTime": "15:00", "checkOutTime": "12:00"}	2026-11-21 19:10:53.407871+00	2026-11-21 19:10:53.407871+00
res-33e13a6fae9272bf	trip-857a1e7f07e1b2ae	FLIGHT	CONFIRMED	All Nippon Airways	ANA-62901	Los Ángeles (LAX)	Tokio Haneda (HND)	2026-11-25 09:00:00+00	2026-11-25 21:00:00+00	1387.2	USD	Sincronizada desde tu bandeja (@ana.co.jp)	EMAIL_SCAN	f	null	{"gate": "114", "seat": "14A", "terminal": "Terminal 3", "cabinClass": "Premium Economy", "flightNumber": "NH 105", "arrivalAirport": "HND", "departureAirport": "LAX"}	2026-11-21 19:25:36.445826+00	2026-11-21 19:25:36.445826+00
res-1f535ab2b306331d	trip-857a1e7f07e1b2ae	HOTEL	CONFIRMED	Booking.com	BKG-62902	Kabukicho, Shinjuku, Tokio	Hotel Gracery Shinjuku	2026-11-25 14:00:00+00	2026-12-04 14:00:00+00	1098.2	USD	Sincronizada desde tu bandeja (@booking.com)	EMAIL_SCAN	f	null	{"address": "1-19-1 Kabukicho, Shinjuku-ku, Tokio 160-8466", "latitude": 35.6954, "roomType": "Superior Double Panorama", "hotelName": "Hotel Gracery Shinjuku", "longitude": 139.7021, "checkInTime": "15:00", "checkOutTime": "11:00"}	2026-11-21 19:25:36.450407+00	2026-11-21 19:25:36.450407+00
res-43b1ee69edaeffc8	trip-857a1e7f07e1b2ae	TRAIN	CONFIRMED	JR Central Shinkansen	JR-62903	Tokio Station	Kioto Station	2026-11-29 08:00:00+00	2026-11-29 10:00:00+00	404.6	USD	Sincronizada desde tu bandeja (@jrpass.com)	EMAIL_SCAN	f	null	{"seat": "12E", "wagon": "Coche 7", "trainLine": "Tokaido Shinkansen", "trainNumber": "Nozomi 21", "arrivalStation": "Kioto Central (Andén 12)", "departureStation": "Tokio Central (Andén 16)"}	2026-11-21 19:25:36.455548+00	2026-11-21 19:25:36.455548+00
res-c914c60f267ad9c4	trip-9aeda99c6c53875b	HOTEL	CONFIRMED	Booking.com	BKG-5508	Bustillo Km 25, Bariloche	Llao Llao Resort, Golf & Spa	2026-11-23 14:00:00+00	2026-11-27 10:00:00+00	1180	USD	Sincronizada desde tu bandeja (@booking.com)	EMAIL_SCAN	f	null	{"address": "Av. Ezequiel Bustillo Km 25, R8401 San Carlos de Bariloche", "latitude": -41.0583, "roomType": "Suite Lago Moreno", "hotelName": "Llao Llao Resort, Golf & Spa", "longitude": -71.5332, "checkInTime": "15:00", "checkOutTime": "11:00"}	2026-11-21 19:25:36.460398+00	2026-11-21 19:25:36.460398+00
res-98a1cf1839bf76e9	trip-247691b21ac4d5fa	HOTEL	CONFIRMED	Booking.com	BKG-5508	Bustillo Km 25, Bariloche	Llao Llao Resort, Golf & Spa	2026-11-23 14:00:00+00	2026-11-27 10:00:00+00	1180	USD	Sincronizada desde tu bandeja (@booking.com)	EMAIL_SCAN	f	null	{"address": "Av. Ezequiel Bustillo Km 25, R8401 San Carlos de Bariloche", "latitude": -41.0583, "roomType": "Suite Lago Moreno", "hotelName": "Llao Llao Resort, Golf & Spa", "longitude": -71.5332, "checkInTime": "15:00", "checkOutTime": "11:00"}	2026-11-21 19:28:22.778187+00	2026-11-21 19:28:22.778187+00
\.


--
-- Data for Name: sender_rules; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.sender_rules (id, user_id, pattern, provider_label, category, enabled, is_default, created_at) FROM stdin;
sr-0d250813ed008f8a	usr-sofia	@trotamundos.app	Catálogo Oficial Trotamundos	Plataforma	t	t	2026-11-17 23:52:32.286535+00
sr-9d06073903925298	usr-sofia	@booking.com	Booking.com	Alojamientos	t	t	2026-11-17 23:52:33.286535+00
sr-78eb070c22dfd480	usr-sofia	@ana.co.jp	All Nippon Airways	Aerolíneas	t	t	2026-11-17 23:52:34.286535+00
sr-a6754e7d41d96436	usr-sofia	@iberia.com	Iberia Líneas Aéreas	Aerolíneas	t	t	2026-11-17 23:52:35.286535+00
sr-c3fc4851084c34f7	usr-sofia	@aerolineas.com.ar	Aerolíneas Argentinas	Aerolíneas	t	t	2026-11-17 23:52:36.286535+00
sr-f3bc99450fbf8c0c	usr-sofia	@latam.com	LATAM Airlines	Aerolíneas	t	t	2026-11-17 23:52:37.286535+00
sr-4d39da5afe6c44d0	usr-sofia	@ita-airways.com	ITA Airways	Aerolíneas	t	t	2026-11-17 23:52:38.286535+00
sr-c27fbe65884204a5	usr-sofia	@airfrance.fr	Air France	Aerolíneas	t	t	2026-11-17 23:52:39.286535+00
sr-aa4a701fc08427b4	usr-sofia	@jrpass.com	JR Central Shinkansen	Trenes	t	t	2026-11-17 23:52:40.286535+00
sr-67e5cdafa6551965	usr-sofia	@renfe.com	Renfe AVE	Trenes	t	t	2026-11-17 23:52:41.286535+00
sr-e95f16ad44219539	usr-sofia	@eurostar.com	Eurostar International	Trenes	t	t	2026-11-17 23:52:42.286535+00
sr-ea5544274d644673	usr-sofia	@perurail.com	PeruRail Vistadome	Trenes	t	t	2026-11-17 23:52:43.286535+00
sr-64ca143adf5ee368	usr-sofia	@hertz.com	Hertz Rent a Car	Vehículos	t	t	2026-11-17 23:52:44.286535+00
sr-160cc241b250b0bc	usr-sofia	@avis.com	Avis Prestige	Vehículos	t	t	2026-11-17 23:52:45.286535+00
sr-4c4a17d829391758	usr-sofia	@alsa.es	ALSA Autobuses	Autobuses	t	t	2026-11-17 23:52:46.286535+00
sr-a1415ed521b8e408	usr-sofia	@viabariloche.com.ar	Vía Bariloche	Autobuses	t	t	2026-11-17 23:52:47.286535+00
sr-cfbba9c577123696	usr-lucas	@trotamundos.app	Catálogo Oficial Trotamundos	Plataforma	t	t	2026-11-17 23:52:32.302218+00
sr-ca839f417093c085	usr-lucas	@booking.com	Booking.com	Alojamientos	t	t	2026-11-17 23:52:33.302218+00
sr-ab04e1aa5de4624f	usr-lucas	@ana.co.jp	All Nippon Airways	Aerolíneas	t	t	2026-11-17 23:52:34.302218+00
sr-f90a9f2b925f935f	usr-lucas	@iberia.com	Iberia Líneas Aéreas	Aerolíneas	t	t	2026-11-17 23:52:35.302218+00
sr-a12fa8564280245f	usr-lucas	@aerolineas.com.ar	Aerolíneas Argentinas	Aerolíneas	t	t	2026-11-17 23:52:36.302218+00
sr-438e47939feec343	usr-lucas	@latam.com	LATAM Airlines	Aerolíneas	t	t	2026-11-17 23:52:37.302218+00
sr-fea907d1c900aed0	usr-lucas	@ita-airways.com	ITA Airways	Aerolíneas	t	t	2026-11-17 23:52:38.302218+00
sr-2d67313f22b4c284	usr-lucas	@airfrance.fr	Air France	Aerolíneas	t	t	2026-11-17 23:52:39.302218+00
sr-10abc2d29099e734	usr-lucas	@jrpass.com	JR Central Shinkansen	Trenes	t	t	2026-11-17 23:52:40.302218+00
sr-27a29596893943cb	usr-lucas	@renfe.com	Renfe AVE	Trenes	t	t	2026-11-17 23:52:41.302218+00
sr-4296c570da145745	usr-lucas	@eurostar.com	Eurostar International	Trenes	t	t	2026-11-17 23:52:42.302218+00
sr-9e124ccfda4590fb	usr-lucas	@perurail.com	PeruRail Vistadome	Trenes	t	t	2026-11-17 23:52:43.302218+00
sr-9392750a4428f58e	usr-lucas	@hertz.com	Hertz Rent a Car	Vehículos	t	t	2026-11-17 23:52:44.302218+00
sr-968ed9e0f401ae40	usr-lucas	@avis.com	Avis Prestige	Vehículos	t	t	2026-11-17 23:52:45.302218+00
sr-53b424f86b841b95	usr-lucas	@alsa.es	ALSA Autobuses	Autobuses	t	t	2026-11-17 23:52:46.302218+00
sr-e6f01ede5591e59c	usr-lucas	@viabariloche.com.ar	Vía Bariloche	Autobuses	t	t	2026-11-17 23:52:47.302218+00
sr-8a8e7c1a629b09bc	usr-28802ee52e6b868c	@trotamundos.app	Catálogo Oficial Trotamundos	Plataforma	t	t	2026-11-17 23:54:38.971046+00
sr-698789d36863564f	usr-28802ee52e6b868c	@booking.com	Booking.com	Alojamientos	t	t	2026-11-17 23:54:39.971046+00
sr-4c662c07aa8a998c	usr-28802ee52e6b868c	@ana.co.jp	All Nippon Airways	Aerolíneas	t	t	2026-11-17 23:54:40.971046+00
sr-1e072684e61dde0c	usr-28802ee52e6b868c	@iberia.com	Iberia Líneas Aéreas	Aerolíneas	t	t	2026-11-17 23:54:41.971046+00
sr-d537dad94d25e0da	usr-28802ee52e6b868c	@aerolineas.com.ar	Aerolíneas Argentinas	Aerolíneas	t	t	2026-11-17 23:54:42.971046+00
sr-624400e5b46f0c39	usr-28802ee52e6b868c	@latam.com	LATAM Airlines	Aerolíneas	t	t	2026-11-17 23:54:43.971046+00
sr-21bab29e4c7fe40f	usr-28802ee52e6b868c	@ita-airways.com	ITA Airways	Aerolíneas	t	t	2026-11-17 23:54:44.971046+00
sr-5e301aed420d9324	usr-28802ee52e6b868c	@airfrance.fr	Air France	Aerolíneas	t	t	2026-11-17 23:54:45.971046+00
sr-bfa5b5d4ae732cad	usr-28802ee52e6b868c	@jrpass.com	JR Central Shinkansen	Trenes	t	t	2026-11-17 23:54:46.971046+00
sr-1ed10a51af2c8d72	usr-28802ee52e6b868c	@renfe.com	Renfe AVE	Trenes	t	t	2026-11-17 23:54:47.971046+00
sr-408c4653188769cf	usr-28802ee52e6b868c	@eurostar.com	Eurostar International	Trenes	t	t	2026-11-17 23:54:48.971046+00
sr-17e574de3c8938ae	usr-28802ee52e6b868c	@perurail.com	PeruRail Vistadome	Trenes	t	t	2026-11-17 23:54:49.971046+00
sr-09a543466c95a971	usr-28802ee52e6b868c	@hertz.com	Hertz Rent a Car	Vehículos	t	t	2026-11-17 23:54:50.971046+00
sr-b2f8d9e9674a9c04	usr-28802ee52e6b868c	@avis.com	Avis Prestige	Vehículos	t	t	2026-11-17 23:54:51.971046+00
sr-fc888f6659277fcd	usr-28802ee52e6b868c	@alsa.es	ALSA Autobuses	Autobuses	t	t	2026-11-17 23:54:52.971046+00
sr-9472ef335e36d17a	usr-28802ee52e6b868c	@viabariloche.com.ar	Vía Bariloche	Autobuses	t	t	2026-11-17 23:54:53.971046+00
sr-176a30687e1c9ffb	usr-ed97b9ed908373ef	@trotamundos.app	Catálogo Oficial Trotamundos	Plataforma	t	t	2026-11-17 23:58:13.309275+00
sr-28d4f0cf93a17758	usr-ed97b9ed908373ef	@booking.com	Booking.com	Alojamientos	t	t	2026-11-17 23:58:14.309275+00
sr-b1a69165cfdb1eb2	usr-ed97b9ed908373ef	@ana.co.jp	All Nippon Airways	Aerolíneas	t	t	2026-11-17 23:58:15.309275+00
sr-ec16e1f256eca765	usr-ed97b9ed908373ef	@iberia.com	Iberia Líneas Aéreas	Aerolíneas	t	t	2026-11-17 23:58:16.309275+00
sr-76ef58179ffc0d19	usr-ed97b9ed908373ef	@aerolineas.com.ar	Aerolíneas Argentinas	Aerolíneas	t	t	2026-11-17 23:58:17.309275+00
sr-df0959a8df96cb9b	usr-ed97b9ed908373ef	@latam.com	LATAM Airlines	Aerolíneas	t	t	2026-11-17 23:58:18.309275+00
sr-a471335111c6340c	usr-ed97b9ed908373ef	@ita-airways.com	ITA Airways	Aerolíneas	t	t	2026-11-17 23:58:19.309275+00
sr-9318de700aa30221	usr-ed97b9ed908373ef	@airfrance.fr	Air France	Aerolíneas	t	t	2026-11-17 23:58:20.309275+00
sr-fd89fdabcd594771	usr-ed97b9ed908373ef	@jrpass.com	JR Central Shinkansen	Trenes	t	t	2026-11-17 23:58:21.309275+00
sr-920a5e33355a1753	usr-ed97b9ed908373ef	@renfe.com	Renfe AVE	Trenes	t	t	2026-11-17 23:58:22.309275+00
sr-017db0cf5790c63e	usr-ed97b9ed908373ef	@eurostar.com	Eurostar International	Trenes	t	t	2026-11-17 23:58:23.309275+00
sr-127855b454ca9b63	usr-ed97b9ed908373ef	@perurail.com	PeruRail Vistadome	Trenes	t	t	2026-11-17 23:58:24.309275+00
sr-22725103abee3821	usr-ed97b9ed908373ef	@hertz.com	Hertz Rent a Car	Vehículos	t	t	2026-11-17 23:58:25.309275+00
sr-f9c75b6ebf7e1048	usr-ed97b9ed908373ef	@avis.com	Avis Prestige	Vehículos	t	t	2026-11-17 23:58:26.309275+00
sr-c3a8c0d1fd3a7602	usr-ed97b9ed908373ef	@alsa.es	ALSA Autobuses	Autobuses	t	t	2026-11-17 23:58:27.309275+00
sr-2b15c5245601ccd8	usr-ed97b9ed908373ef	@viabariloche.com.ar	Vía Bariloche	Autobuses	t	t	2026-11-17 23:58:28.309275+00
sr-703831b426d77c47	usr-96b7e95209c554b6	@trotamundos.app	Catálogo Oficial Trotamundos	Plataforma	t	t	2026-11-21 18:23:59.848644+00
sr-ec7b4090cb0ed903	usr-96b7e95209c554b6	@booking.com	Booking.com	Alojamientos	t	t	2026-11-21 18:24:00.848644+00
sr-3af8a58f98e5df41	usr-96b7e95209c554b6	@ana.co.jp	All Nippon Airways	Aerolíneas	t	t	2026-11-21 18:24:01.848644+00
sr-2d3bcbea4f8ebd7d	usr-96b7e95209c554b6	@iberia.com	Iberia Líneas Aéreas	Aerolíneas	t	t	2026-11-21 18:24:02.848644+00
sr-c40cfbf4c37918d8	usr-96b7e95209c554b6	@aerolineas.com.ar	Aerolíneas Argentinas	Aerolíneas	t	t	2026-11-21 18:24:03.848644+00
sr-818b05bab194222d	usr-96b7e95209c554b6	@latam.com	LATAM Airlines	Aerolíneas	t	t	2026-11-21 18:24:04.848644+00
sr-ff1cc0c0d62b1241	usr-96b7e95209c554b6	@ita-airways.com	ITA Airways	Aerolíneas	t	t	2026-11-21 18:24:05.848644+00
sr-af27e6e2acba5675	usr-96b7e95209c554b6	@airfrance.fr	Air France	Aerolíneas	t	t	2026-11-21 18:24:06.848644+00
sr-fe56635742932281	usr-96b7e95209c554b6	@jrpass.com	JR Central Shinkansen	Trenes	t	t	2026-11-21 18:24:07.848644+00
sr-03c40be853389197	usr-96b7e95209c554b6	@renfe.com	Renfe AVE	Trenes	t	t	2026-11-21 18:24:08.848644+00
sr-88d337e1391165a6	usr-96b7e95209c554b6	@eurostar.com	Eurostar International	Trenes	t	t	2026-11-21 18:24:09.848644+00
sr-abd98518a53482bf	usr-96b7e95209c554b6	@perurail.com	PeruRail Vistadome	Trenes	t	t	2026-11-21 18:24:10.848644+00
sr-a2bd782b6c2cb58e	usr-96b7e95209c554b6	@hertz.com	Hertz Rent a Car	Vehículos	t	t	2026-11-21 18:24:11.848644+00
sr-bce07161dade043b	usr-96b7e95209c554b6	@avis.com	Avis Prestige	Vehículos	t	t	2026-11-21 18:24:12.848644+00
sr-0eff839a6406e175	usr-96b7e95209c554b6	@alsa.es	ALSA Autobuses	Autobuses	t	t	2026-11-21 18:24:13.848644+00
sr-61c24a877f1626c6	usr-96b7e95209c554b6	@viabariloche.com.ar	Vía Bariloche	Autobuses	t	t	2026-11-21 18:24:14.848644+00
sr-6bf91159cb9f1f84	usr-c20e381f93bcba4e	@trotamundos.app	Catálogo Oficial Trotamundos	Plataforma	t	t	2026-11-21 18:50:10.050214+00
sr-0bf77bd3c4ab012f	usr-c20e381f93bcba4e	@booking.com	Booking.com	Alojamientos	t	t	2026-11-21 18:50:11.050214+00
sr-a7ed4ac1abeeddc9	usr-c20e381f93bcba4e	@ana.co.jp	All Nippon Airways	Aerolíneas	t	t	2026-11-21 18:50:12.050214+00
sr-8956eaa3aa69f138	usr-c20e381f93bcba4e	@iberia.com	Iberia Líneas Aéreas	Aerolíneas	t	t	2026-11-21 18:50:13.050214+00
sr-a4e2f4e1d24ab88d	usr-c20e381f93bcba4e	@aerolineas.com.ar	Aerolíneas Argentinas	Aerolíneas	t	t	2026-11-21 18:50:14.050214+00
sr-39585ab0c8e4ccb3	usr-c20e381f93bcba4e	@latam.com	LATAM Airlines	Aerolíneas	t	t	2026-11-21 18:50:15.050214+00
sr-7f1a4d667030c880	usr-c20e381f93bcba4e	@ita-airways.com	ITA Airways	Aerolíneas	t	t	2026-11-21 18:50:16.050214+00
sr-c393a17484395619	usr-c20e381f93bcba4e	@airfrance.fr	Air France	Aerolíneas	t	t	2026-11-21 18:50:17.050214+00
sr-da4c981038671825	usr-c20e381f93bcba4e	@jrpass.com	JR Central Shinkansen	Trenes	t	t	2026-11-21 18:50:18.050214+00
sr-02bdb0420fbc04ae	usr-c20e381f93bcba4e	@renfe.com	Renfe AVE	Trenes	t	t	2026-11-21 18:50:19.050214+00
sr-332939dfbea2b1ac	usr-c20e381f93bcba4e	@eurostar.com	Eurostar International	Trenes	t	t	2026-11-21 18:50:20.050214+00
sr-cfc36c7271ef3ec0	usr-c20e381f93bcba4e	@perurail.com	PeruRail Vistadome	Trenes	t	t	2026-11-21 18:50:21.050214+00
sr-1944a36537841ba3	usr-c20e381f93bcba4e	@hertz.com	Hertz Rent a Car	Vehículos	t	t	2026-11-21 18:50:22.050214+00
sr-8b2bbfa997a7714e	usr-c20e381f93bcba4e	@avis.com	Avis Prestige	Vehículos	t	t	2026-11-21 18:50:23.050214+00
sr-d2f6e8a3677188aa	usr-c20e381f93bcba4e	@alsa.es	ALSA Autobuses	Autobuses	t	t	2026-11-21 18:50:24.050214+00
sr-4b7de65c3ef02323	usr-c20e381f93bcba4e	@viabariloche.com.ar	Vía Bariloche	Autobuses	t	t	2026-11-21 18:50:25.050214+00
sr-c7dcd2137a635e2a	usr-c7ae6c1824782002	@trotamundos.app	Catálogo Oficial Trotamundos	Plataforma	t	t	2026-11-21 19:22:23.709904+00
sr-dd9c2b0fa8e41751	usr-c7ae6c1824782002	@ana.co.jp	All Nippon Airways	Aerolíneas	t	t	2026-11-21 19:22:25.709904+00
sr-5f770409b9f0b92f	usr-c7ae6c1824782002	@iberia.com	Iberia Líneas Aéreas	Aerolíneas	t	t	2026-11-21 19:22:26.709904+00
sr-65dded7ebdb96bde	usr-c7ae6c1824782002	@aerolineas.com.ar	Aerolíneas Argentinas	Aerolíneas	t	t	2026-11-21 19:22:27.709904+00
sr-6945aa872b0ff3c0	usr-c7ae6c1824782002	@latam.com	LATAM Airlines	Aerolíneas	t	t	2026-11-21 19:22:28.709904+00
sr-ccea87f6550b7a2f	usr-c7ae6c1824782002	@ita-airways.com	ITA Airways	Aerolíneas	t	t	2026-11-21 19:22:29.709904+00
sr-69983b59c2ed14ee	usr-c7ae6c1824782002	@airfrance.fr	Air France	Aerolíneas	t	t	2026-11-21 19:22:30.709904+00
sr-36b1b7628560dd16	usr-c7ae6c1824782002	@jrpass.com	JR Central Shinkansen	Trenes	t	t	2026-11-21 19:22:31.709904+00
sr-389b6b768a3008f9	usr-c7ae6c1824782002	@renfe.com	Renfe AVE	Trenes	t	t	2026-11-21 19:22:32.709904+00
sr-3b0bf5551a6d2702	usr-c7ae6c1824782002	@eurostar.com	Eurostar International	Trenes	t	t	2026-11-21 19:22:33.709904+00
sr-557d4effa45c8271	usr-c7ae6c1824782002	@perurail.com	PeruRail Vistadome	Trenes	t	t	2026-11-21 19:22:34.709904+00
sr-0fce51b0b797b6f9	usr-c7ae6c1824782002	@hertz.com	Hertz Rent a Car	Vehículos	t	t	2026-11-21 19:22:35.709904+00
sr-d72b1573baea018c	usr-c7ae6c1824782002	@avis.com	Avis Prestige	Vehículos	t	t	2026-11-21 19:22:36.709904+00
sr-17d1a358bf00805e	usr-c7ae6c1824782002	@alsa.es	ALSA Autobuses	Autobuses	t	t	2026-11-21 19:22:37.709904+00
sr-8eac360af250a90e	usr-c7ae6c1824782002	@viabariloche.com.ar	Vía Bariloche	Autobuses	t	t	2026-11-21 19:22:38.709904+00
sr-477cddd3d43c75b3	usr-c7ae6c1824782002	@booking.com	Booking.com	Alojamientos	t	t	2026-11-21 19:22:24.709904+00
\.


--
-- Data for Name: support_tickets; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.support_tickets (id, reference_code, reservation_id, trip_id, user_id, agency_id, agency_name, issue_summary, status, created_at) FROM stdin;
tkt-7bb961af3b59b1eb	TRM-6C968F	res-5eec4f7372438873	trip-aeaf1492eceae331	usr-28802ee52e6b868c	ag-iberia	Iberia & Europa Travel Desk	Asistencia solicitada para Vía Bariloche (VBR-1920) — Confirmada	OPEN	2026-11-17 23:57:09.577638+00
\.


--
-- Data for Name: travel_agencies; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.travel_agencies (id, name, support_email, phone, sla_hours, specialty) FROM stdin;
ag-meridian	Meridian Concierge Global	concierge@meridiantravel.app	+1 (305) 555-0194	2	Asistencia prioritaria 24/7 en vuelos transcontinentales y hoteles de lujo
ag-andes	Andes & Pacífico Asistencia	soporte@andespacifico.app	+54 11 5550-4820	3	Rutas de Sudamérica, Patagonia, Cusco y cobertura regional multi-tramo
ag-iberia	Iberia & Europa Travel Desk	desk@europatraveldesk.app	+34 91 555 8310	2	Conexiones ferroviarias de alta velocidad y circuitos europeos
ag-nippon	Nippon Express Concierge	assist@nipponconcierge.app	+81 3 5550 9210	2	Especialistas en Japón, pases Shinkansen y hospitalidad en Asia-Pacífico
\.


--
-- Data for Name: trips; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.trips (id, owner_id, shared_with, shared_with_emails, title, destination_city, country_code, cover_image_url, status, start_date, end_date, created_at, updated_at) FROM stdin;
trip-aeaf1492eceae331	usr-28802ee52e6b868c	null	["hermes@hotmail.com"]	Viaje a Terminal San Martín de los Andes	Terminal San Martín de los Andes	INT	https://images.unsplash.com/photo-1502602898657-3e91760cbb34?auto=format&fit=crop&w=1000&q=80	ACTIVE	2026-11-25 08:00:00+00	2026-11-25 12:15:00+00	2026-11-17 23:56:58.492336+00	2026-11-17 23:57:40.307645+00
trip-714723e7ff2894ef	usr-ed97b9ed908373ef	["usr-28802ee52e6b868c"]	["changalardicd@gmail.com"]	Cusco, Valle Sagrado & Machu Picchu	Cusco	PE	https://images.unsplash.com/photo-1526392060635-9d6019884377?auto=format&fit=crop&w=1200&q=85	ACTIVE	2026-11-24 07:00:00+00	2026-11-30 13:00:00+00	2026-11-17 23:58:51.361531+00	2026-11-17 23:59:09.391945+00
trip-20e3ddc7634b610e	usr-96b7e95209c554b6	null	[]	Conociendo Tanidl	Tandil	AR	https://images.unsplash.com/photo-1502602898657-3e91760cbb34?auto=format&fit=crop&w=1000&q=80	ARCHIVED	2026-11-28 18:35:50.262769+00	2026-12-03 18:35:50.262769+00	2026-11-21 18:35:50.262769+00	2026-11-21 19:01:13.213607+00
trip-a082ccac71c2fee2	usr-96b7e95209c554b6	null	[]	Conociendo Italia	Roma	IT	https://images.unsplash.com/photo-1552832230-c0197dd311b5?auto=format&fit=crop&w=1000&q=80	ARCHIVED	2026-11-28 18:35:20.975087+00	2026-12-03 18:35:20.975087+00	2026-11-21 18:35:20.975087+00	2026-11-21 19:01:14.075983+00
trip-db99b55be3a82fe3	usr-96b7e95209c554b6	null	[]	Viaje a Bariloche	Bariloche	AR	https://images.unsplash.com/photo-1518709268805-4e9042af9f23?auto=format&fit=crop&w=1000&q=80	ARCHIVED	2026-11-23 07:00:00+00	2026-11-23 09:25:00+00	2026-11-21 18:39:31.443295+00	2026-11-21 19:01:15.905804+00
trip-16c90494f7187db4	usr-96b7e95209c554b6	null	[]	Viaje a París	París	FR	https://images.unsplash.com/photo-1502602898657-3e91760cbb34?auto=format&fit=crop&w=1000&q=80	ACTIVE	2026-11-23 14:00:00+00	2026-12-01 11:00:00+00	2026-11-21 19:09:40.486523+00	2026-11-21 19:11:03.65522+00
trip-857a1e7f07e1b2ae	usr-c7ae6c1824782002	null	[]	Viaje a Tokio	Tokio	JP	https://images.unsplash.com/photo-1540959733332-eab4deabeeaf?auto=format&fit=crop&w=1000&q=80	ACTIVE	2026-11-25 09:00:00+00	2026-12-04 14:00:00+00	2026-11-21 19:25:36.443065+00	2026-11-21 19:25:36.45555+00
trip-9aeda99c6c53875b	usr-c7ae6c1824782002	null	[]	Viaje a Bariloche	Bariloche	AR	https://images.unsplash.com/photo-1518709268805-4e9042af9f23?auto=format&fit=crop&w=1000&q=80	ACTIVE	2026-11-23 07:00:00+00	2026-11-27 10:00:00+00	2026-11-21 19:25:35.86193+00	2026-11-21 19:25:36.460399+00
trip-247691b21ac4d5fa	usr-96b7e95209c554b6	null	[]	Viaje America	Bariloche	AR	https://images.unsplash.com/photo-1518709268805-4e9042af9f23?auto=format&fit=crop&w=1000&q=80	ACTIVE	2026-11-23 14:00:00+00	2026-11-27 10:00:00+00	2026-11-21 19:10:46.061199+00	2026-11-21 19:28:22.778188+00
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.users (id, name, email, password_hash, role, home_currency, preferred_agency_id, mailbox_authorized, connected_email, avatar_initials, created_at) FROM stdin;
usr-sofia	Sofía Martínez	sofia.martinez@trotamundos.app	pbkdf2_sha256$15000$QtE0WdlCSoO01RcjTjG1GQ$BcpcOrnKZZ99hsKq099msJQEGDAm+rm+kUzYalsytKA	traveler	USD	ag-meridian	t	sofia.martinez@trotamundos.app	SM	2026-10-28 23:00:00+00
usr-lucas	Lucas Ferreyra	lucas.ferreyra@trotamundos.app	pbkdf2_sha256$15000$QtE0WdlCSoO01RcjTjG1GQ$BcpcOrnKZZ99hsKq099msJQEGDAm+rm+kUzYalsytKA	traveler	EUR	ag-iberia	t	lucas.ferreyra@trotamundos.app	LF	2026-11-02 23:00:00+00
usr-28802ee52e6b868c	Carlos Daniel Chang Galardi	changalardicd@gmail.com	pbkdf2_sha256$15000$TWp9mIdWaKntJlXvpPwvyw$LetN3v/Rlh1K+06t5XrgtYhRJtcBSF2WWVVfHe9fp5k	traveler	ARS	ag-iberia	t	changalardicd@gmail.com	CG	2026-11-17 23:54:38.941808+00
usr-ed97b9ed908373ef	Hermes	hermes@hotmail.com	pbkdf2_sha256$15000$JnTbWplPuQ68q65T6t+Cbw$ZPy23DUTwpswXgCXX4wVZhuvZfHOKpkm1Xl0aw3vqSo	traveler	ARS	ag-iberia	t	hermes@hotmail.com	HE	2026-11-17 23:58:13.306734+00
usr-c20e381f93bcba4e	Camila Verificadora	camila.verify.1795287009140@trotamundos.app	pbkdf2_sha256$15000$ckXqLANseGrrC7n2WTWsQg$o1R1IO3FnJ+pgdnoxXXRba1qq/3RzBOryj/h/z2ltuI	traveler	ARS	ag-iberia	t	camila.verify.1795287009140@trotamundos.app	CV	2026-11-21 18:50:10.04795+00
usr-96b7e95209c554b6	Daniel Galardi	chang27@gmail.com	pbkdf2_sha256$15000$z1btyf10IuWIjIAYx8kNkg$ggNysvdMCOsxeRxBvJ1AsvRYumEHp0899xPL1OGUpNg	traveler	ARS	ag-andes	t	chang27@gmail.com	DG	2026-11-21 18:23:59.830626+00
usr-c7ae6c1824782002	Tester Correo	tester.correo.1795288942@trotamundos.app	pbkdf2_sha256$15000$zgaMl52iwjwc2DsocLFhVQ$1FOM+P/nWMumrokEFiojPeZRgJDkpm+FC+A3RGkXUAc	traveler	USD	ag-andes	t	tester.correo.1795288942@trotamundos.app	TC	2026-11-21 19:22:23.667154+00
\.


--
-- Name: analytics_facts analytics_facts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.analytics_facts
    ADD CONSTRAINT analytics_facts_pkey PRIMARY KEY (id);


--
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


--
-- Name: catalog_offers catalog_offers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.catalog_offers
    ADD CONSTRAINT catalog_offers_pkey PRIMARY KEY (id);


--
-- Name: catalog_packages catalog_packages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.catalog_packages
    ADD CONSTRAINT catalog_packages_pkey PRIMARY KEY (id);


--
-- Name: email_confirmations email_confirmations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_confirmations
    ADD CONSTRAINT email_confirmations_pkey PRIMARY KEY (id);


--
-- Name: pending_syncs pending_syncs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pending_syncs
    ADD CONSTRAINT pending_syncs_pkey PRIMARY KEY (id);


--
-- Name: reservations reservations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reservations
    ADD CONSTRAINT reservations_pkey PRIMARY KEY (id);


--
-- Name: sender_rules sender_rules_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sender_rules
    ADD CONSTRAINT sender_rules_pkey PRIMARY KEY (id);


--
-- Name: support_tickets support_tickets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.support_tickets
    ADD CONSTRAINT support_tickets_pkey PRIMARY KEY (id);


--
-- Name: support_tickets support_tickets_reservation_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.support_tickets
    ADD CONSTRAINT support_tickets_reservation_id_key UNIQUE (reservation_id);


--
-- Name: travel_agencies travel_agencies_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.travel_agencies
    ADD CONSTRAINT travel_agencies_pkey PRIMARY KEY (id);


--
-- Name: trips trips_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trips
    ADD CONSTRAINT trips_pkey PRIMARY KEY (id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_catalog_offers_pnr; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_catalog_offers_pnr ON public.catalog_offers USING btree (sample_pnr);


--
-- Name: idx_catalog_offers_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_catalog_offers_type ON public.catalog_offers USING btree (type);


--
-- Name: idx_email_confirmations_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_email_confirmations_user ON public.email_confirmations USING btree (user_id);


--
-- Name: idx_reservations_trip; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_reservations_trip ON public.reservations USING btree (trip_id);


--
-- Name: idx_sender_rules_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sender_rules_user ON public.sender_rules USING btree (user_id);


--
-- Name: idx_trips_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_trips_owner ON public.trips USING btree (owner_id);


--
-- Name: email_confirmations email_confirmations_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_confirmations
    ADD CONSTRAINT email_confirmations_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: reservations reservations_trip_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reservations
    ADD CONSTRAINT reservations_trip_id_fkey FOREIGN KEY (trip_id) REFERENCES public.trips(id) ON DELETE CASCADE;


--
-- Name: sender_rules sender_rules_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sender_rules
    ADD CONSTRAINT sender_rules_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: support_tickets support_tickets_reservation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.support_tickets
    ADD CONSTRAINT support_tickets_reservation_id_fkey FOREIGN KEY (reservation_id) REFERENCES public.reservations(id) ON DELETE CASCADE;


--
-- Name: trips trips_owner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trips
    ADD CONSTRAINT trips_owner_id_fkey FOREIGN KEY (owner_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict ugLdUeSmgfXwZbCGEdl2i6wyo8qmgo7Ya1muWzhkraU6etAktc4bPIaxkZbPWyO

