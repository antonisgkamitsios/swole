package swole

import (
	"encoding/json"
	"math/rand"
)

type Alternatives []Alternative

func (a Alternatives) getNames() []string {
	names := make([]string, 0, len(a))

	for _, alt := range a {
		names = append(names, alt.Name)
	}

	return names

}

type Alternative struct {
	Weight int
	Name   string
}

type Experiment struct {
	Key          string
	Alternatives Alternatives
}

type StartExperimentResponse struct {
	DidStart          bool
	DidStartFirstTime bool
	Alternative       string
}

type FinishExperimentResponse struct {
	DidFinish          bool
	DidFinishFirstTime bool
	Alternative        string
}

func (e Experiment) getFirstAlternative() string {
	return e.Alternatives[0].Name
}

// chooseAlternative returns a random variant from the variants of the experiment
// based on the weights
func (e Experiment) chooseAlternative() string {
	sumWeights := 0
	for _, a := range e.Alternatives {
		sumWeights += a.Weight
	}
	point := rand.Float64() * float64(sumWeights)

	for _, a := range e.Alternatives {
		if point <= float64(a.Weight) {
			return a.Name
		}
		point -= float64(a.Weight)
	}

	// unreachable
	return ""
}

// generateCookieValue creates a string that comes from the experiment's key and alternative
func (e Experiment) generateCookieValue(alternative string) (string, error) {

	// todo calculate key based on steps
	value, err := json.Marshal(map[string]string{
		e.Key: alternative,
	})

	if err != nil {
		return "", err
	}

	return string(value), nil
}
