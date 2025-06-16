// Функция для добавления трека в избранное
async function addToFavorites(trackId) {
  if (!isLoggedIn()) {
    alert('Пожалуйста, войдите в систему, чтобы добавить трек в избранное');
    return false;
  }
  
  try {
    console.log('calling');
    // Получаем токен из localStorage
    const accessToken = localStorage.getItem('jwt-access-token');
    if (!accessToken) {
      console.error('No access token found');
      alert('Необходимо войти в систему для добавления трека в избранное');
      return false;
    }
    
    const response = await fetch(`http://localhost/user/track/favorite?track_id=${trackId}`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    });
    
    if (!response.ok) {
      if (response.status === 401) {
        alert('Необходимо войти в систему для добавления трека в избранное');
        return false;
      }
      throw new Error(`HTTP error! Status: ${response.status}`);
    }
    
    // Получаем данные ответа
    const data = await response.text();
    console.log('Track added to favorites:', data);
    
    // Добавляем трек в локальное хранилище для быстрого отображения статуса
    addToLocalFavorites(trackId);
    return true;
  } catch (error) {
    console.error('Error adding track to favorites:', error);
    alert('Не удалось добавить трек в избранное. Проверьте консоль для деталей.');
    return false;
  }
}

// Функция для удаления трека из избранного
async function removeFromFavorites(trackId) {
  if (!isLoggedIn()) {
    alert('Пожалуйста, войдите в систему');
    return false;
  }
  
  try {
    // Получаем токен из localStorage
    const accessToken = localStorage.getItem('jwt-access-token');
    if (!accessToken) {
      console.error('No access token found');
      alert('Необходимо войти в систему');
      return false;
    }
    
    const response = await fetch(`http://localhost/user/track/favorite?track_id=${trackId}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
      },
    });
    
    if (!response.ok) {
      if (response.status === 401) {
        alert('Необходимо войти в систему');
        return false;
      }
      throw new Error(`HTTP error! Status: ${response.status}`);
    }
    
    // Получаем данные ответа
    const data = await response.text();
    console.log('Track removed from favorites:', data);
    
    // Удаляем трек из локального хранилища
    removeFromLocalFavorites(trackId);
    return true;
  } catch (error) {
    console.error('Error removing track from favorites:', error);
    alert('Не удалось удалить трек из избранного. Проверьте консоль для деталей.');
    return false;
  }
}

// Функция для проверки, находится ли трек в избранном
function isTrackFavorite(trackId) {
  const favoritesData = localStorage.getItem('favoriteTracks');
  if (!favoritesData) return false;
  
  try {
    const favorites = JSON.parse(favoritesData);
    return favorites.includes(trackId);
  } catch (e) {
    return false;
  }
}

// Добавление трека в локальное хранилище избранных
function addToLocalFavorites(trackId) {
  let favorites = [];
  const favoritesData = localStorage.getItem('favoriteTracks');
  
  if (favoritesData) {
    try {
      favorites = JSON.parse(favoritesData);
    } catch (e) {
      favorites = [];
    }
  }
  
  if (!favorites.includes(trackId)) {
    favorites.push(trackId);
    localStorage.setItem('favoriteTracks', JSON.stringify(favorites));
  }
}

// Удаление трека из локального хранилища избранных
function removeFromLocalFavorites(trackId) {
  const favoritesData = localStorage.getItem('favoriteTracks');
  if (!favoritesData) return;
  
  try {
    let favorites = JSON.parse(favoritesData);
    favorites = favorites.filter(id => id !== trackId);
    localStorage.setItem('favoriteTracks', JSON.stringify(favorites));
  } catch (e) {
    console.error('Error removing from local favorites:', e);
  }
}

// Функция для обновления отображения кнопки избранного
function updateFavoriteButton(button, trackId) {
  const isFavorite = isTrackFavorite(trackId);
  
  if (isFavorite) {
    button.innerHTML = '❤️';
    button.classList.add('favorited');
    button.setAttribute('title', 'Удалить из избранного');
    button.onclick = () => {
      removeFromFavorites(trackId).then(success => {
        if (success) {
          updateFavoriteButton(button, trackId);
        }
      });
    };
  } else {
    button.innerHTML = '🤍';
    button.classList.remove('favorited');
    button.setAttribute('title', 'Добавить в избранное');
    button.onclick = () => {
      addToFavorites(trackId).then(success => {
        if (success) {
          updateFavoriteButton(button, trackId);
        }
      });
    };
  }
}

// Загрузка избранных треков с сервера
async function loadFavoriteTracks() {
  if (!isLoggedIn()) {
    console.log('not logged in');
    return;
  };
  
  const userData = getUserData();
  console.log(`user data: ${userData}`);
  if (!userData || !userData.id) {
    console.log('user data is not valid');
    return;
  };

  console.log(`user data: ${userData}`);
  
  try {
    const response = await fetch(`http://localhost/tracks/${userData.id}`, {
      method: 'GET',
    });
    
    if (!response.ok) {
      throw new Error(`HTTP error! Status: ${response.status}`);
    }
    
    const tracks = await response.json();
    
    // Сохраняем ID избранных треков в localStorage
    const trackIds = tracks.map(track => track.id);
    localStorage.setItem('favoriteTracks', JSON.stringify(trackIds));
    
    return tracks;
  } catch (error) {
    console.error('Error loading favorite tracks:', error);
    return [];
  }
} 