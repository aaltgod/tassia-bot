APP_NAME=tassia

run: build
	@echo "====== BUILD and RUN app ======"
	go mod download && go build -o $(APP_NAME) .
	./$(APP_NAME)

build: clean
	@echo "====== RUN postgres ======"
	docker-compose up --build -d

clean:
	docker-compose stop || true
	rm ./$(APP_NAME) || true
