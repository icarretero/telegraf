package kong

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/influxdata/telegraf"
)

type KongWebhook struct {
	Path string
	acc  telegraf.Accumulator
}

func (rb *KongWebhook) Register(router *mux.Router, acc telegraf.Accumulator) {
	router.HandleFunc(rb.Path, rb.eventHandler).Methods("POST")
	log.Printf("I! Started the webhooks_kong on %s\n", rb.Path)
	rb.acc = acc
}

func (rb *KongWebhook) eventHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var event RequestEvent
	err = json.Unmarshal(data, &event)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rb.acc.AddFields("kong_webhooks", event.Fields(), event.Tags(), time.Now())

	w.WriteHeader(http.StatusOK)
}