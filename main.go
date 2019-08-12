package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

var wait time.Duration
var config string
var mapping map[string]string
var hostname string

func main() {

	// Initialise mapping
	mapping = make(map[string]string)

	// Declare variables for command flags
	var filepathVar string
	var hostVar string
	var portVar string
	var localeVar string
	var mirrorVar string
	var fullnameVar string
	var usernameVar string
	var passwordVar string
	var timezoneVar string

	// Set command flags and defaults
	flag.StringVar(&filepathVar, "filepath", "preseed.cfg", "The filepath used by dyndeb.")
	flag.StringVar(&hostVar, "host", "0.0.0.0", "The host for the web server.")
	flag.StringVar(&portVar, "port", "80", "The port for the web server.")
	flag.StringVar(&localeVar, "locale", "en_SG", "The locale of the deployment.")
	flag.StringVar(&mirrorVar, "mirror", "deb.debian.org", "The mirror used for package management.")
	flag.StringVar(&fullnameVar, "fullname", "Debian User", "The name of the user.")
	flag.StringVar(&usernameVar, "username", "debian", "The username of the user.")
	flag.StringVar(&passwordVar, "password", "r00tme", "The password of the user.")
	flag.StringVar(&timezoneVar, "timezone", "Asia/Singapore", "The timezone of the deployment.")

	// Parse args
	flag.Parse()

	// Check for non-flag arguments
	argCount := flag.NArg()
	if argCount != 0 {
		log.Fatal("Run 'dyndeb -h' for usage information")
	}

	// Read config file into variable
	log.Printf("Loading config file from %s", filepathVar)
	content, err := ioutil.ReadFile(filepathVar)
	if err != nil {
		log.Fatal(err)
	}
	config = string(content)

	// Assign
	mapping["custom-locale"] = localeVar
	mapping["custom-mirror"] = mirrorVar
	mapping["custom-fullname"] = fullnameVar
	mapping["custom-username"] = usernameVar
	mapping["custom-password"] = passwordVar
	mapping["custom-timezone"] = timezoneVar

	// Create routes
	r := mux.NewRouter().
		StrictSlash(true)
	r.HandleFunc("/", indexHandler)
	r.HandleFunc("/{hostname:[A-Za-z0-9]+}/", hostnameHandler)
	http.Handle("/", r)

	// Set host:port
	listenAddress := hostVar + ":" + portVar
	log.Printf("Listening on http://%s\n", listenAddress)

	// Run server
	srv := &http.Server{
		Addr:         listenAddress,
		WriteTimeout: time.Second * 15, // Good practice to set timeouts to avoid Slowloris attacks.
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("Shutting down")
	os.Exit(0)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "200 OK\nUsage: example.com/hostname")
}

func hostnameHandler(w http.ResponseWriter, r *http.Request) {
	// Parse variables from URL path
	vars := mux.Vars(r)

	// Append hostname to mapping
	mapping["custom-hostname"] = vars["hostname"]

	// Construct response
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, generateConfig(config, mapping))
}

func generateConfig(config string, mapping map[string]string) string {
	// Parse mapping
	for key, value := range mapping {
		config = strings.Replace(config, key, value, 1)
	}
	return config
}
