package dataset

import "bean/internal/trace"

type DatasetRepository interface {
	Append(token string, t trace.Trace)
	Close()
}
