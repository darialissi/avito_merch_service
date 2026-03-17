## Avito merch shop service

Practicing GO Backend Tech Stack

[Service description](./Backend-trainee-assignment-winter-2025.md)

[API docs](./docs)

### Environment

- prod

- dev

- test

### Configuration

Создать конфиг на основе **.config.[dev|test|prod].example.yml** для необходимого окружения

```
# prod
touch .config.prod.yml
```

```
# dev
touch .config.dev.yml
```

```
# test
touch .config.test.yml
```

### Run

Запуск приложения и его зависимостей [prod]

```
make up-prod
```

Запуск зависимостей [dev]

```
make up-dev
```

Запуск зависимостей [test]

```
make up-test
```

### Dev

После запуска зависимостей в корне проекта сформировался **.env**, где указан CONFIG_PATH (абсолютный путь конфига). Переменную необходимо установить в текущее окружение для локального запуска.

### Test

После запуска зависимостей в корне проекта сформировался **.env**, где указан CONFIG_PATH (абсолютный путь конфига). Переменную необходимо установить в текущее окружение для интеграционного/е2е тестирования.