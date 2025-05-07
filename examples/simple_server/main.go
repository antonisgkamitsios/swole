package main

import (
	"fmt"
	"net/http"

	"github.com/antonisgkamitsios/swole"
)

func main() {
	mux := http.NewServeMux()

	manager := swole.NewExperimentManager()
	manager.RegisterExperiment(swole.Experiment{
		Key: "test_experiment",
		Alternatives: swole.Alternatives{
			{Name: "control"},
			{Name: "variant"},
		},
	})

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		res, err := manager.StartExperiment("test_experiment", w, r)
		if err != nil {
			fmt.Printf("startExeperiment %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, `The experiment start response is: %+v`, res)
	})
	mux.HandleFunc("GET /finish", func(w http.ResponseWriter, r *http.Request) {
		res, err := manager.FinishExperiment("test_experiment", w, r)
		if err != nil {
			fmt.Printf("startExeperiment %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, `The experiment finish response is: %+v`, res)
	})

	fmt.Println("Server is running on port :3000")
	http.ListenAndServe(":3000", mux)
}
