# FreeRADIUS 3.2 with the MySQL driver and a templating entrypoint.
FROM freeradius/freeradius-server:3.2.5

# mariadb-client → used by the entrypoint to wait for the DB / schema.
# gettext-base   → provides `envsubst` for config templating.
RUN apt-get update \
    && apt-get install -y --no-install-recommends mariadb-client gettext-base \
    && rm -rf /var/lib/apt/lists/*

COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

EXPOSE 1812/udp 1813/udp

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
