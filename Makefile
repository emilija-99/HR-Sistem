.PHONY: stop-mongo check-ports clean start restart

# Ports used by the application
PORTS := 27017 8034 8080 9090 3000

stop-mongo:
	@echo "Stopping MongoDB system service..."
	sudo systemctl stop mongod

check-ports:
	@echo "Checking ports: $(PORTS)"
	@if ss -tlnp | grep -E ':(27017|8034|8080|9090|3000)\b'; then \
		echo "ERROR: One or more ports are already in use."; \
		exit 1; \
	else \
		echo "All ports are free."; \
	fi

clean:
	@echo "Removing existing containers..."
	-podman rm -f hr_mongo hr_backend hr_frontend hr_prometheus hr_grafana

start:
	@echo "Starting containers..."
	podman compose up -d

restart: stop-mongo check-ports clean start
	@echo "Application started successfully."
