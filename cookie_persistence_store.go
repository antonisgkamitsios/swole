package swole

import (
	"encoding/json"
	"errors"
	"net/http"
)

const cookieName = "swole"

type CookiePersistenceStore struct {
	MaxAge int
}

func NewCookiePersistenceStore() *CookiePersistenceStore {
	return &CookiePersistenceStore{
		MaxAge: 60 * 60 * 24, // one day,
	}
}

// generateCookie creates a cookie based on a value
func (s *CookiePersistenceStore) generateCookie(value string) http.Cookie {
	return http.Cookie{
		Name:     cookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   s.MaxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}

func (s *CookiePersistenceStore) writeCookie(w http.ResponseWriter, value string) error {
	cookie := s.generateCookie(value)

	return writeCookie(w, cookie)
}

func (s *CookiePersistenceStore) ExperimentExists(key string, w http.ResponseWriter, r *http.Request) (bool, string, error) {
	cookie, err := readCookie(r, cookieName)
	// we didn't find cookie therefore experiment does not exist
	if errors.Is(err, http.ErrNoCookie) {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}

	// we need to check cookie to see if our experiment is in there
	// {"experiment_name": "control", "experiment_name:finished": "true"}
	var parsedCookieValue map[string]string
	err = json.Unmarshal([]byte(cookie.Value), &parsedCookieValue)
	if err != nil {
		return false, "", err
	}

	alternative, found := parsedCookieValue[key]
	if found {
		return true, alternative, nil
	}

	return false, "", nil
}

func (s *CookiePersistenceStore) PersistExperiment(key, alternative string, w http.ResponseWriter, r *http.Request) (err error) {
	cookieExists := true
	cookie, err := readCookie(r, cookieName)
	// we didn't find cookie therefore experiment does not exist
	if errors.Is(err, http.ErrNoCookie) {
		cookieExists = false
	} else if err != nil {
		return err
	}

	parsedCookieValue := make(map[string]string)

	if cookieExists {
		err = json.Unmarshal([]byte(cookie.Value), &parsedCookieValue)
		if err != nil {
			return err
		}
	}

	parsedCookieValue[key] = alternative
	newValue, err := json.Marshal(&parsedCookieValue)
	if err != nil {
		return err
	}
	err = s.writeCookie(w, string(newValue))
	if err != nil {
		return nil
	}

	return nil
}

func (s *CookiePersistenceStore) RefreshTtl(w http.ResponseWriter, r *http.Request) error {

	cookie, err := readCookie(r, cookieName)
	if err != nil {
		return err
	}

	err = s.writeCookie(w, cookie.Value)
	if err != nil {
		return err
	}
	return nil
}

func (s *CookiePersistenceStore) ExperimentFinish(key string, w http.ResponseWriter, r *http.Request) (finishFirstTime bool, err error) {
	cookie, err := readCookie(r, cookieName)
	if err != nil {
		return false, err
	}

	var parsedCookieValue map[string]string
	err = json.Unmarshal([]byte(cookie.Value), &parsedCookieValue)
	if err != nil {
		return false, err
	}

	finishedKey := key + ":finished"
	_, found := parsedCookieValue[finishedKey]

	parsedCookieValue[finishedKey] = "true"

	newValue, err := json.Marshal(&parsedCookieValue)
	if err != nil {
		return false, err
	}

	err = s.writeCookie(w, string(newValue))
	if err != nil {
		return false, err
	}

	// if the finished key was not found this means it is the first time we're finishing it
	return !found, nil
}
