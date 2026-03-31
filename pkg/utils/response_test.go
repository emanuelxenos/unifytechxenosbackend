package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"nome": "teste"}
	JSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("Status esperado 200, recebido %d", w.Code)
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type esperado 'application/json', recebido '%s'", ct)
	}

	var resp APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Error("Success deveria ser true para status 200")
	}
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	Error(w, http.StatusBadRequest, "campo obrigatório")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status esperado 400, recebido %d", w.Code)
	}

	var resp APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Success {
		t.Error("Success deveria ser false para erro")
	}
	if resp.Error != "campo obrigatório" {
		t.Errorf("Mensagem de erro esperada 'campo obrigatório', recebida '%s'", resp.Error)
	}
}

func TestJSONMessage(t *testing.T) {
	w := httptest.NewRecorder()
	JSONMessage(w, http.StatusOK, "Operação realizada")

	var resp APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Message != "Operação realizada" {
		t.Errorf("Mensagem esperada 'Operação realizada', recebida '%s'", resp.Message)
	}
}

func TestJSONPaginated(t *testing.T) {
	w := httptest.NewRecorder()
	data := []string{"item1", "item2"}
	JSONPaginated(w, http.StatusOK, data, 100, 1, 50)

	var resp PaginatedResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Error("Success deveria ser true")
	}
	if resp.Total != 100 {
		t.Errorf("Total esperado 100, recebido %d", resp.Total)
	}
	if resp.Page != 1 {
		t.Errorf("Page esperado 1, recebido %d", resp.Page)
	}
	if resp.Limit != 50 {
		t.Errorf("Limit esperado 50, recebido %d", resp.Limit)
	}
}

func TestValidationError(t *testing.T) {
	w := httptest.NewRecorder()
	erros := map[string]string{
		"nome":  "campo obrigatório",
		"email": "formato inválido",
	}
	ValidationError(w, erros)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("Status esperado 422, recebido %d", w.Code)
	}
}

func TestErrorStatuses(t *testing.T) {
	cases := []struct {
		status int
		msg    string
	}{
		{http.StatusUnauthorized, "não autenticado"},
		{http.StatusForbidden, "sem permissão"},
		{http.StatusNotFound, "não encontrado"},
		{http.StatusInternalServerError, "erro interno"},
	}

	for _, c := range cases {
		w := httptest.NewRecorder()
		Error(w, c.status, c.msg)
		if w.Code != c.status {
			t.Errorf("Status esperado %d, recebido %d", c.status, w.Code)
		}
	}
}
