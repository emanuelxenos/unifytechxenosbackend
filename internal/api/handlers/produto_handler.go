package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

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
	baixoEstoque, _ := strconv.ParseBool(r.URL.Query().Get("baixo_estoque"))
	vencendo, _ := strconv.ParseBool(r.URL.Query().Get("vencendo"))
	var catID *int
	if catStr != "" {
		id, _ := strconv.Atoi(catStr)
		catID = &id
	}

	produtos, total, err := h.produtoService.Listar(r.Context(), claims.EmpresaID, page, limit, catID, search, baixoEstoque, vencendo)
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
		utils.Error(w, http.StatusBadRequest, "Dados inválidos: "+err.Error())
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
		utils.Error(w, http.StatusBadRequest, "Dados inválidos: "+err.Error())
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

func (h *ProdutoHandler) UploadFoto(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	if claims == nil {
		utils.Error(w, http.StatusUnauthorized, "Não autenticado")
		return
	}

	// Limite de 5MB para fotos de produtos
	r.Body = http.MaxBytesReader(w, r.Body, 5*1024*1024)
	if err := r.ParseMultipartForm(5 * 1024 * 1024); err != nil {
		utils.Error(w, http.StatusBadRequest, "Arquivo muito grande (máximo 5MB)")
		return
	}

	file, header, err := r.FormFile("foto")
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Arquivo 'foto' não enviado")
		return
	}
	defer file.Close()

	// Validar extensão
	ext := filepath.Ext(header.Filename)
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".webp" {
		utils.Error(w, http.StatusBadRequest, "Formato inválido. Use PNG, JPG ou WEBP.")
		return
	}

	// Criar nome único
	filename := fmt.Sprintf("prod_%d_%d%s", claims.EmpresaID, time.Now().UnixNano(), ext)
	relativePath := filepath.Join("uploads", "produtos", filename)
	fullPath := filepath.Join(".", relativePath)

	out, err := os.Create(fullPath)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Erro ao criar arquivo no servidor")
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		utils.Error(w, http.StatusInternalServerError, "Erro ao salvar arquivo")
		return
	}

	// Retornar a URL relativa
	url := "/" + filepath.ToSlash(relativePath)

	utils.JSON(w, http.StatusOK, map[string]string{
		"url": url,
	})
}

func (h *ProdutoHandler) AtualizarPrecosLote(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	if claims == nil {
		utils.Error(w, http.StatusUnauthorized, "Não autenticado")
		return
	}

	var req models.BatchPrecoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Dados inválidos")
		return
	}

	if len(req.Updates) == 0 {
		utils.Error(w, http.StatusBadRequest, "Nenhuma atualização fornecida")
		return
	}

	err := h.produtoService.AtualizarPrecosLote(r.Context(), claims.EmpresaID, req.Updates)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONMessage(w, http.StatusOK, "Preços atualizados com sucesso")
}
