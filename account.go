package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"text/template"
)

// Login request JSON
type User struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

// Users a struct of user
type Users struct {
	Users []User `json:"users"`
}

type Username struct {
	UserName string `json:"username"`
}

type Response struct {
	Status  string `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// AuthResponse: Sent to Auth
type AuthResponse struct {
	AccessToken string `json:"accessToken"`
	UserName    string `json:"username"`
}

const error = "ERROR"
const success = "SUCCESS"

func main() {
	http.HandleFunc("/", HandleIndex)
	http.HandleFunc("/v1/account/", account)
	http.HandleFunc("/v1/session/", currentSession)

	listenPort := getenv("SHIPPED_CATALOG_LISTEN_PORT", "8888")

	log.Println("Listening on Port: " + listenPort)
	http.ListenAndServe(fmt.Sprintf(":%s", listenPort), nil)
}

// Get environment variable.  Return default if not set.
func getenv(name string, dflt string) (val string) {
	val = os.Getenv(name)
	if val == "" {
		val = dflt
	}
	return val
}

func account(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	switch req.Method {
	case "GET":
		// Logout
		file, e := ioutil.ReadFile("./session.json")
		if e != nil {
			log.Printf("File error: %v\n", e)
			os.Exit(1)
		}

		// Load JSON File
		var user User
		json.Unmarshal(file, &user)

		if user.UserName != "" {
			// Response
			rw.WriteHeader(http.StatusAccepted)
			success := response(success, http.StatusAccepted, "Logged out user "+user.UserName)
			rw.Write(success)
			log.Println("Succesfully logged out user " + user.UserName)

			// Clean up
			os.Remove("session.json")
			os.Create("session.json")
		} else {
			// No session
			rw.WriteHeader(http.StatusAccepted)
			err := response(error, http.StatusAccepted, "No user logged in")
			rw.Write(err)
			log.Println("No user logged in")
		}
	case "POST":
		// Login
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Println(err)
			return
		}
		// log.Println(string(body))

		var t User
		err = json.Unmarshal(body, &t)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(t.UserName)

		// Store Login
		loginIsSuccessful := verify(t.UserName, t.Password)

		if loginIsSuccessful {
			// Return Token Instead ** Always return JSON, Also have response Code
			rw.WriteHeader(http.StatusAccepted)
			success := AuthResponse{randToken(), t.UserName}
			response, _ := json.MarshalIndent(success, "", "    ")
			rw.Write(response)
			log.Println("Succesfully logged in user " + t.UserName)
		} else {
			rw.WriteHeader(http.StatusUnauthorized)
			error := response(error, http.StatusUnauthorized, "Login "+t.UserName+" failed.")
			rw.Write(error)
			log.Println("Login failed" + t.UserName)
		}

	case "PUT", "DELETE", "OPTIONS":
		// Update an existing record.
		rw.WriteHeader(http.StatusMethodNotAllowed)
		err := response(error, http.StatusMethodNotAllowed, req.Method+" not allowed")
		rw.Write(err)
	default:
		// Give an error message.
		rw.WriteHeader(http.StatusBadRequest)
		err := response(error, http.StatusBadRequest, "Bad request")
		rw.Write(err)
	}
}

func currentSession(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	switch req.Method {
	case "POST":
		// Get UserDB
		file, e := ioutil.ReadFile("users.json")
		if e != nil {
			fmt.Printf("File error: %v\n", e)
			os.Exit(1)
		}

		var jsontype Users
		err := json.Unmarshal(file, &jsontype)
		if err != nil {
			fmt.Println(err)
		}

		// Read User
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Println(err)
			return
		}

		var user Username
		err = json.Unmarshal(body, &user)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(user.UserName)

		for i := 0; i < len(jsontype.Users); i++ {
			if jsontype.Users[i].UserName == user.UserName {
				// Login
				rw.WriteHeader(http.StatusOK)
				success := response(success, http.StatusOK, "User found")
				rw.Write(success)
				return
			}
		}
		// No session
		rw.WriteHeader(http.StatusNoContent)
		errors := response(error, http.StatusNoContent, "No user logged in")
		rw.Write(errors)
		log.Println("No user logged in")
	case "PUT", "DELETE", "OPTIONS", "GET":
		rw.WriteHeader(http.StatusMethodNotAllowed)
		err := response(error, http.StatusMethodNotAllowed, req.Method+" not allowed")
		rw.Write(err)
	default:
		// Give an error message.
		rw.WriteHeader(http.StatusBadRequest)
		err := response(error, http.StatusBadRequest, "Bad request")
		rw.Write(err)
	}
}

func verify(username string, password string) bool {
	file, e := ioutil.ReadFile("users.json")
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	var jsontype Users
	err := json.Unmarshal(file, &jsontype)
	if err != nil {
		fmt.Println(err)
	}
	for i := 0; i < len(jsontype.Users); i++ {
		if jsontype.Users[i].UserName == username && jsontype.Users[i].Password == password {
			// Save Login
			session(jsontype.Users[i], true)
			return true
		}
	}
	return false
}

func response(status string, code int, message string) []byte {
	resp := Response{status, code, message}
	log.Println(resp.Message)
	response, _ := json.MarshalIndent(resp, "", "    ")

	return response
}

func session(user User, login bool) {
	// Clear File
	os.Remove("session.json")
	os.Create("session.json")

	file, e := os.OpenFile("session.json", os.O_RDWR|os.O_APPEND, 0660)
	defer file.Close()
	if e != nil {
		log.Printf("File error: %v\n", e)
		return
	}

	if login {
		b, err := json.Marshal(user)
		if err != nil {
			fmt.Println(err)
			return
		}
		d1 := []byte(b)
		file.Write(d1)
	}
}

func randToken() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// HandleIndex this is the index endpoint will return 200
func HandleIndex(rw http.ResponseWriter, req *http.Request) {
	lp := path.Join("templates", "layout.html")
	fp := path.Join("templates", "index.html")

	// Note that the layout file must be the first parameter in ParseFiles
	tmpl, err := template.ParseFiles(lp, fp)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(rw, nil); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	// Give a success message.
	rw.WriteHeader(http.StatusOK)
	success := response(success, http.StatusOK, "Ready for request.")
	rw.Write(success)
}
