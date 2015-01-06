package web

import (
	"../../config"
	"../client"
	"../databases"
	"../events"
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type errorPage struct {
	Code  int
	Error string
}

type PadeDescription struct {
	Title          string
	StaticTemplate string
	Template       string
	Data           interface{}
}

const (
	NOT_FOUND      = 404
	INTERNAL_ERROR = 500
)

var templates *template.Template

func Start(conf config.HTTPConfig) {
	events.Init()
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, conf.TemplateDir+"/"+r.URL.Path)
	})
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/databases", handleDatabases)
	http.HandleFunc("/databases/settings", handleDBSettings)
	http.HandleFunc("/ws", handleWs)
	var err error
	templates, err = template.ParseGlob(conf.TemplateDir + "/html/*.html")
	if err != nil {
		log.Println("Parse templates failed: " + err.Error())
	}
	log.Panicln(http.ListenAndServe(conf.Host+":"+strconv.Itoa(conf.Port), nil))
}

// Handle Pages
func handleWs(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		return
	}
	client.NewClient(ws, r.RemoteAddr, r.UserAgent())
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "IndexPage", &PadeDescription{Title: "General page", StaticTemplate: "IndexStatic", Template: "IndexPage", Data: r.UserAgent() + " " + r.Host})
	if err != nil {
		log.Println("Error send error's page: " + err.Error())
	}
}

func handleDatabases(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "DatabasesPage", &PadeDescription{Title: "General page", StaticTemplate: "IndexStatic", Template: "IndexPage", Data: r.UserAgent() + " " + r.Host})
	if err != nil {
		log.Println("Error send error's page: " + err.Error())
	}
}

func handleDBSettings(w http.ResponseWriter, r *http.Request) {
	type DBSetDataOnLoad struct {
		User      string
		StepsList []databases.MySQLReplicaStep
	}
	pageData := &DBSetDataOnLoad{User: r.UserAgent() + " " + r.Host, StepsList: databases.GetReplicaStepsForChoose()}
	err := templates.ExecuteTemplate(w, "DatabasesSettingsPage", &PadeDescription{Title: "Database Settings", StaticTemplate: "IndexStatic", Template: "IndexPage", Data: pageData})
	if err != nil {
		log.Println("Error send error's page: " + err.Error())
	}
}
