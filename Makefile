.PHONY: swagger swagger-fmt build test test-unit test-integration test-coverage help clean

build:
	@echo "Сборка приложения..."
	@go build -o meemo cmd/meemo/main.go
	@echo "Бинарный файл создан: ./meemo"

test:
	@echo "Запуск всех тестов..."
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "Тесты завершены"
	@echo "Покрытие кода:"
	@go tool cover -func=coverage.out | grep total || true

test-unit:
	@echo "Запуск unit-тестов..."
	@go test -v -race $$(go list ./... | grep -v /tests/integration) -coverprofile=coverage.out
	@echo "Unit-тесты завершены"

test-integration:
	@echo "Запуск интеграционных тестов..."
	@go test -v -race ./tests/integration/...
	@echo "Интеграционные тесты завершены"

test-coverage:
	@echo "Запуск тестов с генерацией отчета о покрытии..."
	@go test -v -race -coverprofile=coverage.out ./... || true
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Отчет о покрытии сохранен в coverage.html"

swagger:
	@echo "Генерация Swagger документации..."
	@export PATH=$$PATH:$$(go env GOPATH)/bin && swag init -g cmd/meemo/main.go -o docs --parseDependency --parseInternal
	@echo "Документация сгенерирована в директории docs/"

swagger-fmt:
	@echo "Форматирование Swagger документации..."
	@export PATH=$$PATH:$$(go env GOPATH)/bin && swag fmt

clean:
	@echo "Очистка..."
	@rm -f meemo
	@rm -f coverage.out coverage.html
	@echo "Очистка завершена"

help:
	@echo "Доступные команды:"
	@echo ""
	@echo "Сборка и запуск:"
	@echo "  make build             - Собрать приложение"
	@echo ""
	@echo "Тестирование:"
	@echo "  make test              - Запустить все тесты"
	@echo "  make test-unit         - Запустить только unit-тесты"
	@echo "  make test-integration  - Запустить интеграционные тесты"
	@echo "  make test-coverage     - Запустить тесты с HTML отчетом о покрытии"
	@echo ""
	@echo "Документация:"
	@echo "  make swagger           - Сгенерировать Swagger документацию"
	@echo "  make swagger-fmt       - Отформатировать Swagger аннотации"
	@echo ""
	@echo "Прочее:"
	@echo "  make clean             - Удалить сгенерированные файлы"
	@echo "  make help              - Показать эту справку"

