package swole

type ExperimentStore interface {
	Get(key string) (Experiment, bool, error)
	Set(key string, exp Experiment) error
	Delete(key string) error
}
