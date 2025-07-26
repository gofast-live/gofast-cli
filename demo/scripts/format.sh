#!/bin/bash

cd "$(dirname "$0")"
cd ..

if [ -d "service-client" ]; then
    echo ""
    echo "Formatting client"
    cd service-client
    npm run format
    cd ..
fi

if [ -d "service-svelte" ]; then
    echo ""
    echo "Formatting svelte"
    cd service-svelte
    npm run format
    cd ..
fi

if [ -d "service-next" ]; then
    echo ""
    echo "Formatting next"
    cd service-next
    npm run format
    cd ..
fi

if [ -d "service-vue" ]; then
    echo ""
    echo "Formatting vue"
    cd service-vue
    npm run format
    cd ..
fi

