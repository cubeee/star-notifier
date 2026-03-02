package lib

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type Web struct {
	Notifier *StarsNotifier
}

func (w *Web) UpdateDowntimeTimestamp(writer http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		now := time.Now().Unix()
		w.Notifier.LastDowntime = &now
		log.Println("Latest downtime updated to", now)
	}
	_, err := fmt.Fprintf(writer, "%d\n", *w.Notifier.LastDowntime)
	if err != nil {
		log.Println("Failed to write downtime response", err)
	}
}

func (w *Web) Start() error {
	http.HandleFunc("/downtime", w.UpdateDowntimeTimestamp)
	log.Println("Serving http endpoints on port", WebPort)
	return http.ListenAndServe(fmt.Sprintf(":%d", WebPort), nil)
}
