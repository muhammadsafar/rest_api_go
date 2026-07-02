package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type user1 struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	City string `json:"city"`
}

func main_bckp() {

	port := ":3000"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// fmt.Fprintf(w,"Hello root route")
		w.Write([]byte("Hello Root Route"))
		fmt.Println("Hello Root Route")
	})

	http.HandleFunc("/teachers", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Method)

		switch r.Method {
		case http.MethodGet:

			//Acess the request details
			fmt.Println("Body :", r.Body)
			fmt.Println("Form :", r.Form)
			fmt.Println("Header:", r.Header)
			fmt.Println("Context:", r.Context())
			fmt.Println("ContextLength:", r.ContentLength)
			fmt.Println("Host:", r.Host)
			fmt.Println("Method:", r.Method)
			fmt.Println("Protocol:", r.Proto)
			fmt.Println("Remote Addr:", r.RemoteAddr)
			fmt.Println("Request URI:", r.RequestURI)
			fmt.Println("TLS:", r.TLS)
			fmt.Println("Trailer:", r.Trailer)
			fmt.Println("Transfer Encoding:", r.TransferEncoding)
			fmt.Println("URL:", r.URL)
			fmt.Println("User Agent:", r.UserAgent())
			fmt.Println("Port:", r.URL.Port())
			fmt.Println("Scheme:", r.URL.Scheme)

			// Handle GET request
			w.Write([]byte("Hello GET request for /teachers"))
			fmt.Println("Hello GET request for /teachers")
			return

		case http.MethodPost:

			//parse the form data
			err := r.ParseForm()
			if err != nil {
				http.Error(w, "Error parsing form", http.StatusBadRequest)
			}

			fmt.Println("Form : ", r.Form)

			//prepare the response data
			response := make(map[string]interface{})
			for key, values := range r.Form {
				response[key] = values[0]
			}

			fmt.Println("Processed Response Map : ", response)

			//Raw Body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				return
			}
			defer r.Body.Close()
			fmt.Println("Raw Body : ", body)
			fmt.Println("Raw Body : ", string(body))

			//if you expect json data, then unmarshal it into a struct
			var userIntance user //initialize a user instance
			err = json.Unmarshal(body, &userIntance)
			if err != nil {
				return
			}

			fmt.Println("Unmarshalled JSON into an instance of user struct", userIntance)

			//prepare the response data json
			response1 := make(map[string]interface{})
			for key, values := range r.Form {
				response[key] = values[0]
			}

			err = json.Unmarshal(body, &response1)
			if err != nil {
				return
			}

			fmt.Println("Unmarshalled JSON into a map : ", response1)

			//Acess the request details
			fmt.Println("Body :", r.Body)
			fmt.Println("Form :", r.Form)
			fmt.Println("Header:", r.Header)
			fmt.Println("Context:", r.Context())
			fmt.Println("ContextLength:", r.ContentLength)
			fmt.Println("Host:", r.Host)
			fmt.Println("Method:", r.Method)
			fmt.Println("Protocol:", r.Proto)
			fmt.Println("Remote Addr:", r.RemoteAddr)
			fmt.Println("Request URI:", r.RequestURI)
			fmt.Println("TLS:", r.TLS)
			fmt.Println("Trailer:", r.Trailer)
			fmt.Println("Transfer Encoding:", r.TransferEncoding)
			fmt.Println("URL:", r.URL)
			fmt.Println("User Agent:", r.UserAgent())
			fmt.Println("Port:", r.URL.Port())
			fmt.Println("Scheme:", r.URL.Scheme)

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

	})

	http.HandleFunc("/students", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello Students Route"))
		fmt.Println("Hello Students Route")
	})

	http.HandleFunc("/execs", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello Execs Route"))
		fmt.Println("Hello Exces Route")
	})

	fmt.Println("Starting server on port : ", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("Error starting the server:", err)
	}
}
