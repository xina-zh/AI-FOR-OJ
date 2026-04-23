#!/usr/bin/env sh
set -eu

docker compose up -d mysql app frontend

printf '%s\n' 'AI-For-OJ is starting:'
printf '%s\n' '  Backend:  http://127.0.0.1:8080'
printf '%s\n' '  Frontend: http://127.0.0.1:5188'
