package swole

import "fmt"

type InvalidExperimentError struct {
	message string
	key     string
}

func (e *InvalidExperimentError) Error() string {
	return fmt.Sprintf("cannot register experiment with key: `%s`: %s", e.key, e.message)
}

type ExperimentNotFoundError struct {
	message string
	key     string
}

func (e *ExperimentNotFoundError) Error() string {
	return fmt.Sprintf("cannot retrieve experiment with key: `%s`: %s", e.key, e.message)
}
