## run aplication
.PHONY: run
run: 
	@echo "starting pictures-api"
	@go run main.go
docker-run:
	@docker-compose up