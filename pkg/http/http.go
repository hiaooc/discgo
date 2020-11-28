package http

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/hiaooc/discgo/pkg/datastore"
)

type Handler struct {
	ds *datastore.DataStore
}

func NewHandler(ds *datastore.DataStore) *Handler {
	return &Handler{
		ds: ds,
	}
}

func (h *Handler) ReadConfig(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(h.ds.Contents)
	if err != nil {
		log.Printf("json marshal: %v", err)
		return
	}

	w.Header().Set("content-type", "application/json")
	if _, err := w.Write(b); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func (h *Handler) WriteConfig(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Printf("read body: %v", err)
		return
	}

	data := &datastore.Contents{}
	err = json.Unmarshal(b, data)
	if err != nil {
		log.Printf("json unmarshal: %v", err)
		return
	}

	h.ds.Contents = *data
	err = h.ds.Save()
	if err != nil {
		log.Printf("save config: %v", err)
		return
	}
}
