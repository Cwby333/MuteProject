# Frontend

Фронтенд приложения Mute представляет собой набор статических HTML, CSS и JavaScript файлов, которые взаимодействуют с API Gateway для получения и отображения данных.

## Архитектура фронтенда

Фронтенд использует чистый HTML/CSS/JavaScript без применения фреймворков, что делает его легким и быстрым. Основные технологии:

- **HTML5** для структуры страниц
- **CSS3** и **Bootstrap 5** для стилизации
- **JavaScript (ES6+)** для интерактивности

## Структура проекта

```
frontend/
├── src/                  # Исходный код HTML и JavaScript
│   ├── index.html        # Главная страница с плейлистом треков
│   ├── login.html        # Страница входа
│   ├── register.html     # Страница регистрации
│   ├── upload-track.html # Страница загрузки новых треков
│   ├── favorites.html    # Страница с избранными треками
│   ├── profile.html      # Страница профиля пользователя
│   └── js/               # JavaScript модули
│       ├── auth.js       # Модуль авторизации
│       └── favorites.js  # Модуль для работы с избранными треками
├── styles/               # CSS-стили
│   └── styles.css        # Основные стили приложения
└── static/               # Статические ресурсы (изображения, иконки)
    └── images/           # Изображения для интерфейса
```

## Основные функции и страницы

### Главная страница (index.html)

Отображает список всех доступных треков с возможностью:
- Воспроизведения трека
- Добавления трека в избранное
- Просмотра информации об авторе трека

Функционал:
1. Загрузка списка треков с API (`GET /tracks`)
2. Получение информации об авторах треков (`GET /user/{userId}`)
3. Встроенный плеер с элементами управления

### Система авторизации (login.html, register.html)

Обеспечивает регистрацию, вход и управление сессией:
- Регистрация новых пользователей
- Вход в систему
- Хранение токена авторизации в localStorage
- Обновление UI в зависимости от статуса авторизации

### Избранные треки (favorites.html)

Позволяет пользователю:
- Просматривать список избранных треков
- Управлять избранными треками (добавлять/удалять)

### Загрузка треков (upload-track.html)

Дает возможность загружать новые треки:
- Загрузка аудиофайла
- Загрузка обложки
- Указание названия трека
- Автоматическая привязка к текущему пользователю

### Профиль пользователя (profile.html)

Отображает информацию о пользователе и позволяет:
- Просматривать загруженные треки
- Редактировать данные профиля
- Управлять аккаунтом

## Ключевые JavaScript модули

### auth.js

Модуль для управления аутентификацией:

```javascript
// Проверка статуса авторизации
function isLoggedIn() {
  const userData = localStorage.getItem('user');
  if (!userData) return false;
  
  try {
    const user = JSON.parse(userData);
    return user.isLoggedIn === true;
  } catch (e) {
    return false;
  }
}

// Получение информации о пользователе
function getUserData() {
  const userData = localStorage.getItem('user');
  if (!userData) return null;
  
  try {
    return JSON.parse(userData);
  } catch (e) {
    return null;
  }
}

// Выход из системы
function logout() {
  localStorage.removeItem('user');
  
  fetch('http://localhost/user/logout', {
    method: 'POST',
  }).catch(err => console.error('Ошибка при выходе из системы:', err));
  
  window.location.href = 'login.html';
}

// Обновление UI в зависимости от статуса авторизации
function updateAuthUI() {
  const authSection = document.querySelector('.auth-section');
  if (!authSection) return;
  
  if (isLoggedIn()) {
    const userData = getUserData();
    authSection.innerHTML = `
      <div class="dropdown">
        <a href="#" class="d-flex align-items-center text-decoration-none dropdown-toggle" id="userDropdown" data-bs-toggle="dropdown" aria-expanded="false">
          <img src="../static/images/247319.png" alt="User" class="user-avatar" style="height: 40px; margin-right: 10px;">
          <span class="username">${userData.username}</span>
        </a>
        <ul class="dropdown-menu dropdown-menu-end" aria-labelledby="userDropdown">
          <li><a class="dropdown-item" href="profile.html">Мой профиль</a></li>
          <li><a class="dropdown-item" href="settings.html">Настройки</a></li>
          <li><hr class="dropdown-divider"></li>
          <li><a class="dropdown-item" href="#" onclick="logout(); return false;">Выход</a></li>
        </ul>
      </div>
    `;
  } else {
    authSection.innerHTML = `
      <a href="login.html" class="btn btn-sm btn-outline-light">Войти</a>
      <a href="register.html" class="btn btn-sm btn-success ms-2">Регистрация</a>
    `;
  }
}
```

### favorites.js

Модуль для работы с избранными треками:

- Загрузка избранных треков пользователя
- Добавление трека в избранное
- Удаление трека из избранного
- Обновление UI для отображения статуса избранных треков

## Взаимодействие с API

Фронтенд взаимодействует с бэкендом через API Gateway, используя следующие эндпоинты:

| Эндпоинт | Метод | Описание |
|----------|-------|----------|
| `/user/register` | POST | Регистрация нового пользователя |
| `/user/login` | POST | Вход в систему |
| `/user/logout` | POST | Выход из системы |
| `/tracks` | GET | Получение списка всех треков |
| `/tracks/{userId}` | GET | Получение избранных треков пользователя |
| `/user/{userId}` | GET | Получение информации о пользователе |
| `/track/{userId}` | POST | Загрузка нового трека |
| `/user/track/favorite` | POST | Добавление трека в избранное |
| `/user/track/favorite` | DELETE | Удаление трека из избранного |

## Безопасность фронтенда

1. **Хранение токенов**: Токены авторизации хранятся в localStorage браузера
2. **Проверка авторизации**: Доступ к защищенным функциям (загрузка треков, управление избранным) проверяется на стороне клиента перед отправкой запроса
3. **CORS**: Фронтенд отправляет запросы только на разрешенные домены

## Пользовательский интерфейс

1. **Адаптивный дизайн**: Интерфейс корректно отображается на устройствах с разными размерами экранов
2. **Темная тема**: Приложение использует темную цветовую схему для комфортного использования
3. **Интуитивная навигация**: Логичное расположение элементов и простая навигация по разделам

## Плеер

Встроенный музыкальный плеер имеет следующие функции:
- Воспроизведение/пауза
- Переход к следующему/предыдущему треку
- Отображение информации о текущем треке
- Отображение обложки трека

Поскольку фронтенд состоит только из статических файлов, он не требует запуска серверных процессов и может быть легко размещен на любом хостинге для статических сайтов.