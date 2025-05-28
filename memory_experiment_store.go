package swole

type MemoryExperimentStore struct {
	activeExperiments map[string]Experiment
}

func NewMemoryExperimentStore() *MemoryExperimentStore {
	return &MemoryExperimentStore{
		activeExperiments: make(map[string]Experiment),
	}
}
func (m *MemoryExperimentStore) Get(key string) (Experiment, bool, error) {
	experiment, ok := m.activeExperiments[key]

	return experiment, ok, nil
}

func (m *MemoryExperimentStore) Set(key string, exp Experiment) error {
	m.activeExperiments[key] = exp

	return nil
}

func (m *MemoryExperimentStore) Delete(key string) error {
	delete(m.activeExperiments, key)

	return nil
}
