package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"erp-backend/pkg/utils"
)

func TestAuthMiddleware_NoToken(t *testing.T) {
	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Esperado 401, recebido %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Esperado 401, recebido %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	utils.SetJWTSecret("test-secret")

	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-here")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Esperado 401, recebido %d", w.Code)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	utils.SetJWTSecret("test-secret")
	token, _ := utils.GenerateToken(1, 1, "admin", "Admin", "admin", 8)

	var receivedClaims *utils.JWTClaims
	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedClaims = GetUserClaims(r)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Esperado 200, recebido %d", w.Code)
	}
	if receivedClaims == nil {
		t.Fatal("Claims não deveria ser nil")
	}
	if receivedClaims.UserID != 1 {
		t.Errorf("UserID esperado 1, recebido %d", receivedClaims.UserID)
	}
	if receivedClaims.Perfil != "admin" {
		t.Errorf("Perfil esperado 'admin', recebido '%s'", receivedClaims.Perfil)
	}
}

func TestProfileLevel(t *testing.T) {
	if profileLevel("admin") != 4 {
		t.Error("admin deveria ter nível 4")
	}
	if profileLevel("gerente") != 3 {
		t.Error("gerente deveria ter nível 3")
	}
	if profileLevel("supervisor") != 2 {
		t.Error("supervisor deveria ter nível 2")
	}
	if profileLevel("caixa") != 1 {
		t.Error("caixa deveria ter nível 1")
	}
	if profileLevel("desconhecido") != 0 {
		t.Error("perfil desconhecido deveria ter nível 0")
	}
}

func TestRequireProfile_AdminAccessAll(t *testing.T) {
	utils.SetJWTSecret("test-secret")
	token, _ := utils.GenerateToken(1, 1, "admin", "Admin", "admin", 8)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := AuthMiddleware(RequireProfile("gerente")(inner))

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Admin deveria acessar rota de gerente, recebido %d", w.Code)
	}
}

func TestRequireProfile_CaixaDenied(t *testing.T) {
	utils.SetJWTSecret("test-secret")
	token, _ := utils.GenerateToken(1, 1, "caixa", "Caixa", "caixa01", 8)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := AuthMiddleware(RequireProfile("gerente")(inner))

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Caixa não deveria acessar rota de gerente, recebido %d", w.Code)
	}
}
