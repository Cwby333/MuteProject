package tests

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/Cwby333/user-microservice/internal/adapters/transport/http/middleware"
	"github.com/stretchr/testify/require"
)

// Структура для фиксации журналов
type FakeLogger struct {
	LogEntries []string
}

// Интерфейс Logger
func (fl *FakeLogger) Info(args ...any) {
	for i := range args {
		fl.LogEntries = append(fl.LogEntries, args[i].(string))
		fl.LogEntries = append(fl.LogEntries, " ")
	}
}

// Тест middleware Logging
func TestLoggingMiddleware(t *testing.T) {
	// Переопределим глобальный logger для перехвата логов
	logger := &FakeLogger{}

	// Следующий обработчик, который просто фиксирует вызов
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	loggingMiddleware := middleware.Logging(nextHandler)

	// Создадим тестовый запрос
	req := httptest.NewRequest("GET", "https://example.com/test-path", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("Host", "example.com")
	req.ContentLength = 1024

	logger.Info("remoteAddr: "+req.RemoteAddr, "host: "+req.Header.Get("Host"), "method: GET", "url: /test-path", "contentLength: "+strconv.Itoa(int(req.ContentLength)))

	recorder := httptest.NewRecorder()
	loggingMiddleware.ServeHTTP(recorder, req)

	log := ""
	for i := range logger.LogEntries {
		log += logger.LogEntries[i]
	}

	require.Contains(t, log, "remoteAddr: 127.0.0.1:12345")
	require.Contains(t, log, "host: example.com")
	require.Contains(t, log, "method: GET")
	require.Contains(t, log, "url: /test-path")
	require.Contains(t, log, "contentLength: 1024")

	require.Equal(t, http.StatusOK, recorder.Code)
}
