# Задание 

Требуется написать сервис расчета параметров ипотеки (`ипотечный калькулятор`). 
Рассчитываемые параметры:
   - процентная ставка, исходя из запрошенной `программы кредитования`
   - сумма кредита
   - аннуитетный ежемесячный платеж
   - переплата за весь срок кредита
   - дата последнего платежа

Все расчеты требуется сохранять в локальном `кэше`.

## Программы кредитования
Есть 3 программы кредитования. Каждая из них предполагает свою годовую процентную ставку:

1. Программа для корпоративных клиентов. Годовая процентная ставка по кредиту - `8%`.
2. Военная ипотека. Годовая процентная ставка по кредиту - `9%`.
3. Базовая программа. Годовая процентная ставка по кредиту - `10%`.

Для каждой программы `первоначальный взнос должен быть не ниже 20%` от стоимости объекта.

Программа кредита указывается в запросе (входном JSON).

## Запрос (входной JSON)

Запрос на сервис:
```json
{
    "object_cost": 5000000,     // стоимость объекта
    "initial_payment": 1000000, // первоначальный взнос
    "months": 240,              // срок
    "program": {                // блок программы кредита
        "salary": true,         // программа для корпоративных клиентов
        "military": true,       // военная ипотека
        "base": true            // базовая программа
    }
}
```

Для указания программы требуется передавать только 1 поле. Например:
```json
{
    "object_cost": 5000000,
    "initial_payment": 1000000,
    "months": 240,
    "program": {
        "salary": true
    }
}
```

## Эндпоинты

Требуется реализовать 2 эндпоинта:
1. `/execute` - расчет ипотеки (POST).
2. `/cache` - получение всех рассчитанных ипотек из кэша (GET). 

## /execute
В качестве входных данных эндпоинт принимает JSON:
```json
{
    "object_cost": 5000000,
    "initial_payment": 1000000,
    "months": 240,
    "program": {
        "salary": true
    }
}
```
В качестве ответа возвращается JSON и `status code: 200`:
```json
{
   "result": {
      "params": {                           // запрашиваемые параметры кредита
         "object_cost": 5000000,
         "initial_payment": 1000000,                
         "months": 240
      },
      "program": {                          // программа кредита
         "salary": true
      },
      "aggregates": {                       // блок с агрегатами
         "rate": 8,                         // годовая процентная ставка
         "loan_sum": 4000000,               // сумма кредита
         "monthly_payment": 33458,          // аннуитетный ежемесячный платеж
         "overpayment": 4029920,            // переплата за весь срок кредита
         "last_payment_date": "2044-02-18"  // последняя дата платежа
      }
   }
}
```

В случае, если не выбрана ни одна из программ (во входном JSON все поля программы `false`), то требуется возвращать `status code: 400` и ошибку формата:
```json
{
    "error": "choose program"
}
```

В случае, если выбрана более, чем одна программа, то требуется возвращать `status code: 400` и ошибку формата:
```json
{
   "error": "choose only 1 program"
}
```

В случае, если первоначальный взнос ниже 20% от стоимости объекта, то требуется возвращать `status code: 400` и ошибку формата:
```json
{
    "error": "the initial payment should be more"
}
```


Результат расчета кредита требуется сохранять в `кэш`.

## /cache
Сервис возвращает массив из рассчитанных кредитов и `status code: 200`:
```json
[
   {
      "id": 0, // id расчета в кэше
      "params": {
         "object_cost": 5000000,
         "initial_payment": 1000000,
         "months": 240
      },
      "program": {
         "salary": true
      },
      "aggregates": {
         "rate": 8,
         "loan_sum": 4000000,
         "monthly_payment": 33458,
         "overpayment": 4029920,
         "last_payment_date": "2044-02-18"
      }
   },
   {
      "id": 1,
      "params": {
         "object_cost": 8000000,
         "initial_payment": 2000000,
         "months": 200
      },
      "program": {
         "military": true
      },
      "aggregates": {
         "rate": 9,
         "loan_sum": 6000000,
         "monthly_payment": 58019,
         "overpayment": 5603800,
         "last_payment_date": "2040-10-18"
      }
   },
   {
      "id": 2,
      "params": {
         "object_cost": 12000000,
         "initial_payment": 3000000,
         "months": 120
      },
      "program": {
         "base": true
      },
      "aggregates": {
         "rate": 10,
         "loan_sum": 9000000,
         "monthly_payment": 118936,
         "overpayment": 5272320,
         "last_payment_date": "2034-02-18"
      }
   }
]
```
Если кэш пустой, то требуется возвращать `status code: 400` и ошибку формата:
```json
{
   "error": "empty cache"
}
```

