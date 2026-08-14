# CI/CD

В проекте настроен автоматический CI/CD pipeline на **GitHub Actions**.

Цель — проверять каждое изменение до попадания в production и автоматически разворачивать проверенную версию приложения.

## Pull Request / Develop

При изменениях в `develop` и Pull Request запускаются автоматические проверки.

### Backend

* генерация кода;
* проверка конфигурации `golangci-lint`;
* статический анализ;
* unit-тесты;
* integration-тесты.

Основные команды:

```bash
make generate
make lint-config
make lint
make test
make test-integration
```

### Frontend

* установка зависимостей через `npm ci`;
* ESLint;
* TypeScript type checking;
* production build.

```bash
npm ci
npm run lint
npm run build
```

Таким образом, ошибки компиляции, линтинга и тестов обнаруживаются до merge изменений.


## Production release

Production pipeline запускается после попадания изменений в `main`.

Процесс релиза:

1. Запускаются integration-тесты.
2. Запускаются E2E-тесты основного пользовательского сценария.
3. Собирается Docker image backend.
4. Собирается Docker image frontend.
5. Образы публикуются в **GitHub Container Registry (GHCR)**.
6. Запускается production deployment.

Так deployment выполняется только после успешного прохождения проверок.


## E2E-тестирование

Перед production deployment проверяется полный пользовательский сценарий:

1. получение профиля;
2. создание recap;
3. получение персонального recap;
4. создание публичной ссылки;
5. получение shared recap;
6. проверка отсутствия приватных данных в публичной версии.

E2E-тесты работают с реальным PostgreSQL через Testcontainers.


## Почему pipeline разделён

### Pull Request / Develop

Основная задача — быстро проверить качество изменений:

* линтинг;
* сборка;
* unit-тесты;
* integration-тесты.

### Main

Основная задача — убедиться в готовности версии к production:

* integration-тесты;
* E2E-тесты;
* сборка Docker images;
* публикация образов;
* deployment.

Так проверки разработки отделены от процесса выпуска новой версии.


## Локальная проверка

Полная проверка backend:

```bash
make check
```

Frontend:

```bash
cd frontend
npm run lint
npm run build
```

Рекомендуется запускать эти проверки перед созданием Pull Request.


## Секреты и конфигурация

Секреты и production-конфигурация не хранятся в Git.

Настройки приложения передаются через:

* environment variables;
* GitHub Actions Secrets;
* отдельные production-конфигурации для deployment.

Пример локальной конфигурации находится в `.env.example`.


## Rollback

Для production deployment предусмотрен отдельный rollback-сценарий.

Скрипт:

```text
deploy/rollback-release.sh
```

Он позволяет откатить deployment при возникновении проблем после выпуска новой версии.


## Общая схема

```text
Pull Request / develop
        │
        ├── Backend lint
        ├── Unit tests
        ├── Integration tests
        └── Frontend lint + build
                │
                ▼
              main
                │
        ├── Integration tests
        ├── E2E tests
        ├── Build Docker images
        ├── Publish to GHCR
        └── Production deployment
```
