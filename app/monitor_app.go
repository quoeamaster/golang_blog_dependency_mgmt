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
	"strings"
)

type MonitorApp struct {
	// file-logger
	Logger zerolog.Logger
	LogFilePointer *os.File // close the file when term signal received
}

func NewMonitorApp() *MonitorApp {
	pInstance := new(MonitorApp)
	err := pInstance.Init()
	if err != nil {
		panic(err)
	}
	return pInstance
}

func (m *MonitorApp) Init() (err error) {
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

func (m *MonitorApp) GetAllLogs(w rest.ResponseWriter, req *rest.Request) {
	bContent, err := ioutil.ReadFile(m.LogFilePointer.Name())
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bContent))
}