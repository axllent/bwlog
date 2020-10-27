package app

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// BasicAuthFromFile opens a file and sets a global user/password if a line with two words is found
func BasicAuthFromFile(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		l := scanner.Text()
		if l != "" {
			words := strings.Fields(l)
			if len(words) == 2 {
				Config.AuthUser = words[0]
				Config.AuthPass = words[1]
			}
		}

		if Config.AuthUser == "" || Config.AuthPass == "" {
			return fmt.Errorf("Basic auth user or password not found in %s\nThe format should be <user> <password>", filePath)
		}
	}
	err = scanner.Err()
	return err
}

// basicAuthWrapper uses login details
func basicAuthWrapper(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if Config.AuthUser != "" && Config.AuthPass != "" {
			user, pass, ok := r.BasicAuth()

			if !ok {
				basicAuthResponse(w)
				return
			}

			if user != Config.AuthUser || pass != Config.AuthPass {
				basicAuthResponse(w)
				return
			}
		}
		// pass on request
		handler(w, r)
	}
}

// BasicAuthResponse returns an basic auth response to the browser
func basicAuthResponse(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Login"`)
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("Unauthorised.\n"))
}
