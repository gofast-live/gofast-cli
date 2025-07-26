#!/bin/bash

cd "$(dirname "$0")"

openssl genpkey -algorithm Ed25519 -out private.pem
openssl pkey -in private.pem -pubout -out public.pem

cp private.pem ../app/service-core/private.pem
cp public.pem ../app/service-core/public.pem
cp public.pem ../app/service-admin/public.pem

if [ -d "../service-client" ]; then
    cp public.pem ../service-client/public.pem
fi

rm private.pem public.pem
