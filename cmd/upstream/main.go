package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

var page = template.Must(template.New("upstream").Parse(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Protected Resource Demo</title>
  </head>
  <body>
    <main>
      <h1>Protected Demo</h1>
      <hr />
      <h2>User Identity</h2>
      <p><strong>X-Auth-Request-User:</strong> {{ .AuthRequestUser }}</p>
      <p><strong>Remote-User:</strong> {{ .RemoteUser }}</p>
      <p><strong>X-Forwarded-User:</strong> {{ .ForwardedUser }}</p>
      <hr />
      <h2>Profile</h2>
      <p><strong>X-Auth-Request-Email:</strong> {{ .Email }}</p>
      <p><strong>X-Auth-Request-Preferred-Username:</strong> {{ .PreferredUsername }}</p>
      <p><strong>X-Auth-Request-Name:</strong> {{ .Name }}</p>
      <p><strong>X-Auth-Request-Picture:</strong> {{ .Picture }}</p>
      <hr />
      <h2>Request Info</h2>
      <p><strong>Host:</strong> {{ .Host }}</p>
      <p><strong>Path:</strong> {{ .Path }}</p>
    </main>
  </body>
</html>`))

type pageData struct {
	AuthRequestUser   string
	RemoteUser        string
	ForwardedUser     string
	Email             string
	PreferredUsername string
	Name              string
	Picture           string
	Host              string
	Path              string
}

func main() {
	port := os.Getenv("UPSTREAM_PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := page.Execute(w, pageData{
			AuthRequestUser:   r.Header.Get("X-Auth-Request-User"),
			RemoteUser:        r.Header.Get("Remote-User"),
			ForwardedUser:     r.Header.Get("X-Forwarded-User"),
			Email:             r.Header.Get("X-Auth-Request-Email"),
			PreferredUsername: r.Header.Get("X-Auth-Request-Preferred-Username"),
			Name:              r.Header.Get("X-Auth-Request-Name"),
			Picture:           r.Header.Get("X-Auth-Request-Picture"),
			Host:              r.Host,
			Path:              r.URL.Path,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	address := fmt.Sprintf("127.0.0.1:%s", port)
	log.Printf("Upstream demo listening on http://%s", address)
	log.Fatal(http.ListenAndServe(address, nil))
}
