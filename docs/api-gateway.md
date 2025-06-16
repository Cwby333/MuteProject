# API Gateway

API Gateway выступает единой точкой входа для всех клиентских запросов к микросервисам mute. Он маршрутизирует входящие запросы к соответствующим микросервисам и объединяет их ответы.

## Архитектура API Gateway

API Gateway состоит из двух основных компонентов:
- **Nginx**: работает как обратный прокси, принимая запросы от клиентов
- **Go-сервис**: обрабатывает маршрутизацию запросов к соответствующим микросервисам

## Конфигурация Nginx

Nginx выступает в роли прокси-сервера, который принимает все запросы и перенаправляет их на API Gateway:

```nginx
upstream api-gateway {
    server api-gateway:3000;
}

server {
    listen 80;

    client_max_body_size 20M; # Разрешает загрузку файлов до 20МБ

    access_log /var/log/nginx/access.log;
    error_log /var/log/nginx/error.log;

    location /static/ {
        alias /usr/share/nginx/html/static/;
    }

    location /stream {
        proxy_pass http://api-gateway;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        # Отключаем кэширование
        proxy_no_cache 1;
        proxy_cache_bypass 1;
        add_header Cache-Control "no-cache, no-store, must-revalidate";
        add_header Pragma "no-cache";
        add_header Expires "0";
    }

    location /play/ {
        proxy_pass http://api-gateway;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        # Отключаем кэширование
        proxy_no_cache 1;
        proxy_cache_bypass 1;
        add_header Cache-Control "no-cache, no-store, must-revalidate";
        add_header Pragma "no-cache";
        add_header Expires "0";
    }

    location / {
        proxy_pass http://api-gateway;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_no_cache 1;
        proxy_cache_bypass 1;
        add_header Cache-Control "no-cache, no-store, must-revalidate";
        add_header Pragma "no-cache";
        add_header Expires "0";
    }
}
```

Конфигурация Nginx:
- Прослушивает порт 80
- Разрешает загрузку файлов размером до 20MB
- Настраивает логирование в стандартные файлы
- Определяет специальные обработчики для `/static/`, `/stream`, `/play/` и общий обработчик для всех остальных маршрутов
- Отключает кэширование для всех запросов, чтобы всегда получать свежие данные

## Маршрутизация запросов

API Gateway маршрутизирует запросы по следующей схеме:

### Сервис пользователей (users-service)

| Метод | Путь | Назначение |
|-------|------|------------|
| POST | /user/register | Регистрация пользователя |
| POST | /user/login | Аутентификация пользователя |
| POST | /user/logout | Выход из системы |
| POST | /user/refresh | Обновление токенов |
| GET | /user/get | Получение информации о текущем пользователе |
| GET | /user/{userId} | Получение информации о пользователе по ID |
| GET | /user/all | Получение списка всех пользователей (только для админов) |
| DELETE | /user/delete | Удаление пользователя |
| PUT | /user/update | Обновление данных пользователя |
| POST | /user/track/favorite | Добавление трека в избранное |
| DELETE | /user/track/favorite | Удаление трека из избранного |

### Сервис музыки (music-service)

| Метод | Путь | Назначение |
|-------|------|------------|
| GET | /tracks | Получение списка всех треков |
| GET | /tracks/{userId} | Получение избранных треков пользователя |
| POST | /track/{userId} | Добавление нового трека (загрузка файла) |
| DELETE | /track/{trackId} | Удаление трека |
| POST | /track/{trackId} | Обновление трека |

### Плеер (заглушки)

| Метод | Путь | Назначение |
|-------|------|------------|
| GET | /stream | Потоковое воспроизведение аудио |
| GET | /play/{trackName} | Запуск воспроизведения трека |
| GET | /pause | Приостановка воспроизведения |
| GET | /resume | Возобновление воспроизведения |

## Конфигурация и переменные окружения

API Gateway использует следующие переменные окружения для настройки:

- `USERS_SERVICE_URL`: URL сервиса пользователей (по умолчанию "http://users-service:8888")
- `MUSIC_SERVICE_URL`: URL сервиса музыки (по умолчанию "http://music-service:8080")

## Dockerfile для API Gateway

```dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o api-gateway .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/api-gateway .

EXPOSE 3000

CMD ["./api-gateway"]
```

Этот Dockerfile:
1. Использует golang:1.24-alpine как базовый образ для сборки
2. Копирует файл go.mod и загружает зависимости
3. Копирует весь исходный код и собирает бинарный файл
4. Создает минимальный образ на базе alpine
5. Копирует бинарный файл из образа для сборки
6. Открывает порт 3000
7. Запускает API Gateway

## Dockerfile для Nginx

```dockerfile
FROM nginx:alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
```

Этот Dockerfile:
1. Использует nginx:alpine как базовый образ
2. Копирует файл конфигурации nginx.conf
3. Открывает порт 80

## Особенности работы API Gateway

1. **Проксирование запросов**: API Gateway проксирует запросы к соответствующим микросервисам, объединяя их в единый API.

2. **Сохранение заголовков**: При проксировании запросов сохраняются важные заголовки, такие как Host, X-Real-IP, X-Forwarded-For и X-Forwarded-Proto.

3. **CORS**: API Gateway настроен для поддержки Cross-Origin Resource Sharing (CORS), что позволяет безопасно обрабатывать запросы от фронтенд-приложений, размещенных на разных доменах.
