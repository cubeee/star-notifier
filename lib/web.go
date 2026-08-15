package lib

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Web struct {
	Notifier *StarsNotifier
}

func (w *Web) UpdateDowntimeTimestamp(writer http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		var ts int64
		if param := req.URL.Query().Get("timestamp"); param != "" {
			parsed, err := strconv.ParseInt(param, 10, 64)
			if err != nil {
				http.Error(writer, "invalid timestamp", http.StatusBadRequest)
				return
			}
			ts = parsed
		} else {
			ts = time.Now().Unix()
		}
		w.Notifier.LastDowntime = ts
		// force update and listing refresh on next loop
		w.Notifier.LastStarCheck = 0
		w.Notifier.LastListingUpdate = 0
		log.Println("Latest downtime updated to", ts)
	}
	_, err := fmt.Fprintf(writer, "%d\n", w.Notifier.LastDowntime)
	if err != nil {
		log.Println("Failed to write downtime response", err)
	}
}

func (w *Web) Start() error {
	http.HandleFunc("/downtime", w.UpdateDowntimeTimestamp)
	log.Println("Serving http endpoints on port", WebPort)
	return http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", WebPort), nil)
}
