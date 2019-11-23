/*
Copyright © 2019 quo master

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package app

import (
	"fmt"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/rs/zerolog"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)


type MonitorApp struct {
	// file-logger
	Logger zerolog.Logger
	LogFilePointer *os.File // close the file when term signal received
}

// create / factory (??) method for the monitor_app structure
func NewMonitorApp() *MonitorApp {
	pInstance := new(MonitorApp)
	err := pInstance.Init()
	if err != nil {
		panic(err)
	}
	return pInstance
}

// init the monitor_app
// 1. setup signal detector
// 2. setup logger
// 3. setup REST
func (m *MonitorApp) Init() (err error) {
	// setup signal intercept
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		// get value from signal channel and stop the channel from operating (since term signal received)
		signalValue := <-signalChannel
		signal.Stop(signalChannel)

		fmt.Println("signal received for termination:", signalValue)
		if err := m.LogFilePointer.Close(); err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		os.Exit(0)
	}()

	// setup the file logger
	m.LogFilePointer, err = os.Create("monitor.log")
	if err != nil {
		panic(err)
	}
	m.Logger = zerolog.New(m.LogFilePointer).With().Timestamp().Logger()

	// create REST api
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Put("/log/:id", m.LogMsgWithId),
		rest.Get("/logs", m.GetAllLogs),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8100", api.MakeHandler()))

	return
}

// REST endpoint implementation for put /log/:id
func (m *MonitorApp) LogMsgWithId(w rest.ResponseWriter, req *rest.Request) {
	id := req.PathParam("id")

	defer req.Body.Close()
	bContent, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	// trim, parse json
	content := strings.Trim(string(bContent), "\n")
	key := TrimQuotes(content[0:strings.Index(content, ":")])
	if strings.Compare(key, "message") == 0 {
		raw := TrimQuotes(content[strings.Index(content, ":")+1:])
		m.Logger.Info().Str("id", id).Str("raw", raw).Msg("")
	} else {
		m.Logger.Info().Str("id", id).Str("UNKNOWN_MESSAGE_TYPE", content).Msg("")
	}
}

// REST endpoint implementation for get /logs
func (m *MonitorApp) GetAllLogs(w rest.ResponseWriter, req *rest.Request) {
	bContent, err := ioutil.ReadFile(m.LogFilePointer.Name())
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bContent))
}