APP_NAME=tassia

run: build
	@echo "====== BUILD and RUN app ======"
	go mod download && go build -o .bin/$(APP_NAME) cmd/main.go
	.bin/$(APP_NAME)

build: clean
	@echo "====== RUN postgres ======"
	docker-compose up --build -d

clean:
	docker-compose stop || true
	rm .bin/$(APP_NAME) || true
