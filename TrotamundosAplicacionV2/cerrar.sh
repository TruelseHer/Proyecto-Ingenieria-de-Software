#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

BACKUP_DIR="$SCRIPT_DIR/backups"
BACKUP_FILE="$BACKUP_DIR/trotamundos_backup.sql"
POSTGRES_USER="${POSTGRES_USER:-trotamundos}"
POSTGRES_DB="${POSTGRES_DB:-trotamundos_v2}"

echo "Deteniendo Trotamundos V2 de manera segura..."

if docker compose ps --status running --services 2>/dev/null | grep -qx "db"; then
  echo "Respaldando base de datos PostgreSQL ('$POSTGRES_DB')..."
  mkdir -p "$BACKUP_DIR"
  TMP_BACKUP="$BACKUP_FILE.tmp"

  if docker compose exec -T db pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" --clean --if-exists --no-owner --no-privileges > "$TMP_BACKUP" && [ -s "$TMP_BACKUP" ]; then
    mv "$TMP_BACKUP" "$BACKUP_FILE"
    BACKUP_SIZE="$(du -h "$BACKUP_FILE" | cut -f1)"
    echo "Respaldo guardado en: $BACKUP_FILE ($BACKUP_SIZE)"
  else
    rm -f "$TMP_BACKUP"
    echo "Advertencia: no se pudo generar el respaldo SQL; se mantiene cualquier respaldo previo y el volumen de Docker."
  fi
else
  echo "El servicio 'db' no está en ejecución; se omitió el volcado SQL."
fi

echo "Cerrando contenedores (conservando volúmenes de datos)..."
docker compose down

echo "Trotamundos V2 cerrado correctamente. Todos los datos quedaron guardados."
