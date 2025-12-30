package rule

import (
	"os"

	"github.com/google/cel-go/cel"
	"github.com/stretchr/testify/assert/yaml"
)

func LoadFromFile(file string, envProvider func() (*cel.Env, error)) ([]Rule, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	rules := []Rule{}

	err = yaml.Unmarshal(content, &rules)
	if err != nil {
		return nil, err
	}

	for i := range rules {
		env, err := envProvider()
		if err != nil {
			return nil, err
		}

		err = rules[i].Init(env)
		if err != nil {
			return nil, err
		}
	}
	return rules, nil
}
