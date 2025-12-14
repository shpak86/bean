package server

import (
	"bean/internal/score"
	"bean/internal/trace"
	"context"
	"net/http"
	"time"
)

// Server инкапсулирует HTTP-сервер приложения, предоставляя контролируемый запуск и остановку.
// Использует настраиваемый маршрутизатор и обеспечивает таймауты для безопасности и стабильности.
type Server struct {
	// server — встроенный HTTP-сервер из пакета net/http, полностью настроенный и готовый к работе.
	server *http.Server
}

// ListenAndServe запускает HTTP-сервер и начинает прослушивание указанного адреса.
// Блокирует выполнение до тех пор, пока сервер не будет остановлен или не возникнет ошибка.
// Если сервер остановлен через Shutdown, метод вернёт http.ErrServerClosed.
func (s *Server) ListenAndServe() error {
	return s.server.ListenAndServe()
}

// Shutdown корректно останавливает сервер с переданным контекстом.
// Завершает прослушивание, останавливает приём новых соединений и даёт активным соединениям
// возможность завершиться в течение таймаута, указанного в контексте.
// Должен вызываться при graceful shutdown приложения.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// NewServer создаёт и настраивает новый экземпляр сервера.
//
// Параметры:
//   - address: адрес и порт для прослушивания (например, ":8080").
//   - static: путь к директории со статическими файлами, которые будут раздаваться.
//   - tokenCookie: имя cookie, используемой для аутентификации запросов.
//   - tracesRepo: репозиторий для хранения и получения поведенческих трейсов.
//   - scoreCalculator: калькулятор, используемый для вычисления оценок на основе трейсов.
//
// Настраивает маршруты API v1, включая обработку статики и поведенческих метрик.
// Устанавливает безопасные таймауты на чтение и запись, а также ограничение на заголовки.
//
// Возвращает указатель на готовый к запуску сервер.
func NewServer(
	address string,
	static string,
	tokenCookie string,
	tracesRepo *trace.TracesRepository,
	scoreCalculator *score.RulesScoreCalculator,
) *Server {
	router := NewApiV1Router(static, tokenCookie, tracesRepo, scoreCalculator)
	s := Server{&http.Server{
		Addr:           address,
		Handler:        router.Mux(),
		ReadTimeout:    time.Second * 3,
		WriteTimeout:   time.Second * 3,
		MaxHeaderBytes: 1024 * 10,
	}}
	return &s
}
