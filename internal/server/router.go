package server

import (
	"bean/internal/score"
	"bean/internal/trace"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

// ApiV1Router управляет маршрутами для API версии 1.
// Обрабатывает приём поведенческих трейсов, вычисление оценок и раздачу статики.
// Все эндпоинты соответствуют REST-подобной структуре.
type ApiV1Router struct {
	// tracesRepo — хранилище для сохранения и получения поведенческих трейсов по токену.
	tracesRepo *trace.TracesRepository

	// scoreCalculator — компонент для вычисления оценок на основе трейсов и правил.
	scoreCalculator *score.RulesScoreCalculator

	// static — путь к директории со статическими файлами (например, collector.js).
	// Если пусто, раздача статики отключена.
	static string

	// tokenCookie — имя cookie, используемой для идентификации сессии при отправке трейсов.
	tokenCookie string
}

// Mux возвращает настроенный *http.ServeMux с зарегистрированными обработчиками.
// Регистрирует следующие маршруты:
//   - POST /api/v1/traces — приём нового трейса
//   - GET /api/v1/scores/{token} — получение оценки по токену
//   - GET /static/... — раздача статических файлов (если включено)
func (ar *ApiV1Router) Mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/traces", ar.traceHandler)
	mux.HandleFunc("GET /api/v1/scores/{token}", ar.scoreHandler)
	if len(ar.static) != 0 {
		fs := http.FileServer(http.Dir(ar.static))
		mux.Handle("GET /static/", http.StripPrefix("/static/", fs))
	}
	return mux
}

// traceHandler обрабатывает POST-запросы с поведенческими метриками.
// Ожидает JSON-тело с данными трейса и cookie с именем, указанным в tokenCookie.
// Если данные валидны, трейс сохраняется в хранилище.
// В случае ошибки возвращает соответствующий HTTP-статус.
func (ar *ApiV1Router) traceHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Warn("Empty trace request body", "error", err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	defer r.Body.Close()

	var trace trace.Trace
	err = json.Unmarshal(body, &trace)
	if err != nil {
		slog.Warn("Unable to marshal trace request body", "error", err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	var token string
	cookies := r.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == ar.tokenCookie {
			token = cookie.Value
			break
		}
	}
	if len(token) == 0 {
		slog.Warn("Empty trace token")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	ar.tracesRepo.Append(token, trace)
	w.WriteHeader(http.StatusOK)
}

// scoreHandler обрабатывает запросы на получение оценки по токену.
// Токен извлекается из пути URL: /api/v1/scores/{token}.
// Если оценка найдена — возвращается в формате JSON.
// Если нет — возвращается статус 404.
func (ar *ApiV1Router) scoreHandler(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if len(token) == 0 {
		slog.Warn("Empty trace token")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	score, err := ar.scoreCalculator.Score(token)
	if err != nil {
		slog.Warn("Score not found", "id", token, "error", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	body, err := json.Marshal(score)
	if err != nil {
		slog.Warn("Unable to marshal score", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Write(body)
}

// NewApiV1Router создаёт новый маршрутизатор API v1.
// Принимает:
//   - static: путь к статике (может быть пустым)
//   - tokenCookie: имя cookie для идентификации сессии
//   - tracesRepo: хранилище трейсов
//   - scoreCalculator: калькулятор оценок
//
// Возвращает указатель на настроенный ApiV1Router.
func NewApiV1Router(
	static string,
	tokenCookie string,
	tracesRepo *trace.TracesRepository,
	scoreCalculator *score.RulesScoreCalculator,
) *ApiV1Router {
	return &ApiV1Router{
		tracesRepo:      tracesRepo,
		scoreCalculator: scoreCalculator,
		static:          static,
		tokenCookie:     tokenCookie,
	}
}
