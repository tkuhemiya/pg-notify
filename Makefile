SHELL := /bin/bash

PYTHON := $(strip $(or $(shell command -v python3), $(shell command -v python)))
DEMO_PORT := $(if $(PYTHON),$(shell $(PYTHON) -c "import json, pathlib; cfg=pathlib.Path('pg-notify.cfg'); data=json.loads(cfg.read_text()) if cfg.exists() else {}; print(data.get('port', 8080))"),8080)
DB_CONTAINER ?= pg-notify-db
SIMULATION_BATCH ?= 50
SIMULATION_ROUNDS ?= 4
SIMULATION_INTERVAL ?= 1
DEMO_LOG ?= /tmp/pg-notify-demo.log

.DEFAULT_GOAL := demo

.PHONY: demo ensure-db simulate-burst

demo: ensure-db
	@printf "Dashboard available at http://localhost:%s\n" $(DEMO_PORT)
	@bash -c '\
		set -euo pipefail; \
		go run . >"$(DEMO_LOG)" 2>&1 & \
		server_pid=$$!; \
		trap "kill $$server_pid" EXIT; \
		sleep 2; \
		$(MAKE) simulate-burst; \
		echo; \
		echo "Server log: $(DEMO_LOG)"; \
		echo "Press Ctrl+C to stop the server."; \
		wait $$server_pid; \
	'

ensure-db:
	@command -v docker >/dev/null 2>&1 || { echo "docker CLI not found; demo needs Docker to seed activity."; exit 1; }
	@if ! docker ps --format '{{.Names}}' | grep -wqx "$(DB_CONTAINER)"; then \
		echo "Container \"$(DB_CONTAINER)\" is not running."; \
		echo "Launch it via the instructions in db/README.md (for example: docker run --name $(DB_CONTAINER) ...)."; \
		exit 1; \
	fi

simulate-burst:
	@echo "Simulating $(SIMULATION_ROUNDS) bursts of $(SIMULATION_BATCH) inserts..."
	@r=0; \
	while [ $$r -lt $(SIMULATION_ROUNDS) ]; do \
		r=$$((r + 1)); \
		echo "  burst $$r/$(SIMULATION_ROUNDS)"; \
		i=0; \
		while [ $$i -lt $(SIMULATION_BATCH) ]; do \
			i=$$((i + 1)); \
			echo "    request $$i/$(SIMULATION_BATCH)"; \
			docker exec -i $(DB_CONTAINER) psql -U postgres -d pg_notify -c "SELECT simulate_orders(1);" >/dev/null; \
			if [ $$i -lt $(SIMULATION_BATCH) ]; then \
				sleep 0.1; \
			fi; \
		done; \
	done


