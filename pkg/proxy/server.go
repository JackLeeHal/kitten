package proxy

import (
	"bytes"
	"context"
	"google.golang.org/grpc"
	"io"
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
	file, fileHeader, err := r.FormFile("file")
	defer file.Close()
	if err != nil || file == nil {
		wr.WriteHeader(http.StatusBadRequest)
		_, _ = wr.Write([]byte("invalid request"))

		return
	}
	ctx := context.Background()
	// directory
	directoryResp, err := s.directory.Upload(ctx, &dpb.UploadRequest{Filename: fileHeader.Filename})
	if err != nil {
		wr.WriteHeader(http.StatusInternalServerError)
		_, _ = wr.Write([]byte("directory error"))

		return
	}
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		wr.WriteHeader(http.StatusInternalServerError)
		_, _ = wr.Write([]byte("copy file error"))

		return
	}
	// store
	storeResp, err := s.store.UploadFile(ctx, &pb.UploadFileRequest{
		Vid:    directoryResp.Vid,
		Key:    directoryResp.Key,
		Cookie: directoryResp.Cookie,
		Data:   buf.Bytes(),
	})
	if err != nil {
		wr.WriteHeader(http.StatusInternalServerError)
		_, _ = wr.Write([]byte("store upload error"))

		return
	}

	wr.WriteHeader(http.StatusOK)
	_, _ = wr.Write([]byte(storeResp.Message))
}

func (s *Server) handleDelete(wr http.ResponseWriter, r *http.Request) {

}
