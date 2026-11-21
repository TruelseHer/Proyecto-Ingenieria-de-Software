#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

BACKUP_FILE="$SCRIPT_DIR/backups/trotamundos_backup.sql"
POSTGRES_USER="${POSTGRES_USER:-trotamundos}"
POSTGRES_DB="${POSTGRES_DB:-trotamundos_v2}"

FORCE_BUILD="false"
RESTAURAR="false"

for arg in "$@"; do
  case "$arg" in
    --build)
      FORCE_BUILD="true"
      ;;
    --restaurar)
      RESTAURAR="true"
      ;;
    -h|--help)
      echo "Uso: ./iniciar.sh [--build] [--restaurar]"
      echo "  --build      Reconstruye la imagen Docker de la aplicación antes de iniciar."
      echo "  --restaurar  Restaura obligatoriamente la base de datos desde backups/trotamundos_backup.sql."
      exit 0
      ;;
    *)
      echo "Opción no reconocida: $arg" >&2
      echo "Uso: ./iniciar.sh [--build] [--restaurar]" >&2
      exit 1
      ;;
  esac
done

if [ "$RESTAURAR" = "true" ] && [ ! -s "$BACKUP_FILE" ]; then
  echo "Error: se indicó --restaurar pero no existe un respaldo válido en $BACKUP_FILE." >&2
  exit 1
fi

echo "Iniciando base de datos y caché..."
docker compose up -d db cache

echo "Esperando a que PostgreSQL esté listo..."
DB_READY="false"
for _ in $(seq 1 30); do
  if docker compose exec -T db pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB" >/dev/null 2>&1; then
    DB_READY="true"
    break
  fi
  sleep 1
done

if [ "$DB_READY" != "true" ]; then
  echo "Error: PostgreSQL no respondió a tiempo después de 30 segundos." >&2
  exit 1
fi

if [ -s "$BACKUP_FILE" ]; then
  USERS_TABLE="$(docker compose exec -T db psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -tAc "SELECT to_regclass('public.users');" | tr -d '[:space:]')"
  if [ -z "$USERS_TABLE" ] || [ "$USERS_TABLE" = "NULL" ] || [ "$USERS_TABLE" = "null" ] || [ "$RESTAURAR" = "true" ]; then
    echo "Restaurando datos de PostgreSQL desde $BACKUP_FILE..."
    docker compose exec -T db psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1 < "$BACKUP_FILE" >/dev/null
    echo "Base de datos restaurada correctamente desde backups/trotamundos_backup.sql."
  else
    echo "Datos existentes en el volumen de PostgreSQL intactos (usa './iniciar.sh --restaurar' para recargar desde $BACKUP_FILE)."
  fi
else
  echo "No se encontró respaldo previo en $BACKUP_FILE; la app inicializará las tablas y el catálogo si es necesario."
fi

NEED_BUILD="$FORCE_BUILD"
if [ "$NEED_BUILD" != "true" ]; then
  APP_IMAGE="$(docker compose config --images 2>/dev/null | grep -E '(^|[-_/])app(:|$)' | head -n1 || true)"
  if [ -z "$APP_IMAGE" ] || ! docker image inspect "$APP_IMAGE" >/dev/null 2>&1; then
    NEED_BUILD="true"
  fi
fi

if [ "$NEED_BUILD" = "true" ]; then
  echo "Construyendo e iniciando la aplicación..."
  docker compose up -d --build app
else
  echo "Iniciando la aplicación..."
  docker compose up -d app
fi

echo "Verificando disponibilidad de http://localhost:8080/health..."
APP_READY="false"
for _ in $(seq 1 30); do
  if command -v curl >/dev/null 2>&1; then
    if curl -fsS http://localhost:8080/health >/dev/null 2>&1; then
      APP_READY="true"
      break
    fi
  elif command -v wget >/dev/null 2>&1; then
    if wget -q -O /dev/null http://localhost:8080/health >/dev/null 2>&1; then
      APP_READY="true"
      break
    fi
  elif docker compose ps --status running --services 2>/dev/null | grep -qx "app"; then
    APP_READY="true"
    break
  fi
  sleep 1
done

docker compose ps

if [ "$APP_READY" = "true" ]; then
  echo "Trotamundos V2 iniciado y operativo en: http://localhost:8080"
else
  echo "Los contenedores están iniciados; verifica el estado en unos segundos en: http://localhost:8080"
fi
