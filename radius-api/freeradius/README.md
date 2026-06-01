# FreeRADIUS configuration

This directory contains the FreeRADIUS config used by the `freeradius` Docker
service. FreeRADIUS authenticates hotspot users against the MariaDB tables that
the **radius-api** service owns and migrates.

## How it fits together

```
Mikrotik hAP ac2  ──RADIUS(1812/1813)──►  FreeRADIUS  ──SQL──►  MariaDB
       ▲                                                            ▲
       │                                                            │
       └──────── CoA / Disconnect (3799) ◄── radius-api ───────────┘
                                                  ▲
                                                  │ HTTP (X-API-Key)
                                            billing backend
```

- The **billing backend** mints vouchers and calls **radius-api** over HTTP.
- **radius-api** writes `radcheck` / `radusergroup` / `radgroupreply` rows and
  registers the Mikrotik in the `nas` table.
- **FreeRADIUS** reads those tables (`read_clients = yes`, `read_groups = yes`)
  to authenticate and authorize hotspot logins.
- To kick a user offline, radius-api sends a RADIUS **Disconnect-Request** (CoA)
  to the Mikrotik on port 3799.

## Files

| File | Purpose |
|------|---------|
| `Dockerfile` | FreeRADIUS 3.2 + MySQL driver + templating entrypoint. |
| `docker-entrypoint.sh` | Waits for the DB, templates config from env vars, enables the SQL module, starts `radiusd`. |
| `raddb/clients.conf` | Static NAS definition (your Mikrotik) + localhost. |
| `raddb/mods-available/sql` | SQL module pointed at MariaDB. |

## Mikrotik dictionary

The `Mikrotik-Rate-Limit`, `Mikrotik-Total-Limit`, etc. attributes that the
radius-api writes into `radgroupreply` are part of the **Mikrotik dictionary**,
which ships with FreeRADIUS by default (`/usr/share/freeradius/dictionary.mikrotik`).
No extra steps are required.

## Debugging

Set `RADIUS_DEBUG=yes` in the environment to run `radiusd -X` (full debug). Then
test a voucher named `ABC123` from inside the container:

```sh
radtest ABC123 ABC123 127.0.0.1 0 <NAS_SHARED_SECRET>
```

A successful login returns `Access-Accept` and shows the `Mikrotik-Rate-Limit`
reply attribute from the package's profile.
