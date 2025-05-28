package swole

import (
	"net/http"
)

type PersistenceStore interface {
	StartExperiment(experiment Experiment, w http.ResponseWriter, r *http.Request) (*StartExperimentResponse, error)
	FinishExperiment(experiment Experiment, w http.ResponseWriter, r *http.Request) (*FinishExperimentResponse, error)
}
