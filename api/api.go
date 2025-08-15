package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/HappYness-Project/ChatBackendServer/internal/repository"
	"github.com/HappYness-Project/ChatBackendServer/internal/route"
	"github.com/go-chi/chi/v5"
)

type ApiServer struct {
	addr string
	db   *sql.DB
}

func NewApiServer(addr string, db *sql.DB) *ApiServer {

	return &ApiServer{
		addr: addr,
		db:   db,
	}
}

func (s *ApiServer) Setup() *chi.Mux {
	mux := chi.NewRouter()

	repo := repository.NewRepository(s.db)

	mux.Get("/", Home)
	mux.Get("/health", Home)
	msgHandler := route.NewHandler(*repo)

	mux.Group(func(r chi.Router) {
		msgHandler.RegisterRoutes(r)
	})
	go msgHandler.HandleMessages()
	return mux
}

func (s *ApiServer) Run(mux *chi.Mux) error {
	log.Println("Listening on ", s.addr)
	return http.ListenAndServe(s.addr, mux)
}
func Home(w http.ResponseWriter, r *http.Request) {
	var payload = struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Version string `json:"version"`
	}{
		Status:  "active",
		Message: "Message Service server",
		Version: "1.0.0",
	}
	route.WriteJsonWithEncode(w, http.StatusOK, payload)
}
