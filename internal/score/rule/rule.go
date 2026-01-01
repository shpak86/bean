package rule

import (
	"bean/internal/score"
	"bean/internal/trace"

	"github.com/google/cel-go/cel"
)

// Rule represents a rule for calculating a score based on behavioral traces.
// The When field contains a CEL expression that defines the trigger condition.
// The Then field contains a Score that will be applied if the condition is true.
// The CEL program is compiled when Init is called and used during trace evaluation.
type Rule struct {
	// When — CEL expression defining the rule trigger condition.
	// Must return a boolean value.
	When string `yaml:"when"`
	// Then — score that will be added to the final result if the condition is true.
	Then score.Score `yaml:"then"`
	// program — compiled CEL program used to execute the condition.
	program cel.Program
}

// emptyScore — empty Score object returned on failed evaluation.
// Used to avoid allocations when returning nil-score.
var emptyScore = make(score.Score)

// Init compiles the string expression in the When field into an executable CEL program
// using the provided env environment.
// In case of syntax or semantic errors, returns the corresponding error.
// After successful initialization, the rule is ready for use in Eval.
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

// Eval executes the compiled rule on the provided trace t.
// The input trace is converted to map[string]any for compatibility with CEL.
// If the expression returns false or an execution error occurs, an empty Score is returned.
// If the condition is true, the value from the Then field is returned.
//
// Important: the method does not return errors in normal cases — on execution errors
// an empty Score is returned to prevent interrupting the evaluation chain.
func (r *Rule) Eval(t trace.Trace) (score.Score, error) {
	result, _, err := r.program.Eval(map[string]any(t))
	if err != nil || result.Value() == false {
		return emptyScore, nil
	}

	return r.Then, nil
}
