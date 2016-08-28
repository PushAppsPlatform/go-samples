func main() {
	semaphore = make(chan bool, 10)
	for i := 0; i < cap(semaphore); i++ {
		semaphore <- true
	}
	r := mux.NewRouter()
	r.Handle("/version", RecoverWrap(requestQueueHandler(http.HandlerFunc(Version)))).Methods("GET")
}

func requestQueueHandler(fn http.Handler) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		<-semaphore
		if (cap(semaphore)- len(semaphore)) > maxRequestLength {
			maxRequestLength = cap(semaphore)- len(semaphore)
		}
		fn.ServeHTTP(rw,req)
	}
}

func RecoverWrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var err error
		defer func() {
			semaphore<- true
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
