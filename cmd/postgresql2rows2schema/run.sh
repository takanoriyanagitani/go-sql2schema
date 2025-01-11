#!/bin/sh

export ENV_TABLE_NAME=table2025jan08test1544

export PGDATABASE=db2025jan08test1547
export PGHOST=127.0.0.1
export PGUSER=postgres

PGPORT=${PGPORT:-5432}
PGPASSWORD=${PGPASSWORD:-postgres}

container_name=go-sql2schema-test-db-2025jan11test0823
image_name=postgres:17.2-alpine3.21

_run_docker() {
	echo TODO: run using docker
}

_run_podman() {
	podpath=/etc/paths.d/podman-pkg
	test -f "${podpath}" || exec sh -c 'echo podman missing.; exit 1'
	podbin=$(cat "${podpath}")
	test -d "$podbin" || exec sh -c 'echo podman missing.; exit 1'
	export PATH="${podbin}:${PATH}"
	pd=podman
	$pd ps --filter=name="${container_name}" |
		fgrep -q "${container_name}" ||
		$pd \
			run \
			--detach \
			--name "${container_name}" \
			--env POSTGRES_PASSWORD=$PGPASSWORD \
			--env POSTGRES_HOST_AUTH_METHOD=trust \
			--publish $PGPORT:5432 \
			$image_name

	while sleep 0; do
		echo checking connection...
		$pd exec "${container_name}" env PGUSER=postgres pg_isready | fgrep accepting && break
		sleep 1
	done

	$pd exec "${container_name}" env PGUSER=postgres psql --list |
		fgrep -q "${PGDATABASE}" ||
		$pd \
			exec \
			--interactive \
			"${container_name}" \
			env PGUSER=postgres \
			PGDATABASE=postgres \
			psql \
			-c "CREATE DATABASE ${PGDATABASE}"

	echo "
		CREATE TABLE IF NOT EXISTS ${ENV_TABLE_NAME} (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			height FLOAT8 NOT NULL,
			amount INTEGER NOT NULL,
			active BOOL NOT NULL,
			data BYTEA NOT NULL,
			created BIGINT NOT NULL,
			rate FLOAT4 NOT NULL
		)
	" |
		$pd \
			exec \
			--interactive \
			"${container_name}" \
			env PGUSER=postgres \
			PGDATABASE=$PGDATABASE \
			psql
}

_run() {
	which docker | fgrep -q docker && _run_docker || _run_podman
}

_run

./postgresql2rows2schema > ./sample.d/sample.jsonl

cat ./sample.d/sample.jsonl |
  jaq -c
