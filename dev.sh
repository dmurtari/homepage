#!/usr/bin/env bash

trap 'kill 0' EXIT

air -c .air.toml &
air -c .air.webhooks.toml &
wait
