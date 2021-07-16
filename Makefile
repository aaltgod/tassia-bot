APP_NAME=tassia
DB_NAME=storage

run: build

	@echo "====== BUILD and RUN app ======"
	go mod download && go build -o $(APP_NAME) .
	./$(APP_NAME)

build: clean
	@echo "====== RUN postgres ======"
	docker run -d --name $(DB_NAME) --rm -p 5432:5432 postgres:10.5


clean:
	docker stop $(DB_NAME) || true
	rm ./$(APP_NAME) || true
