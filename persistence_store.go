package swole

import (
	"net/http"
)

type PersistenceStore interface {
	ExperimentExists(experiment Experiment, w http.ResponseWriter, r *http.Request) (exists bool, alternative string, err error)
	PersistExperiment(experiment Experiment, w http.ResponseWriter, r *http.Request) (alternative string, err error)
	RefreshTtl(experiment Experiment, w http.ResponseWriter, r *http.Request) (err error)
	ExperimentFinish(experiment Experiment, w http.ResponseWriter, r *http.Request) (finishFirstTime bool, err error)
}
