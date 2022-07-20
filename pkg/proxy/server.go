package proxy

import (
	"google.golang.org/grpc"
	dpb "kitten/api/directory"
	pb "kitten/api/store"
	"log"
	"net/http"
	"time"
)

type Server struct {
	addr      string
	store     pb.StoreClient
	directory dpb.DirectoryClient
}

func NewServer(conf *Config) *Server {
	// Set up a connection to the server.
	storeConn, err := grpc.Dial(conf.StoreAddr)
	if err != nil {
		log.Fatalf("error connect to store: %v", err)
	}
	dirConn, err := grpc.Dial(conf.StoreAddr)
	if err != nil {
		log.Fatalf("error connect to directory: %v", err)
	}

	return &Server{
		addr:      conf.HttpAddr,
		store:     pb.NewStoreClient(storeConn),
		directory: dpb.NewDirectoryClient(dirConn),
	}
}

// Start the proxy server
func (s *Server) Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handle)
	server := &http.Server{
		Addr:         s.addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("failed to start proxy server: %v", err)
	}
}

// handle the http request
func (s *Server) handle(wr http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.handleGet(wr, r)

		return
	case "POST":
		s.handlePost(wr, r)

		return
	case "DELETE":
		s.handleDelete(wr, r)

		return
	default:
		http.Error(wr, "", http.StatusMethodNotAllowed)

		return
	}
}

func (s *Server) handleGet(wr http.ResponseWriter, r *http.Request) {

}

func (s *Server) handlePost(wr http.ResponseWriter, r *http.Request) {

}

func (s *Server) handleDelete(wr http.ResponseWriter, r *http.Request) {

}
