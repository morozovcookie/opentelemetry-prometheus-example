# migrate/migrate:v4.15.1
FROM migrate/migrate@sha256:254d67343a2389383b530d9049610f21f4d302557288c649ebbb761ea96b10af

COPY ./scripts/docker/schema/docker-entrypoint.sh /docker-entrypoint.sh

COPY ./migrations /migrations

ENTRYPOINT ["/bin/sh", "/docker-entrypoint.sh"]
