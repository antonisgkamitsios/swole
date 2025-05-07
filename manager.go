package swole

import (
	"encoding/json"
	"errors"
	"net/http"
)

const cookieName = "swole"

type ExperimentManager struct {
	activeExperiments map[string]Experiment
}

func NewExperimentManager() *ExperimentManager {
	return &ExperimentManager{
		activeExperiments: make(map[string]Experiment),
	}
}

// generateCookie creates a cookie based on a value
func (m *ExperimentManager) generateCookie(value string) http.Cookie {
	return http.Cookie{
		Name:     cookieName, // this should come from config
		Value:    value,
		Path:     "/",
		MaxAge:   60 * 60 * 24, // one day todo: this should come from config
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}

func (m *ExperimentManager) writeCookie(w http.ResponseWriter, value string) error {
	cookie := m.generateCookie(value)

	return writeCookie(w, cookie)
}

func (m *ExperimentManager) writeFreshCookie(w http.ResponseWriter, e Experiment) (*StartExperimentResponse, error) {
	alternative := e.chooseAlternative()
	value, err := e.generateCookieValue(alternative)
	if err != nil {
		return nil, err
	}

	cookie := m.generateCookie(value)

	err = writeCookie(w, cookie)
	if err != nil {
		return nil, err
	}

	return &StartExperimentResponse{
		Alternative:       alternative,
		DidStart:          true,
		DidStartFirstTime: true,
	}, nil
}

func (g *ExperimentManager) getExperiment(key string) (Experiment, error) {
	experiment, ok := g.activeExperiments[key]
	if !ok {
		return experiment, &ExperimentNotFoundError{
			message: "you should register it first via `RegisterExperiment`",
			key:     key,
		}
	}

	return experiment, nil
}

func (m *ExperimentManager) RegisterExperiment(experiment Experiment) {
	key := experiment.Key

	if len(key) == 0 {
		panic(&InvalidExperimentError{
			message: "the key cannot be empty",
			key:     key,
		})
	}

	if _, found := m.activeExperiments[key]; found {
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
		if experiment.Alternatives[i].Weight == 0 {
			experiment.Alternatives[i].Weight = 1
		}
	}

	m.activeExperiments[key] = experiment
}

func (m *ExperimentManager) StartExperiment(key string, w http.ResponseWriter, r *http.Request) (*StartExperimentResponse, error) {
	experiment, err := m.getExperiment(key)
	if err != nil {
		return nil, err
	}

	cookie, err := readCookie(r, cookieName)
	// we didn't find cookie so we should write our own
	if errors.Is(err, http.ErrNoCookie) {
		return m.writeFreshCookie(w, experiment)
	}

	if err != nil {
		return nil, err
	}

	// {"experiment_name": "control", "experiment_name:finished": "true"}
	var parsedCookieValue map[string]string

	err = json.Unmarshal([]byte(cookie.Value), &parsedCookieValue)
	if err != nil {
		return nil, err
	}

	storedExperimentAlternative, ok := parsedCookieValue[key]
	// we found the experiment stored in the cookies, we should resend the cookie to refresh it
	if ok {
		err = m.writeCookie(w, cookie.Value)
		if err != nil {
			return nil, err
		}

		return &StartExperimentResponse{
			Alternative:       storedExperimentAlternative,
			DidStart:          true,
			DidStartFirstTime: false,
		}, nil
	}

	// there is a cookie but the experiment is not stored
	// we should append to the existing values our experiment
	variant := experiment.chooseAlternative()
	parsedCookieValue[key] = variant
	newValue, err := json.Marshal(&parsedCookieValue)
	if err != nil {
		return nil, err
	}
	err = m.writeCookie(w, string(newValue))
	if err != nil {
		return nil, err
	}

	return &StartExperimentResponse{
		Alternative:       variant,
		DidStart:          true,
		DidStartFirstTime: true,
	}, nil
}

func (m *ExperimentManager) FinishExperiment(key string, w http.ResponseWriter, r *http.Request) (*FinishExperimentResponse, error) {
	experiment, err := m.getExperiment(key)
	if err != nil {
		return nil, err
	}

	cookie, err := readCookie(r, cookieName)
	// we didn't find the cookie so we shouldn't finish it
	if errors.Is(err, http.ErrNoCookie) {
		return &FinishExperimentResponse{
			Alternative:        experiment.getFirstAlternative(),
			DidFinish:          false,
			DidFinishFirstTime: false,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	// {"experiment_name": "control", "experiment_name:finished": "true"}
	var parsedCookieValue map[string]string
	err = json.Unmarshal([]byte(cookie.Value), &parsedCookieValue)
	if err != nil {
		return nil, err
	}

	storedExpAlt, ok := parsedCookieValue[key]
	// the experiment key is not present we cannot finish it
	if !ok {
		return &FinishExperimentResponse{
			Alternative:        experiment.getFirstAlternative(),
			DidFinish:          false,
			DidFinishFirstTime: false,
		}, nil
	}

	// find if the key:finished is present
	finishedKey := key + ":finished"
	_, ok = parsedCookieValue[finishedKey]
	// the experiment is already finished, we don't want to finish it again
	if ok {
		return &FinishExperimentResponse{
			Alternative:        storedExpAlt,
			DidFinish:          true,
			DidFinishFirstTime: false,
		}, nil
	}

	// We are ready to finish the experiment
	parsedCookieValue[finishedKey] = "true"
	newValue, err := json.Marshal(&parsedCookieValue)
	if err != nil {
		return nil, err
	}

	err = m.writeCookie(w, string(newValue))
	if err != nil {
		return nil, err
	}

	return &FinishExperimentResponse{
		Alternative:        storedExpAlt,
		DidFinish:          true,
		DidFinishFirstTime: true,
	}, nil

}
