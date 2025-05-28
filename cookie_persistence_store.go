package swole

import (
	"encoding/json"
	"errors"
	"net/http"
)

type CookiePersistenceStore struct {
	cookieName string
}

func NewCookiePersistenceStore(cookieName string) *CookiePersistenceStore {
	return &CookiePersistenceStore{
		cookieName: cookieName,
	}
}

// generateCookie creates a cookie based on a value
func (s *CookiePersistenceStore) generateCookie(value string) http.Cookie {
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

func (s *CookiePersistenceStore) writeCookie(w http.ResponseWriter, value string) error {
	cookie := s.generateCookie(value)

	return writeCookie(w, cookie)
}

func (s *CookiePersistenceStore) writeFreshCookie(w http.ResponseWriter, e Experiment) (*StartExperimentResponse, error) {
	alternative := e.chooseAlternative()
	value, err := e.generateCookieValue(alternative)
	if err != nil {
		return nil, err
	}

	cookie := s.generateCookie(value)

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
func (s *CookiePersistenceStore) StartExperiment(experiment Experiment, w http.ResponseWriter, r *http.Request) (*StartExperimentResponse, error) {

	key := experiment.Key
	cookie, err := readCookie(r, cookieName)
	// we didn't find cookie so we should write our own
	if errors.Is(err, http.ErrNoCookie) {
		return s.writeFreshCookie(w, experiment)
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
		err = s.writeCookie(w, cookie.Value)
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
	err = s.writeCookie(w, string(newValue))
	if err != nil {
		return nil, err
	}

	return &StartExperimentResponse{
		Alternative:       variant,
		DidStart:          true,
		DidStartFirstTime: true,
	}, nil
}
func (s *CookiePersistenceStore) FinishExperiment(experiment Experiment, w http.ResponseWriter, r *http.Request) (*FinishExperimentResponse, error) {
	key := experiment.Key

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

	err = s.writeCookie(w, string(newValue))
	if err != nil {
		return nil, err
	}

	return &FinishExperimentResponse{
		Alternative:        storedExpAlt,
		DidFinish:          true,
		DidFinishFirstTime: true,
	}, nil

}
