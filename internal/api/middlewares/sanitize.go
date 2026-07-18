package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"restapi/pkg/utils"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

func XSSMiddleware(next http.Handler) http.Handler {
	fmt.Println("****Initializing XSSMiddleware")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("********XSSMiddleware Ran")

		// Sanitize the URL Path
		sanitizePath, err := clean(r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Println("Original path:", r.URL.Path)
		fmt.Println("Sanitize path:", sanitizePath)

		//Sanitize Query param
		params := r.URL.Query()
		sanitizeQuery := make(map[string][]string)
		for key, values := range params {
			sanitizeKey, err := clean(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			var sanitizeValues []string
			for _, value := range values {
				cleanValue, err := clean(value)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				sanitizeValues = append(sanitizeValues, cleanValue.(string))
			}
			sanitizeQuery[sanitizeKey.(string)] = sanitizeValues
			fmt.Printf("Original Query %s:%s\n", key, strings.Join(values, ", "))
			fmt.Printf("Sanitized Query %s:%s\n", sanitizeKey, strings.Join(sanitizeValues, ", "))
		}

		r.URL.Path = sanitizePath.(string)
		r.URL.RawQuery = url.Values(sanitizeQuery).Encode()
		fmt.Println("Update URL:", r.URL.String())

		//sanitized request body
		if r.Header.Get("Content-Type") == "application/json" {
			if r.Body != nil {
				bodyByte, err := io.ReadAll(r.Body)
				if err != nil {

					http.Error(w, utils.ErrorHandler(err, "Error reading request body").Error(), http.StatusBadRequest)
					return
				}

				bodyString := strings.TrimSpace(string(bodyByte))
				fmt.Println("Original Body: ", bodyString)

				//reset the request body
				// r.Body =

				r.Body = io.NopCloser(bytes.NewReader([]byte(bodyString)))

				if len(bodyString) > 0 {
					var inputData interface{}
					err := json.NewDecoder(bytes.NewReader([]byte(bodyString))).Decode(&inputData)
					if err != nil {
						http.Error(w, utils.ErrorHandler(err, "Invalid JSON body").Error(), http.StatusBadRequest)
						return
					}
					fmt.Println("Original JSON data : ", inputData)

					//Sanitized the json body
					sanitizedData, err := clean(inputData)
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
					fmt.Println("Sanitized JSON data : ", sanitizedData)

					//Marshal the sanitized data
					sanitizedBody, err := json.Marshal(sanitizedData)
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}

					r.Body = io.NopCloser(bytes.NewReader(sanitizedBody))
					fmt.Println("Sanitized body:", string(sanitizedBody))
				} else {
					log.Println("Request Body is empty")
				}

			} else {
				log.Println("No body in the request")
			}
		} else if r.Header.Get("Content-Type") != "" {
			log.Printf("Received request with unsupported Content-Type: %s. Expected application/json\n", r.Header.Get("Content-Type"))
			http.Error(w, "Unsupported Content-Type, please use application/json", http.StatusUnsupportedMediaType)
		}

		next.ServeHTTP(w, r)
		fmt.Println("********Sending response from XSSMiddleware Ran")
	})
}

// clean sanitize input to prevent XSS attacks sd
func clean(data interface{}) (interface{}, error) {

	switch v := data.(type) {
	case map[string]interface{}:
		for k, value := range v {
			v[k] = sanitizeValue(value)
		}
		return v, nil
	case []interface{}:
		for i, value := range v {
			v[i] = sanitizeValue(value)
		}
		return v, nil
	case string:
		return sanitizeString(v), nil
	default:
		//error
		return nil, utils.ErrorHandler(fmt.Errorf("unsupported type: %T", data), fmt.Sprintf("unsupported type: %T", data))
	}

}

func sanitizeValue(data interface{}) interface{} {
	switch v := data.(type) {
	case string:
		return sanitizeString(v)

	case map[string]interface{}:
		for k, value := range v {
			v[k] = sanitizeValue(value)
		}
		return v
	case []interface{}:
		for i, value := range v {
			v[i] = sanitizeValue(value)
		}
		return v

	default:
		return v //return v as it is
	}
}

func sanitizeString(value string) string {
	return bluemonday.UGCPolicy().Sanitize(value)
}
