.PHONY: up stop start clean pg-port-forward pg-port-close rustfs-port-forward rustfs-port-close migrate-up migrate-down migrate-status swagger help

up:
	@docker-compose up -d --build
stop:
	@docker-compose --profile pg-port-forward --profile rustfs-port-forward --profile migrator stop
start:
	@docker-compose start
clean:
	@docker-compose --profile pg-port-forward --profile rustfs-port-forward --profile migrator down --rmi local

pg-port-forward:
	@docker-compose --profile pg-port-forward up -d pg-port-forwarder
pg-port-close:
	@docker-compose rm -f --stop pg-port-forwarder
rustfs-port-forward:
	@docker-compose --profile rustfs-port-forward up -d rustfs-port-forwarder
rustfs-port-close:
	@docker-compose rm -f --stop rustfs-port-forwarder

migrate-up:
	@docker-compose --profile migrator run --rm postgres-migrator ./migrator -command=up
migrate-down:
	@docker-compose --profile migrator run --rm postgres-migrator ./migrator -command=down
migrate-status:
	@docker-compose --profile migrator run --rm postgres-migrator ./migrator -command=status

swagger:
	go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/app/main.go --parseDependency --parseInternal --useStructName

help:
	@echo "Docker Management:"
	@echo "  up                       - Start docker containers"
	@echo "  stop                     - Stop running containers"
	@echo "  start                    - Start stopped containers"
	@echo "  clean                    - Clean current docker containers and images"
	@echo ""
	@echo "Port Forwarding:"
	@echo "  make pg-port-forward     - Start PostgreSQL proxy"
	@echo "  make pg-port-close       - Stop PostgreSQL proxy"
	@echo "  make rustfs-port-forward - Start RustFS proxy"
	@echo "  make rustfs-port-close   - Stop RustFS proxy"
	@echo ""
	@echo "Database Migrations:"
	@echo "  migrate-up               - Apply all pending migrations"
	@echo "  migrate-down             - Rolling back the last migration"
	@echo "  migrate-status           - Show migration status"
	@echo ""
	@echo "Documentation:"
	@echo "  swagger                  - Generate Swagger documentation"
