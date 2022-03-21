package rest

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"dataframe-service/src/app/logging"
)

//APIServer Receiver object for API Rest Server
type APIServer struct {
	http.Server
	shutdownReq chan bool
	wg          *sync.WaitGroup
	reqCount    uint32
}

//NewAPIServer provides an instance of APIServer
func NewAPIServer(wg *sync.WaitGroup) *APIServer {
	//create server
	s := &APIServer{
		Server: http.Server{
			Addr:         ":8080",
			ReadTimeout:  1000 * time.Second,
			WriteTimeout: 1000 * time.Second,
		},
		wg:          wg,
		shutdownReq: make(chan bool),
	}
	router := NewRouter()
	//set http server handler
	s.Handler = router
	router.HandleFunc("/shutdown", s.APIShutdownHandler)
	return s
}

//APIShutdownHandler provides shutdown handler for closing the REST Server
//It also closes the Database.
func (s *APIServer) APIShutdownHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Shutdown server"))
	//Do nothing if shutdown request already issued
	if !atomic.CompareAndSwapUint32(&s.reqCount, 0, 1) {
		log.Printf("Shutdown through API call in progress...")
		return
	}
	go func() {
		s.shutdownReq <- true
	}()
}

//RunAPIServer starts the Database and Rest Server
func (s *APIServer) RunAPIServer() {
	//Start the server
	server := s
	//Add Default Fabric from here
	alog := logging.AuditLog{Request: &logging.Request{Command: "default operations"}}
	alog.LogMessageInit(context.Background())
	alog.LogMessageReceived()
	success := true
	statusMsg := "default operations"
	log.Infoln("start server")
	alog.LogMessageEnd(&success, &statusMsg)
	done := make(chan bool)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Printf("Listen and serve: %v", err)
		}
		done <- true
	}()
	//wait shutdown
	server.waitShutdown()
	<-done
	log.Printf("DONE!")
}
func (s *APIServer) waitShutdown() {
	irqSig := make(chan os.Signal, 1)
	signal.Notify(irqSig, syscall.SIGINT, syscall.SIGTERM)
	//Wait interrupt or shutdown request through /shutdown
	select {
	case sig := <-irqSig:
		log.Printf("Shutdown request (signal: %v)", sig)
	case sig := <-s.shutdownReq:
		log.Printf("Shutdown request (/shutdown %v)", sig)
	}
	log.Printf("Stoping http server ...")
	//Create shutdown context with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	//shutdown the server
	err := s.Shutdown(ctx)
	if err != nil {
		log.Printf("Shutdown request error: %v", err)
	}
	s.wg.Done()
}