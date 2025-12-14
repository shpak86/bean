package score

import (
	"bean/internal/trace"

	"github.com/google/cel-go/cel"
)

// Rule представляет собой правило для вычисления оценки на основе поведенческих трейсов.
// Поле When содержит CEL-выражение, которое определяет условие срабатывания.
// Поле Then содержит оценку (Score), которая будет применена, если условие истинно.
// Программа CEL компилируется при вызове Init и используется при оценке трейсов.
type Rule struct {
	// When — CEL-выражение, определяющее условие срабатывания правила.
	// Должно возвращать логическое значение.
	When string `yaml:"when"`

	// Then — оценка, которая будет добавлена к итоговому результату, если условие истинно.
	Then Score `yaml:"then"`

	// program — скомпилированная CEL-программа, используется для выполнения условия.
	program cel.Program
}

// emptyScore — пустой объект Score, возвращаемый при неудачной оценке.
// Используется для избежания аллокаций при возврате nil-оценки.
var emptyScore = make(Score)

// Init компилирует строковое выражение в поле When в исполняемую CEL-программу
// с использованием переданного окружения env.
// В случае синтаксических или семантических ошибок возвращает соответствующую ошибку.
// После успешной инициализации правило готово к использованию в Eval.
func (r *Rule) Init(env *cel.Env) error {
	ast, iss := env.Parse(r.When)
	if iss.Err() != nil {
		return iss.Err()
	}
	checked, iss := env.Check(ast)
	if iss.Err() != nil {
		return iss.Err()
	}
	var err error
	r.program, err = env.Program(checked)
	if err != nil {
		return err
	}
	return nil
}

// Eval выполняет скомпилированное правило на переданном трейсе t.
// Входной трейс преобразуется в map[string]any для совместимости с CEL.
// Если выражение возвращает false или возникает ошибка выполнения, возвращается пустой Score.
// Если условие истинно, возвращается значение из поля Then.
//
// Важно: метод не возвращает ошибки в обычных случаях — при ошибках выполнения
// возвращается пустой Score, чтобы не прерывать цепочку оценки.
func (r *Rule) Eval(t trace.Trace) (Score, error) {
	result, _, err := r.program.Eval(map[string]any(t))
	if err != nil || result.Value() == false {
		return emptyScore, nil
	}
	return r.Then, nil
}
