#!/bin/sh
# FreeRADIUS entrypoint: templates clients.conf + the SQL module from env vars,
# waits for MariaDB, then runs the server. The MariaDB schema itself is created
# by the radius-api migrations (goose), so we only wait for connectivity here.
#
# The source templates are bind-mounted read-only under ${RADDB}/templates/ and
# rendered into the real raddb paths. We must NOT render in-place over a
# bind-mounted file: renaming/overwriting a mountpoint fails ("Device or
# resource busy"), and writing through it would pollute the repo template with
# resolved secrets. Rendering template → real path avoids both.
set -e

RADDB=/etc/raddb
TPL="${RADDB}/templates"

# 1) Wait for MariaDB to accept connections.
until mariadb -h"${SQL_HOST}" -P"${SQL_PORT:-3306}" -u"${SQL_USER}" -p"${SQL_PASSWORD}" -e "SELECT 1" >/dev/null 2>&1; do
  echo "[entrypoint] waiting for MariaDB at ${SQL_HOST}:${SQL_PORT:-3306}…"
  sleep 2
done
echo "[entrypoint] database is reachable."

# 2) Render the config. CRITICAL: limit envsubst to OUR variables only.
# Plain `envsubst` would also expand FreeRADIUS's own ${...} config variables
# (e.g. ${modconfdir}, ${.:name}, ${dialect}) and blank them out, breaking the
# parser. Passing an explicit SHELL-FORMAT list makes envsubst touch only these.
export NAS_SHARED_SECRET NAS_CLIENT_SUBNET SQL_HOST SQL_PORT SQL_USER SQL_PASSWORD SQL_DB
VARS='$NAS_SHARED_SECRET $NAS_CLIENT_SUBNET $SQL_HOST $SQL_PORT $SQL_USER $SQL_PASSWORD $SQL_DB'

if [ -f "${TPL}/clients.conf" ]; then
  envsubst "${VARS}" < "${TPL}/clients.conf" > "${RADDB}/clients.conf"
  echo "[entrypoint] rendered clients.conf"
fi

# 3) Render the SQL module config with DB credentials.
if [ -f "${TPL}/sql" ]; then
  envsubst "${VARS}" < "${TPL}/sql" > "${RADDB}/mods-available/sql"
  echo "[entrypoint] rendered mods-available/sql"
fi

# 4) Enable the SQL module (symlink into mods-enabled) — without this FreeRADIUS
#    logs `Ignoring "sql"` and cannot read vouchers/NAS from MariaDB.
if [ ! -e "${RADDB}/mods-enabled/sql" ]; then
  ln -sf ../mods-available/sql "${RADDB}/mods-enabled/sql"
  echo "[entrypoint] enabled sql module"
fi

# NOTE on Simultaneous-Use: the "1 voucher = 1 concurrent device" limit is
# enforced at the MIKROTIK layer (hotspot user profile shared-users=1, set by
# the generated .rsc), which is reliable and independent of accounting state.
# We deliberately do NOT enable FreeRADIUS's session{sql} simul counter here:
# without a working checkrad it can lock a customer out on a stale (un-stopped)
# radacct session — a real support/revenue risk — for no proven gain over the
# Mikrotik-side limit.

# 5) Start FreeRADIUS in the foreground (-f), logging to stdout. Enable debug
#    output (-x) only when RADIUS_DEBUG=yes.
DEBUG_FLAG=""
if [ "${RADIUS_DEBUG}" = "yes" ]; then
  DEBUG_FLAG="-x"
fi
set +e
exec freeradius -f -l stdout ${DEBUG_FLAG}
