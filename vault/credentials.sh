#!/bin/bash

if [ -f "/vault/config/.init_data" ]; then
    cat /vault/config/.init_data
else
    echo "The vault is not initialized !"
fi
