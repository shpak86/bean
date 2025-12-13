package score

import (
	"bean/internal/trace"

	"github.com/google/cel-go/cel"
)

type Rule struct {
	When    string `yaml:"when"`
	Then    Score  `yaml:"then"`
	program cel.Program
}

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

func (r *Rule) Eval(t trace.Trace) (Score, error) {
	result, _, err := r.program.Eval(t)
	if err != nil {
		return nil, err
	}
	if result.Value() == false {
		return nil, nil
	}
	return r.Then, nil
}
