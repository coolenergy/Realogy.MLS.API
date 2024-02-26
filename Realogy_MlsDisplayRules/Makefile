 all: lint build unit integration stop coverage

lint:
	$(info ************  Go Lint ************)
	@go vet

build:
	$(info ************  Building ************)
	@go build

unit:
	$(info ************  Unit Test ************)
	#@go test -v ./... -count=1

clean:
	$(info ************  Cleaning Containers ************)
	@docker-compose kill
	@docker-compose rm -f

start: clean
	$(info ************  Starting Test Containers ************)
	@docker-compose up --build --force-recreate -d

integration: start
	$(info ************  Integration Test ************)
	sleep 30 	# temporary fix until grpc health check is implemented.
	@go test -tags integration ./... -count=1
	make clean 	# Clean stopped containers

stop:
	$(info ************  Stopping Test Containers ************)
	@docker-compose logs
	@docker-compose down --rmi local

coverage:
	$(info ************  Generating Code Coverage ************)
	#go test -cover ./...

