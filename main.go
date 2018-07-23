package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/spf13/viper"
)

type IRCMessage struct {
	Messages []string
	Channel  string
}

var messageChannel = make(chan IRCMessage, 10)

func main() {
	confDirPtr := flag.String("config", "/etc/cpthook.yml", "Path to the configfile")
	flag.Parse()

	// Load configuration from file
	confDir, confName := path.Split(*confDirPtr)
	viper.SetConfigName(strings.Split(confName, ".")[0])
	if len(confDir) > 0 {
		viper.AddConfigPath(confDir)
	} else {
		viper.AddConfigPath(".")
	}
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	var moduleList = viper.Sub("modules")

	// Status module
	if moduleList.GetBool("status.enabled") {
		log.Println("Status module is active")
		http.HandleFunc("/status", statusHandler)
	} else {
		log.Println("Status module disabled of not configured")
	}

	// Prometheus module
	if moduleList.GetBool("prometheus.enabled") {
		log.Println("Prometheus module is active")
		http.HandleFunc("/prometheus", prometheusHandler(viper.Sub("modules.prometheus")))
	}

	// Gitlab module
	if moduleList.GetBool("gitlab.enabled") {
		log.Println("Gitlab module is active")
		http.HandleFunc("/gitlab", gitlabHandler(viper.Sub("modules.gitlab")))
	}

	// Simple module
	if moduleList.GetBool("simple.enabled") {
		log.Println("Simple module is active")
		http.HandleFunc("/simple", simpleHandler(viper.Sub("modules.simple")))
	}

	// Start IRC connection
	go ircConnection(viper.Sub("irc"))

	// Start HTTP server
	log.Fatal(http.ListenAndServe(viper.GetString("http.listen"), nil))

}
