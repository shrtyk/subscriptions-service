[![codecov](https://codecov.io/gh/shrtyk/subscriptions-service/graph/badge.svg?token=ULNNWGVP52)](https://codecov.io/gh/shrtyk/subscriptions-service)
[![Go Report Card](https://goreportcard.com/badge/github.com/shrtyk/subscriptions-service)](https://goreportcard.com/report/github.com/shrtyk/subscriptions-service)

# Subscriptions Service

REST-сервис для управления подписками пользователей.

## Быстрый старт

1.  **Настройте файл окружения:**

    Создайте файл `.env` по примеру `.env_example`. Файл должен содержать все необходимые переменные окружения для запуска приложения.

    Для использования дефолтных значений из `.env_example` можно использовать команду:

    ```sh
    make setup
    ```

    или можно сделать это вручную:

    ```sh
    cp .env_example .env
    ```

2.  **Соберите и запустите приложение:**

    Эта команда соберет Docker-образы и запустит все сервисы (приложение и базу данных PostgreSQL) в фоновом режиме.

    ```sh
    make docker/up
    ```

    API будет доступен по адресу `http://localhost:8080`.

## Доступные команды

Небольшой список `make` команд доступных в проекте:

### Docker и Приложение

- `make docker/up`: Собрать и запустить все сервисы в фоновом режиме
- `make docker/down`: Остановить и удалить все сервисы и их тома (volumes)
- `make app/run`: Запустить только сервис приложения. Он будет автоматически остановлен при выходе
- `make db/start`: Запустить только базу данных и применить миграции
- `make db/stop`: Остановить контейнер с базой данных

### База данных и Миграции

- `make psql`: Подключиться к базе данных PostgreSQL внутри контейнера с помощью `psql`
- `make migrations/new NAME=<migration_name>`: Создать новый файл миграции SQL
- `make migrations/up`: Применить все доступные миграции
- `make migrations/down`: Откатить последнюю миграцию
- `make migrations/down-all`: Откатить все миграции
- `make migrations/status`: Показать статус всех миграций

### Разработка и Тестирование

- `make unit-tests/run`: Запустить все юнит-тесты и сгенерировать отчет о покрытии в `coverage.out`
- `make linter/run`: Запустить линтер `golangci-lint`
- `make mocks/generate`: Сгенерировать моки для интерфейсов с помощью `mockery`
- `make dto/generate`: Сгенерировать DTO и серверный код из спецификации OpenAPI (`api/swagger.yaml`)

## Архитектура и структура проекта

Проект построен с использованием _Hexagonal architecture_. Краткое описание директорий:

- `/api` Спецификация OpenAPI
- `/cmd/app` Точка входа в приложение
- `/internal` Вся основная логика приложения
  - `/api/http` Код, связанный с HTTP-слоем: обработчики запросов (хендлеры), DTO, роутинг и middleware
  - `/core` Ядро бизнес-логики
    - `/domain` Основные модели данных (сущности)
    - `/ports` "Порты" архитектуры: интерфейсы для репозиториев, сервисов и других внешних зависимостей
    - `/subservice` Реализация бизнес-логики (сервисный слой)
  - `/infra` Реализация "адаптеров" для внешних систем
    - `/postgres` Реализация репозитория для работы с PostgreSQL
  - `/config` Загрузка и валидация конфигурации
- `/pkg` Пакеты, которые можно безопасно использовать в других проектах
- `/migrations` Файлы миграций базы данных
