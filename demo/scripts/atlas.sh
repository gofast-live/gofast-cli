#!/bin/bash

cd "$(dirname "$0")"
cd ../app/service-core/storage

echo "Applying schema ..."
atlas schema apply -u "postgres://postgres:postgres@localhost:5432/db?sslmode=disable" -f ./schema.sql --dev-url "docker://postgres/15/dev"
