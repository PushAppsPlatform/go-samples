func main() {
  r := mux.NewRouter()
  r.Handle("/version", RecoverWrap(http.HandlerFunc(Version))).Methods("GET")
  http.Handle("/", r)
  //...
}

func RecoverWrap(h http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
     var err error
     defer func() {
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
           bugsnag.Notify(err,req)
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
