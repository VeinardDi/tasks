package date

import "time"

type Service interface {
	// Возвращает time.Time на начало текущих суток
	Today() time.Time
}
