# Линтинг

В проекте статический анализ настроен отдельно для backend и frontend и автоматически запускается в CI.

Цель линтеров — находить ошибки до code review и поддерживать единые правила качества кода.

## Backend

Для Go используется **golangci-lint**.

Конфигурация: `/.golangci.yml`

Запуск:

```bash
make lint
```

Полная проверка backend:

```bash
make check
```

### Основные правила

| Линтер                                  | Зачем нужен                                                   |
| --------------------------------------- | ------------------------------------------------------------- |
| `errcheck`, `errorlint`, `nilerr`       | Корректная обработка ошибок                                   |
| `contextcheck`, `containedctx`, `noctx` | Правильная работа с `context.Context`, cancellation и timeout |
| `bodyclose`                             | Предотвращение утечек HTTP-соединений                         |
| `gosec`                                 | Поиск потенциально небезопасных конструкций                   |
| `govet` + `shadow`                      | Поиск распространённых ошибок Go и shadowing переменных       |
| `cyclop`                                | Контроль чрезмерной сложности функций                         |
| `dupl`                                  | Обнаружение дублирования кода                                 |
| `gocritic`                              | Дополнительные проверки корректности и читаемости             |
| `unparam`, `unconvert`, `wastedassign`  | Поиск лишнего и неиспользуемого кода                          |
| `nolintlint`                            | Контроль использования `//nolint`                             |

Для `//nolint` обязательно указываются конкретный линтер и причина:

```go
//nolint:gosec // deterministic test fixture
```

Необъяснённый `//nolint` запрещён.

### Форматирование

Используются:

* `gofumpt` — единый формат Go-кода;
* `gci` — единый порядок и группировка импортов.

Для тестов отключены `cyclop`, `dupl` и `gosec`, так как test setup и fixtures естественно содержат больше повторений и служебной логики.


## Frontend

Для React + TypeScript используются:

* **ESLint**
* **typescript-eslint**
* **eslint-plugin-react-hooks**
* **eslint-plugin-react-refresh**
* **Prettier**

Конфигурация: `/frontend/eslint.config.js`

Запуск:

```bash
cd frontend
npm run lint
npm run build
npm run format
```

### Основные правила

| Правило                                    | Зачем нужно                                           |
| ------------------------------------------ | ----------------------------------------------------- |
| `typescript-eslint/recommendedTypeChecked` | Type-aware проверка TypeScript-кода                   |
| `react-hooks/recommended`                  | Контроль корректного использования React Hooks        |
| `consistent-type-imports`                  | Явное разделение импортов типов и runtime-значений    |
| `eqeqeq`                                   | Защита от неявного преобразования типов при сравнении |
| `no-console`                               | Предупреждение о случайно оставленных debug-логах     |
| `react-refresh/only-export-components`     | Корректная работа Fast Refresh                        |


**Prettier** отвечает за форматирование кода, **ESLint** — за корректность и потенциально проблемные конструкции.



## CI

Backend:

```bash
make generate
make lint-config
make lint
```

Frontend:

```bash
npm ci
npm run lint
npm run build
```

Ошибки линтинга и TypeScript-проверок блокируют прохождение CI.

## Принцип выбора правил

Мы включили не все доступные проверки, а правила, которые:

1. обнаруживают реальный класс ошибок;
2. повышают безопасность или надёжность;
3. поддерживают единый стандарт команды без лишнего шума.
