package swole

import (
	"maps"
	"net/http"
)

type RegisteredExperiments map[string]Experiment
type ExperimentManager struct {
	registeredExperiments RegisteredExperiments
	// ExperimentStore  ExperimentStore
	PersistenceStore PersistenceStore
}

func (m *ExperimentManager) getExperiment(key string) (Experiment, bool) {
	experiment, ok := m.registeredExperiments[key]

	return experiment, ok
}

func NewExperimentManager() *ExperimentManager {
	return &ExperimentManager{
		registeredExperiments: make(RegisteredExperiments),
		// ExperimentStore:  NewMemoryExperimentStore(),
		PersistenceStore: NewCookiePersistenceStore(),
	}
}
func (m *ExperimentManager) GetRegisterExperiments() RegisteredExperiments {
	return maps.Clone(m.registeredExperiments)
}

func (m *ExperimentManager) RegisterExperiment(experiment Experiment) error {
	key := experiment.Key

	if len(key) == 0 {
		panic(&InvalidExperimentError{
			message: "the key cannot be empty",
			key:     key,
		})
	}

	_, found := m.getExperiment(key)
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

	m.registeredExperiments[key] = experiment

	return nil
}

func (m *ExperimentManager) StartExperiment(key string, w http.ResponseWriter, r *http.Request) (*StartExperimentResponse, error) {
	experiment, found := m.getExperiment(key)
	if !found {
		return nil, &ExperimentNotFoundError{
			key:     key,
			message: "StartExperiment failed, make sure you called `RegisterExperiment` first",
		}
	}

	exists, alternative, err := m.PersistenceStore.ExperimentExists(key, w, r)
	if err != nil {
		return nil, err
	}
	if !exists {
		alternative = experiment.chooseAlternative()
		err = m.PersistenceStore.PersistExperiment(key, alternative, w, r)
		if err != nil {
			return nil, err
		}
		return &StartExperimentResponse{
			Alternative:       alternative,
			DidStart:          true,
			DidStartFirstTime: true,
		}, nil
	}

	// here experiment exists
	err = m.PersistenceStore.RefreshTtl(w, r)
	if err != nil {
		return nil, err
	}

	return &StartExperimentResponse{
		Alternative:       alternative,
		DidStart:          true,
		DidStartFirstTime: false,
	}, nil
}

func (m *ExperimentManager) FinishExperiment(key string, w http.ResponseWriter, r *http.Request) (*FinishExperimentResponse, error) {
	experiment, found := m.getExperiment(key)
	if !found {
		return nil, &ExperimentNotFoundError{
			key:     key,
			message: "FinishExperiment failed, make sure you called `RegisterExperiment` first",
		}
	}
	exists, alternative, err := m.PersistenceStore.ExperimentExists(key, w, r)
	if err != nil {
		return nil, err
	}

	// experiment does not exist therefore we shouldn't finish it
	if !exists {
		return &FinishExperimentResponse{
			Alternative:        experiment.getFirstAlternative(),
			DidFinish:          false,
			DidFinishFirstTime: false,
		}, nil
	}

	finishFirstTime, err := m.PersistenceStore.ExperimentFinish(key, w, r)
	if err != nil {
		return nil, err
	}

	return &FinishExperimentResponse{
		Alternative:        alternative,
		DidFinish:          true,
		DidFinishFirstTime: finishFirstTime,
	}, nil

}
