## Avito merch shop service 🧝🏻‍♀️

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

После запуска зависимостей в корне проекта сформировался **.env**, где указан CONFIG_PATH (абсолютный путь конфига). Переменную необходимо установить в текущее окружение для локального запуска и накатки миграций.

### Test

После запуска зависимостей в корне проекта сформировался **.env**, где указан CONFIG_PATH (абсолютный путь конфига). Переменную необходимо установить в текущее окружение для интеграционного/е2е тестирования.

. . .

```
# unit
make usecase-test
```

### Samples

#### /api/register
```
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"itsme","password":"OMGsecret007"}'
```

```
{"id":"1e353f05-5ba4-4b2a-8b66-93b3f81e2885","username":"itsme","coins":1000,"created_at":"2026-03-22T21:34:51.498693-07:00"}
```

```
{"errors":"Username already exists"}
```

#### /api/auth

```
curl -X POST http://localhost:8080/api/auth \
  -H "Content-Type: application/json" \
  -d '{"username":"itsme","password":"OMGsecret007"}'
```

```
{"access_token":"eyJhbGciOiJIUzI1NiIs...","refresh_token":"eyJhbGciOiJIUzI1NiIs..."}
```

```
{"errors":"Incorrect password"}
```

#### /api/buy/{item}

```
curl -X POST http://localhost:8080/api/buy/book \
  -H "Content-Type: application/json" \
  -d '{"quantity":2}' \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." 
```

```
{"errors":"Item not found"}
```

```
{"errors":"Not enough coins"}
```

#### /api/sendCoin

```
  curl -X POST http://localhost:8080/api/sendCoin \
  -H "Content-Type: application/json" \
  -d '{"toUser":"itsnotme", "amount":100}' \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." 
```

```
{"errors":"You cannot send coins to yourself"}
```

```
{"errors":"Not enough coins"}
```

#### /api/info

```
 curl -X GET http://localhost:8080/api/info \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." 
```

```
{"coins":300,"inventory":[{"type":"book","quantity":2},{"type":"hoodie","quantity":1}],"coinHistory":{"sent":[{"toUser":"83d19252-e491-4d97-9e4f-735f30a876ca","amount":500}],"received":[{"fromUser":"83d19252-e491-4d97-9e4f-735f30a876ca","amount":200}]}}
```
