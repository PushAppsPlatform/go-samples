package main
import (
	"github.com/gorilla/mux"
	"net/http"
	"errors"
	"github.com/bugsnag/bugsnag-go"
	"encoding/json"
)

var semaphore chan bool
var maxRequestLength int

func main() {
	semaphore = make(chan bool, 10)
	for i := 0; i < cap(semaphore); i++ {
		semaphore <- true
	}
	bugsnag.Configure(bugsnag.Configuration{
		APIKey: "API KEY",
		ReleaseStage: "test"})
	r := mux.NewRouter()
	r.Handle("/version", recoverWrap(requestQueueHandler(http.HandlerFunc(Version)))).Methods("GET")
	http.Handle("/", r)
	err := http.ListenAndServe(":3001", nil)
	if err != nil {
		panic(err)
	}
}

func requestQueueHandler(fn http.Handler) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		<-semaphore
		if (cap(semaphore) - len(semaphore)) > maxRequestLength {
			maxRequestLength = cap(semaphore) - len(semaphore)
		}
		fn.ServeHTTP(rw, req)
	}
}

func recoverWrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var err error
		defer func() {
			semaphore <- true
			r := recover()
			if r != nil {
				switch t := r.(type) {
				case string:
					err = errors.New(t)
				case error:
					err = t
				default:
					err = errors.New("Unknown error")
				}
				bugsnag.Notify(err, req)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, req)
	})
}

func Version(writer http.ResponseWriter, req *http.Request) {
	res := VersionResponse{Version: "1"}
	js, _ := json.Marshal(res)
	writer.Header().Set("Content-Type", "application/json")
	writer.Write(js)
}

type VersionResponse struct {
	Version string `json:"version"`
}


