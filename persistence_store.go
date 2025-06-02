package swole

import (
	"net/http"
)

type PersistenceStore interface {
	ExperimentExists(key string, w http.ResponseWriter, r *http.Request) (exists bool, alternative string, err error)
	PersistExperiment(key, alternative string, w http.ResponseWriter, r *http.Request) (err error)
	RefreshTtl(w http.ResponseWriter, r *http.Request) (err error)
	ExperimentFinish(key string, w http.ResponseWriter, r *http.Request) (finishFirstTime bool, err error)
}
