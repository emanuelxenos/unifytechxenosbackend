package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"erp-backend/internal/api/middleware"
	"erp-backend/internal/domain/models"
	"erp-backend/internal/infrastructure/database"
	"erp-backend/internal/service"
	"erp-backend/pkg/utils"
)

type EmpresaHandler struct {
	empresaService *service.EmpresaService
}

func NewEmpresaHandler(db *database.PostgresDB) *EmpresaHandler {
	return &EmpresaHandler{empresaService: service.NewEmpresaService(db)}
}

func (h *EmpresaHandler) Buscar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	if claims == nil {
		utils.Error(w, http.StatusUnauthorized, "Não autenticado")
		return
	}

	empresa, err := h.empresaService.Buscar(r.Context(), claims.EmpresaID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "Empresa não encontrada")
		return
	}

	utils.JSON(w, http.StatusOK, empresa)
}

func (h *EmpresaHandler) Atualizar(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	if claims == nil {
		utils.Error(w, http.StatusUnauthorized, "Não autenticado")
		return
	}

	var empresa models.Empresa
	if err := json.NewDecoder(r.Body).Decode(&empresa); err != nil {
		utils.Error(w, http.StatusBadRequest, "Dados inválidos")
		return
	}

	// Forçar o ID da empresa vindo do token para segurança
	empresa.ID = claims.EmpresaID

	if err := h.empresaService.Atualizar(r.Context(), &empresa); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.JSONMessage(w, http.StatusOK, "Configurações da empresa atualizadas com sucesso")
}

func (h *EmpresaHandler) UploadLogo(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	if claims == nil {
		utils.Error(w, http.StatusUnauthorized, "Não autenticado")
		return
	}

	// Limite de 2MB
	r.Body = http.MaxBytesReader(w, r.Body, 2*1024*1024)
	if err := r.ParseMultipartForm(2 * 1024 * 1024); err != nil {
		utils.Error(w, http.StatusBadRequest, "Arquivo muito grande (máximo 2MB)")
		return
	}

	file, header, err := r.FormFile("logo")
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Arquivo 'logo' não enviado")
		return
	}
	defer file.Close()

	// Validar extensão
	ext := filepath.Ext(header.Filename)
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		utils.Error(w, http.StatusBadRequest, "Formato inválido. Use PNG ou JPG.")
		return
	}

	// Buscar empresa para saber o logo antigo e deletar
	empresa, err := h.empresaService.Buscar(r.Context(), claims.EmpresaID)
	if err == nil && empresa.LogotipoURL != nil {
		oldPath := *empresa.LogotipoURL
		// Se for um arquivo local (começa com /uploads/)
		if strings.HasPrefix(oldPath, "/uploads/logos/") {
			fullPath := filepath.Join(".", oldPath)
			os.Remove(fullPath) // Remove o antigo sem erro se não existir
		}
	}

	// Criar nome único
	filename := fmt.Sprintf("logo_%d_%d%s", claims.EmpresaID, time.Now().Unix(), ext)
	relativePath := filepath.Join("uploads", "logos", filename)
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
	// O frontend deve concatenar com o host se necessário, mas salvar como /uploads/... é melhor
	url := "/" + filepath.ToSlash(relativePath)

	// Persistência Automática: Atualizar o banco de dados imediatamente
	if err := h.empresaService.UpdateLogoURL(r.Context(), claims.EmpresaID, url); err != nil {
		utils.Error(w, http.StatusInternalServerError, "Upload OK, mas falha ao salvar no banco")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"url": url,
	})
}
