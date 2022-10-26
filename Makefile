b=tassia

run: build
	@echo "====== BUILD and RUN app ======"
	go mod download && go build -o .bin/$(b) cmd/main.go
	.bin/$(b)

build: clean
	@echo "====== RUN postgres ======"
	docker-compose up --build -d

clean:
	docker-compose stop || true
	rm .bin/$(APP_NAME) || true
