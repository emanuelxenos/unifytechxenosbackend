package handlers

import (
	"encoding/json"
	"net/http"

	"erp-backend/internal/domain/models"
	"erp-backend/internal/infrastructure/database"
	"erp-backend/internal/service"
	"erp-backend/pkg/config"
	"erp-backend/pkg/utils"
)

type AuthHandler struct {
	authService *service.AuthService
	db          *database.PostgresDB
	cfg         *config.Config
}

func NewAuthHandler(db *database.PostgresDB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: service.NewAuthService(db),
		db:          db,
		cfg:         cfg,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.UsuarioLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Dados inválidos")
		return
	}

	if req.Login == "" || req.Senha == "" {
		utils.Error(w, http.StatusBadRequest, "Login e senha são obrigatórios")
		return
	}

	resp, err := h.authService.Login(r.Context(), req, h.cfg.JWTExpiryHours)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	utils.JSONRaw(w, http.StatusOK, resp)
}

func (h *AuthHandler) Discover(w http.ResponseWriter, r *http.Request) {
	utils.JSONRaw(w, http.StatusOK, map[string]string{
		"service": "mercado-caixa",
		"version": h.cfg.AppVersion,
	})
}

func (h *AuthHandler) Health(w http.ResponseWriter, r *http.Request) {
	dbStatus := "connected"
	if err := h.db.HealthCheck(); err != nil {
		dbStatus = "disconnected"
	}

	utils.JSONRaw(w, http.StatusOK, map[string]string{
		"status":   "ok",
		"database": dbStatus,
	})
}
