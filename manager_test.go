package swole

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestRegisterExperiment(t *testing.T) {

	tests := []struct {
		name        string
		experiment  Experiment
		wantPanic   bool
		isDuplicate bool
	}{
		{
			name: "Empty key",
			experiment: Experiment{
				Alternatives: Alternatives{
					{
						Name: "control",
					},
					{
						Name: "variant",
					},
				},
			},
			wantPanic: true,
		},
		{
			name: "Empty alternatives",
			experiment: Experiment{
				Key: "Test key",
			},
			wantPanic: true,
		},
		{
			name: "Duplicate alternatives",
			experiment: Experiment{
				Key: "Test key",
				Alternatives: Alternatives{
					{
						Name:   "control",
						Weight: 1,
					},
					{
						Name:   "control",
						Weight: 2,
					},
				},
			},
			wantPanic: true,
		},
		{
			name: "Duplicate experiment",
			experiment: Experiment{
				Key: "Test key",
				Alternatives: Alternatives{
					{
						Name: "control",
					},
					{
						Name: "variant",
					},
				},
			},
			wantPanic:   true,
			isDuplicate: true,
		},
		{
			name: "Negative weights",
			experiment: Experiment{
				Key: "Test key",
				Alternatives: Alternatives{
					{
						Name:   "control",
						Weight: -1,
					},
					{
						Name: "variant",
					},
				},
			},
			wantPanic: true,
		},
		{
			name: "Valid experiment",
			experiment: Experiment{
				Key: "Test key",
				Alternatives: Alternatives{
					{
						Name: "control",
					},
					{
						Name: "variant",
					},
				},
			},
		},
		{name: "Valid experiment with weights",
			experiment: Experiment{
				Key: "Test key",
				Alternatives: Alternatives{
					{
						Name:   "control",
						Weight: 10,
					},
					{
						Name:   "variant",
						Weight: 20,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewExperimentManager()

			if tt.isDuplicate {
				manager.RegisterExperiment(tt.experiment)
			}

			if tt.wantPanic {
				assertPanic(t, func() {
					manager.RegisterExperiment(tt.experiment)
				})
			} else {
				manager.RegisterExperiment(tt.experiment)
				createdExperiment, err := manager.getExperiment(tt.experiment.Key)
				if err != nil {
					t.Fatal(err)
				}
				for i, alt := range createdExperiment.Alternatives {
					if tt.experiment.Alternatives[i].Name != alt.Name {
						t.Errorf("expected alternative to be %s got %s", tt.experiment.Alternatives[i].Name, alt.Name)
					}

					if alt.Weight == 0 {
						t.Errorf("expected alternative to be not 0 got %d", alt.Weight)
					}

					// weight was provided
					if tt.experiment.Alternatives[i].Weight != 0 {

						if tt.experiment.Alternatives[i].Weight != alt.Weight {
							t.Errorf("expected alternative to have weight %d got %d", tt.experiment.Alternatives[i].Weight, alt.Weight)
						}
					}
				}
			}

		})
	}
}

func TestStartExperiment(t *testing.T) {
	manager := NewExperimentManager()

	t.Run("experiment not registered", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		_, err := manager.StartExperiment("non_existent", w, r)
		if err == nil {
			t.Error("expected to error but did not")
		}
	})

	t.Run("experiment start first time", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		key := "experiment_key"
		manager.RegisterExperiment(Experiment{
			Key: key,
			Alternatives: Alternatives{
				{
					Name: "control",
				},
				{
					Name: "variant",
				},
			},
		})
		response, err := manager.StartExperiment(key, w, r)
		if err != nil {
			t.Errorf("expected not to error but got: %v", err)
		}
		if response.Alternative != "control" && response.Alternative != "variant" {
			t.Errorf("expected alternative to be control or variant but got: %s", response.Alternative)
		}

		if !response.DidStart {
			t.Error("expected DidStart to be true but got false")
		}

		if !response.DidStartFirstTime {
			t.Error("expected DidStartFirstTime to be true but got false")
		}

		// cookieValue := getExperimentCookieValue(t, w)
		// if cookieValue[key] != response.Alternative {
		// 	t.Errorf("expected alternative to be: %s but got: %s", response.Alternative, cookieValue[key])
		// }
		// if len(cookieValue) != 1 {
		// 	t.Errorf("expected to have only one experiment but got: %d", len(cookieValue))
		// }

		// manager.RegisterExperiment(Experiment{
		// 	Key: "second_experiment",
		// 	Alternatives: Alternatives{
		// 		{
		// 			Name: "control",
		// 		},
		// 		{
		// 			Name: "variant",
		// 		},
		// 	},
		// })

		// r.AddCookie(getExperimentCookie(t, w))

		// w = httptest.NewRecorder()
		// manager.StartExperiment("second_experiment", w, r)
		// cookieValue = getExperimentCookieValue(t, w)
		// fmt.Printf("cookieVal %+v\n", cookieValue)

		manager.ExperimentStore.Delete(key)
	})
	t.Run("Experiment was already started", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		key := "experiment_key"
		manager.RegisterExperiment(Experiment{
			Key: key,
			Alternatives: Alternatives{
				{
					Name: "control",
				},
				{
					Name: "variant",
				},
			},
		})

		firstResponse, err := manager.StartExperiment(key, w, r)

		fmt.Printf("Cookies are: %+v\n", w.Result().Cookies())
		if err != nil {
			t.Errorf("expected not to error but got: %v", err)
		}

		w2 := httptest.NewRecorder()
		r2 := newRequestFromResponse(w)

		response, err := manager.StartExperiment(key, w2, r2)
		if err != nil {
			t.Errorf("expected not to error but got: %v", err)
		}

		if response.Alternative != firstResponse.Alternative {
			t.Errorf("expected alternative to be %s but got: %s", firstResponse.Alternative, response.Alternative)
		}

		if !response.DidStart {
			t.Error("expected DidStart to be true but got false")
		}

		if response.DidStartFirstTime {
			t.Error("expected DidStartFirstTime to be false but got true")
		}

		manager.ExperimentStore.Delete(key)
	})

	t.Run("Experiment with different name has already started", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		key := "experiment_key"
		manager.RegisterExperiment(Experiment{
			Key: key,
			Alternatives: Alternatives{
				{
					Name: "control",
				},
				{
					Name: "variant",
				},
			},
		})

		_, err := manager.StartExperiment(key, w, r)
		if err != nil {
			t.Errorf("expected not to error but got: %v", err)
		}

		secondKey := "second_experiment"

		manager.RegisterExperiment(Experiment{
			Key: secondKey,
			Alternatives: Alternatives{
				{
					Name: "control",
				},
				{
					Name: "variant",
				},
			},
		})

		w2 := httptest.NewRecorder()
		r2 := newRequestFromResponse(w)

		response, err := manager.StartExperiment(secondKey, w2, r2)
		if err != nil {
			t.Errorf("expected not to error but got: %v", err)
		}

		if err != nil {
			t.Errorf("expected not to error but got: %v", err)
		}
		if response.Alternative != "control" && response.Alternative != "variant" {
			t.Errorf("expected alternative to be control or variant but got: %s", response.Alternative)
		}

		if !response.DidStart {
			t.Error("expected DidStart to be true but got false")
		}

		if !response.DidStartFirstTime {
			t.Error("expected DidStartFirstTime to be true but got false")
		}
	})
}

func assertPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected to panic but did not")
		}
	}()
	f()
}

func newRequestFromResponse(rr *httptest.ResponseRecorder) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	cookies := rr.Result().Cookies()
	for _, c := range cookies {
		r.AddCookie(c)
	}

	return r
}

func getExperimentCookieValue(t *testing.T, w *httptest.ResponseRecorder, cookieName string) map[string]string {
	t.Helper()
	cookie := getExperimentCookie(t, w, cookieName)

	return parseValue(t, cookie)
}

func getExperimentCookie(t *testing.T, w *httptest.ResponseRecorder, cookieName string) *http.Cookie {
	t.Helper()

	res := http.Response{Header: w.Header()}
	for _, c := range res.Cookies() {
		if c.Name == cookieName {
			return c
		}
	}
	t.Fatal("could not find the cookie")
	return nil
}

func parseValue(t *testing.T, cookie *http.Cookie) map[string]string {

	var parsedVal map[string]string

	unescapedValue, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		t.Fatal(err)
	}

	err = json.Unmarshal([]byte(unescapedValue), &parsedVal)
	if err != nil {
		t.Fatal(err)
	}

	return parsedVal
}
