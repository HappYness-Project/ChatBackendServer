package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/HappYness-Project/ChatBackendServer/common"
	chatRepo "github.com/HappYness-Project/ChatBackendServer/internal/chat/repository"
	chatRoute "github.com/HappYness-Project/ChatBackendServer/internal/chat/route"
	messageRepo "github.com/HappYness-Project/ChatBackendServer/internal/message/repository"
	messageRoute "github.com/HappYness-Project/ChatBackendServer/internal/message/route"

	"github.com/HappYness-Project/ChatBackendServer/loggers"
	"github.com/go-chi/chi/v5"
)

type ApiServer struct {
	addr      string
	secretKey string
	db        *sql.DB
	logger    *loggers.AppLogger
}

func NewApiServer(addr string, secretKey string, db *sql.DB, logger *loggers.AppLogger) *ApiServer {

	return &ApiServer{
		addr:      addr,
		secretKey: secretKey,
		db:        db,
		logger:    logger,
	}
}

func (s *ApiServer) Setup() *chi.Mux {
	mux := chi.NewRouter()

	msgRepo := messageRepo.NewRepository(s.db)
	chatRepo := chatRepo.NewRepository(s.db)

	mux.Get("/", Home)
	mux.Get("/health", Home)
	msgHandler := messageRoute.NewHandler(s.logger, *msgRepo, *chatRepo, s.secretKey)
	chatHandler := chatRoute.NewHandler(s.logger, *chatRepo, s.secretKey)

	mux.Group(func(r chi.Router) {
		msgHandler.RegisterRoutes(r)
		chatHandler.RegisterRoutes(r)
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
	common.WriteJsonWithEncode(w, http.StatusOK, payload)
}
