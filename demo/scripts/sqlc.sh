#!/bin/bash

cd "$(dirname "$0")"
cd ../app/service-core/storage
rm -rf query


echo "Generating SQLC code ..."
sqlc generate -f sqlc.yaml
