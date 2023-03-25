help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
%:
	@:

create_db: ## Creates db for this project.
	docker run --name=db -e POSTGRES_PASSWORD='123456' -p 5432:5432 -d --rm postgres
	sleep 2
	docker exec -it db createdb -U postgres documentize_db

run: ## Runs the project.
	go run ./cmd/main.go

quick_start: ## Creates db and starts the project.
	make create_db
	make run
