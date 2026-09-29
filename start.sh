#!/usr/bin/env bash

trap 'kill 0' EXIT

(cd backend && air) &

(cd frontend && npm run dev) &

wait