package trace

import (
	"bean/internal/utils"
	"sync"
	"time"
)

// TracesRepository — потокобезопасное хранилище трейсов с автоматической очисткой устаревших записей.
// Для каждого идентификатора (id) поддерживается кольцевой буфер фиксированной длины.
// Трейсы, которые не обновлялись дольше указанного времени жизни (TTL), удаляются фоновым процессом.
//
// Пример использования:
//
//	repo := trace.NewTracesRepository(10, 5*time.Minute)
//	go repo.Serve()  // запуск фоновой очистки
//	repo.Append("user-123", trace.Trace{"MouseMoves": 5})
type TracesRepository struct {
	length int           // максимальное количество трейсов на один идентификатор
	ttl    time.Duration // время жизни трейса; после этого он считается устаревшим

	traces        map[string]*utils.RingBuffer[Trace] // хранилище трейсов по ID
	tracesUpdates map[string]time.Time                // время последнего обновления для каждого ID

	cleanTicker *time.Ticker // тикер для периодической очистки
	tracesMu    sync.RWMutex // мьютекс для защиты доступа к maps
}

// Append добавляет трейс t в буфер, связанный с указанным идентификатором id.
// Если для данного id ещё нет буфера, он создаётся автоматически.
// Время последнего обновления для id обновляется при создании или первом добавлении.
// Метод потокобезопасен.
func (tr *TracesRepository) Append(id string, t Trace) {
	tr.tracesMu.RLock()
	buffer, found := tr.traces[id]
	tr.tracesMu.RUnlock()

	if !found {
		tr.tracesMu.Lock()
		// Повторная проверка под блокировкой (double-checked locking)
		if buffer, found = tr.traces[id]; !found {
			buffer = utils.NewRingBuffer[Trace](tr.length)
			tr.traces[id] = buffer
			// Обновляем время последнего обновления
			tr.tracesUpdates[id] = time.Now()
		}
		tr.tracesMu.Unlock()
	}
	buffer.Push(t)
}

// Get возвращает копию всех трейсов для указанного идентификатора id в порядке от старых к новым.
// Если трейсы для данного id отсутствуют, возвращается (nil, false).
// Метод потокобезопасен.
func (tr *TracesRepository) Get(id string) ([]Trace, bool) {
	tr.tracesMu.Lock()
	defer tr.tracesMu.Unlock()
	buffer, found := tr.traces[id]
	if !found {
		return nil, false
	}
	return buffer.ToSlice(), true
}

// Serve запускает фоновую горутину, которая периодически (раз в минуту) проверяет
// и удаляет устаревшие трейсы — те, для которых с момента последнего обновления прошло больше, чем ttl.
// Метод блокирует выполнение и должен вызываться в отдельной горутине:
//
//	go repo.Serve()
//
// Для остановки используется метод Stop.
func (tr *TracesRepository) Serve() {
	tr.cleanTicker = time.NewTicker(time.Minute)
	for range tr.cleanTicker.C {
		var outdated []string

		// Собираем список устаревших ID под read-блокировкой
		tr.tracesMu.RLock()
		now := time.Now()
		for id, ts := range tr.tracesUpdates {
			if now.Sub(ts) > tr.ttl {
				outdated = append(outdated, id)
			}
		}
		tr.tracesMu.RUnlock()

		// Удаляем устаревшие записи под write-блокировкой
		if len(outdated) > 0 {
			tr.tracesMu.Lock()
			for _, id := range outdated {
				delete(tr.traces, id)
				delete(tr.tracesUpdates, id)
			}
			tr.tracesMu.Unlock()
		}
	}
}

// Stop останавливает фоновую очистку, отменяя тикер.
// Должен вызываться при завершении работы, чтобы избежать утечек ресурсов.
// Метод безопасен для вызова даже если Serve ещё не запускался.
func (tr *TracesRepository) Stop() {
	if tr.cleanTicker != nil {
		tr.cleanTicker.Stop()
	}
}

// NewTracesRepository создаёт новый экземпляр хранилища трейсов.
// Параметры:
//   - length: максимальное количество трейсов, хранимых на один идентификатор (буфер переписывается по кругу).
//   - ttl: время, после которого неактивные трейсы считаются устаревшими и удаляются фоновым процессом.
//
// Возвращает указатель на новый экземпляр TracesRepository.
// Для начала автоматической очистки необходимо вызвать Serve в отдельной горутине.
func NewTracesRepository(length int, ttl time.Duration) *TracesRepository {
	repo := TracesRepository{
		length:        length,
		ttl:           ttl,
		traces:        make(map[string]*utils.RingBuffer[Trace]),
		tracesUpdates: make(map[string]time.Time),
	}
	return &repo
}