## Кэш
Требуется сохранять рассчитанные кредиты и отдавть их по запросу на /cache.
Кэш должен быть реализован в `RAM`, без использования сторонних БД.

## Middleware
Требуется реализовать middleware, который будет выводить в консоль информацию о запросе:
   - `status_code` - http код запроса
   - `duration` - время работы эндпоинта (ns)
   ```text
   2022/02/17 19:26:52 status_code: 200, duration: 243042 ns
   2022/02/17 19:26:53 status_code: 400, duration: 18875 ns
   ```
Требование обязательно, даже если используемый вами web-framework предоставляет такой функционал "из коробки".

## Требования и ограничения
1. Можно использовать любой web-framework.
2. Сервис должен иметь настраиваемую конфигурацию. 
Порт, на котором поднимается сервис, должен указываться в файле `config.yml` и именть значение `8080`.
    ```yaml
    port: 8080
    ```
3. Покрытие unit-тестами > 80%.
4. Код должен проходить проверку линтером `golangci-lint`. Конфигурация:
   ```yaml
   run:
     concurrency: 4
     timeout: 1m
     tests: false
     issues-exit-code: 0
     go: '1.21'
   
   output:
     format: colored-line-number
   
   issues:
     max-issues-per-linter: 1000
     max-same-issues: 1000
     exclude-use-default: false
   
   linters-settings:
     prealloc:
       for-loops: true
   
     stylecheck:
       go: "1.21"
       checks: ["all"]
   
     staticcheck:
       go: "1.21"
       checks: ["all"]
   
     nolintlint:
       allow-unused: true
       require-explanation: true
       require-specific: true
   
     gosimple:
       go: "1.21"
       checks: ["all"]
   
     gocyclo:
       min-complexity: 20
   
     gocritic:
       disabled-checks:
         - hugeParam
       enabled-tags:
         - diagnostic
         - style
         - performance
         - experimental
         - opinionated
   
     dupl:
       threshold: 70
   
     dogsled:
       max-blank-identifiers: 2
   
     errcheck:
       check-type-assertions: true
       check-blank: true
   
     govet:
       check-shadowing: true
       enable-all: true
       shadow:
         strict: true
   
     funlen:
       lines: 60
       statements: 60
       ignore-comments: true
   
   linters:
     disable-all: true
     enable:
       - bodyclose
       - dogsled
       - dupl
       - errcheck
       - forbidigo
       - funlen
       - gochecknoinits
       - goconst
       - gocritic
       - gocyclo
       - gofmt
       - gomodguard
       - revive
       - gosec
       - gosimple
       - govet
       - godot
       - ineffassign
       - misspell
       - nakedret
       - nolintlint
       - exportloopref
       - prealloc
       - staticcheck
       - stylecheck
       - typecheck
       - unconvert
       - unparam
       - unused
       - whitespace
   
     presets:
       - comment
       - error
   ```
5. Все внешние зависимости должны быть "завендорины".
6. Проект содержит `Dockerfile`. "Вес" образа должен быть не более `30 MB`.
7. Проект содержит `Makefile`, в котором указаны команды:
    - запуска тестов
    - запуска линтеров
    - сборки образа
    - запуска контейнера
    - остановки и удаления контейнера
8. Данный сервис должен храниться на github.com