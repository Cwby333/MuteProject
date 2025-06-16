// Функция для проверки, вошел ли пользователь в систему
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
  
  // Опционально: можно сделать запрос на сервер для сброса сессии
  fetch('http://localhost/user/logout', {
    method: 'POST',
  }).catch(err => console.error('Ошибка при выходе из системы:', err));
  
  // Перенаправляем на страницу логина
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

// Запускаем функцию при загрузке страницы
document.addEventListener('DOMContentLoaded', updateAuthUI); 