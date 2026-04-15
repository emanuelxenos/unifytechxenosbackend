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

type ProdutoHandler struct {
	produtoService *service.ProdutoService
}

func NewProdutoHandler(db *database.PostgresDB) *ProdutoHandler {
	return &ProdutoHandler{produtoService: service.NewProdutoService(db)}
}

func (h *ProdutoHandler) Listar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	search := r.URL.Query().Get("search")
	catStr := r.URL.Query().Get("categoria_id")
	var catID *int
	if catStr != "" {
		id, _ := strconv.Atoi(catStr)
		catID = &id
	}

	produtos, total, err := h.produtoService.Listar(r.Context(), claims.EmpresaID, page, limit, catID, search)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}
	utils.JSONPaginated(w, http.StatusOK, produtos, total, page, limit)
}

func (h *ProdutoHandler) Buscar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	codigo := r.URL.Query().Get("codigo")
	nome := r.URL.Query().Get("nome")
	var codigoPtr, nomePtr *string
	if codigo != "" {
		codigoPtr = &codigo
	}
	if nome != "" {
		nomePtr = &nome
	}

	produto, err := h.produtoService.Buscar(r.Context(), claims.EmpresaID, codigoPtr, nomePtr)
	if err != nil {
		utils.Error(w, http.StatusNotFound, err.Error())
		return
	}
	utils.JSONRaw(w, http.StatusOK, produto)
}

func (h *ProdutoHandler) BuscarPorID(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "ID inválido")
		return
	}

	produto, err := h.produtoService.BuscarPorID(r.Context(), claims.EmpresaID, id)
	if err != nil {
		utils.Error(w, http.StatusNotFound, err.Error())
		return
	}
	utils.JSON(w, http.StatusOK, produto)
}

func (h *ProdutoHandler) Criar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	var req models.CriarProdutoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Dados inválidos")
		return
	}
	if req.Nome == "" || req.PrecoVenda <= 0 {
		utils.Error(w, http.StatusBadRequest, "Nome e preço de venda são obrigatórios")
		return
	}

	produto, err := h.produtoService.Criar(r.Context(), claims.EmpresaID, req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSON(w, http.StatusCreated, produto)
}

func (h *ProdutoHandler) Atualizar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var req models.CriarProdutoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Dados inválidos")
		return
	}

	if err := h.produtoService.Atualizar(r.Context(), claims.EmpresaID, id, req); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSONMessage(w, http.StatusOK, "Produto atualizado com sucesso")
}

func (h *ProdutoHandler) Inativar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "ID inválido")
		return
	}

	if err := h.produtoService.Inativar(r.Context(), claims.EmpresaID, id); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.JSONMessage(w, http.StatusOK, "Produto inativado com sucesso")
}
