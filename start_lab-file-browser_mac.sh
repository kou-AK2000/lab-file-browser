#!/bin/bash

PORT=8080
HOST="user@ip"

if lsof -i :$PORT > /dev/null; then
    echo "Port $PORT already in use."
else
    ssh -f -N -L ${PORT}:localhost:${PORT} ${HOST}
fi

open http://localhost:${PORT}