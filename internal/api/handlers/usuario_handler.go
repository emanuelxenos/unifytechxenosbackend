package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"erp-backend/internal/api/middleware"
	"erp-backend/internal/domain/models"
	"erp-backend/internal/infrastructure/database"
	"erp-backend/internal/service"
	"erp-backend/pkg/utils"
)

type UsuarioHandler struct {
	usuarioService *service.UsuarioService
}

func NewUsuarioHandler(db *database.PostgresDB) *UsuarioHandler {
	return &UsuarioHandler{usuarioService: service.NewUsuarioService(db)}
}

func (h *UsuarioHandler) Listar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	usuarios, err := h.usuarioService.Listar(r.Context(), claims.EmpresaID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, usuarios)
}

func (h *UsuarioHandler) Criar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	var req models.CriarUsuarioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Dados inválidos")
		return
	}
	if req.Nome == "" || req.Login == "" || req.Senha == "" {
		utils.Error(w, http.StatusBadRequest, "Nome, login e senha são obrigatórios")
		return
	}

	usuario, err := h.usuarioService.Criar(r.Context(), claims.EmpresaID, req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSON(w, http.StatusCreated, usuario)
}

func (h *UsuarioHandler) Atualizar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var req models.AtualizarUsuarioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Dados inválidos")
		return
	}

	if err := h.usuarioService.Atualizar(r.Context(), claims.EmpresaID, id, req); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONMessage(w, http.StatusOK, "Usuário atualizado com sucesso")
}

func (h *UsuarioHandler) Inativar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "ID inválido")
		return
	}

	if err := h.usuarioService.Inativar(r.Context(), claims.EmpresaID, id); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONMessage(w, http.StatusOK, "Usuário excluído com sucesso")
}
