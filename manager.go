package swole

import (
	"net/http"
)

const cookieName = "swole"

type ExperimentManager struct {
	ExperimentStore  ExperimentStore
	PersistenceStore PersistenceStore
}

func NewExperimentManager() *ExperimentManager {
	return &ExperimentManager{
		ExperimentStore:  NewMemoryExperimentStore(),
		PersistenceStore: NewCookiePersistenceStore(cookieName),
	}
}

func (m *ExperimentManager) getExperiment(key string) (Experiment, error) {
	experiment, ok, err := m.ExperimentStore.Get(key)
	if err != nil {
		return Experiment{}, err
	}

	if !ok {
		return experiment, &ExperimentNotFoundError{
			message: "you should register it first via `RegisterExperiment`",
			key:     key,
		}
	}

	return experiment, nil
}

func (m *ExperimentManager) RegisterExperiment(experiment Experiment) error {
	key := experiment.Key

	if len(key) == 0 {
		panic(&InvalidExperimentError{
			message: "the key cannot be empty",
			key:     key,
		})
	}

	_, found, err := m.ExperimentStore.Get(key)
	if err != nil {
		return err
	}
	if found {
		panic(&InvalidExperimentError{
			message: "each experiment must be registered only once",
			key:     key,
		})
	}

	if len(experiment.Alternatives) < 2 {
		panic(&InvalidExperimentError{
			message: "should have at least 2 alternatives",
			key:     key,
		})
	}

	if !unique(experiment.Alternatives.getNames()) {
		panic(&InvalidExperimentError{
			message: "alternatives must be unique",
			key:     key,
		})
	}

	for i := range experiment.Alternatives {
		if experiment.Alternatives[i].Weight < 0 {
			panic(&InvalidExperimentError{
				message: "weights must be positive",
				key:     key,
			})
		}

		if experiment.Alternatives[i].Weight == 0 {
			experiment.Alternatives[i].Weight = 1
		}
	}

	return m.ExperimentStore.Set(key, experiment)
}

func (m *ExperimentManager) StartExperiment(key string, w http.ResponseWriter, r *http.Request) (*StartExperimentResponse, error) {
	experiment, err := m.getExperiment(key)
	if err != nil {
		return nil, err
	}

	return m.PersistenceStore.StartExperiment(experiment, w, r)
}

func (m *ExperimentManager) FinishExperiment(key string, w http.ResponseWriter, r *http.Request) (*FinishExperimentResponse, error) {
	experiment, err := m.getExperiment(key)
	if err != nil {
		return nil, err
	}

	return m.PersistenceStore.FinishExperiment(experiment, w, r)
}
