package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

type user struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	City string `json:"city"`
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprintf(w,"Hello root route")
	w.Write([]byte("Hello Root Route"))
	fmt.Println("Hello Root Route")
}

func teacherHandler(w http.ResponseWriter, r *http.Request) {

	//path teachers/123
	//query teachers?key1=value1&key2=value2&key3=value3

	switch r.Method {
	case http.MethodGet:

		fmt.Println(r.URL.Path)
		path := strings.TrimPrefix(r.URL.Path, "/teachers/")
		userID := strings.TrimSuffix(path, "/")
		fmt.Println("User ID:", userID)

		fmt.Println("Query params : ", r.URL.Query())

		queryParams := r.URL.Query()
		sortby := queryParams.Get("sortby")
		key := queryParams.Get("key")
		sortorder := queryParams.Get("sortorder")

		if sortorder == "" {
			sortorder = "DESC"
		}

		fmt.Printf("Sort By: %v, Key: %v, Sort Order: %v\n", sortby, key, sortorder)

		// Handle GET request
		w.Write([]byte("Hello GET request for /teachers"))
		fmt.Println("Hello GET request for /teachers")
		return

	case http.MethodPost:

		// Handle POST request
		w.Write([]byte("Hello POST request for /teachers"))
		fmt.Println("Hello POST request for /teachers")
		return

	case http.MethodPut:
		// Handle PUT request
		w.Write([]byte("Hello PUT request for /teachers"))
		fmt.Println("Hello PUT request for /teachers")
		return

	case http.MethodPatch:
		w.Write([]byte("Hello PATCH request for /teachers"))
		fmt.Println("Hello PATCH request for /teachers")
		return

	case http.MethodDelete:
		// Handle DELETE request
		w.Write([]byte("Hello DELETE request for /teachers"))
		fmt.Println("Hello DELETE request for /teachers")
		return

	default:
		// Handle other HTTP methods
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
		fmt.Println("Method not allowed")
		return
	}

}

func studentHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:

		// Handle GET request
		w.Write([]byte("Hello GET request for /students"))
		fmt.Println("Hello GET request for /students")
		return

	case http.MethodPost:

		// Handle POST request
		w.Write([]byte("Hello POST request for /students"))
		fmt.Println("Hello POST request for /students")
		return

	case http.MethodPut:
		// Handle PUT request
		w.Write([]byte("Hello PUT request for /students"))
		fmt.Println("Hello PUT request for /students")
		return

	case http.MethodPatch:
		w.Write([]byte("Hello PATCH request for /students"))
		fmt.Println("Hello PATCH request for /students")
		return

	case http.MethodDelete:
		// Handle DELETE request
		w.Write([]byte("Hello DELETE request for /students"))
		fmt.Println("Hello DELETE request for /students")
		return

	default:
		// Handle other HTTP methods
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
		fmt.Println("Method not allowed")
		return
	}
}

func execHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:

		// Handle GET request
		w.Write([]byte("Hello GET request for /execs"))
		fmt.Println("Hello GET request for /execs")
		return

	case http.MethodPost:

		// Handle POST request
		w.Write([]byte("Hello POST request for /execs"))
		fmt.Println("Hello POST request for /execs")
		return

	case http.MethodPut:
		// Handle PUT request
		w.Write([]byte("Hello PUT request for /execs"))
		fmt.Println("Hello PUT request for /execs")
		return

	case http.MethodPatch:
		w.Write([]byte("Hello PATCH request for /execs"))
		fmt.Println("Hello PATCH request for /execs")
		return

	case http.MethodDelete:
		// Handle DELETE request
		w.Write([]byte("Hello DELETE request for /execs"))
		fmt.Println("Hello DELETE request for /execs")
		return

	default:
		// Handle other HTTP methods
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
		fmt.Println("Method not allowed")
		return
	}
}

func main() {

	port := ":3000"

	http.HandleFunc("/", rootHandler)

	http.HandleFunc("/teachers/", teacherHandler)

	http.HandleFunc("/students/", studentHandler)

	http.HandleFunc("/execs/", execHandler)

	fmt.Println("Starting server on port : ", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("Error starting the server:", err)
	}
}
