package utils

import (
	"testing"
)

func TestGenerateAndValidateToken(t *testing.T) {
	SetJWTSecret("test-secret-key-12345")

	token, err := GenerateToken(1, 1, "admin", "Admin User", "admin", 8)
	if err != nil {
		t.Fatalf("Erro ao gerar token: %v", err)
	}

	if token == "" {
		t.Fatal("Token não deve ser vazio")
	}

	// Validar token
	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("Erro ao validar token: %v", err)
	}

	if claims.UserID != 1 {
		t.Errorf("UserID esperado 1, recebido %d", claims.UserID)
	}
	if claims.EmpresaID != 1 {
		t.Errorf("EmpresaID esperado 1, recebido %d", claims.EmpresaID)
	}
	if claims.Perfil != "admin" {
		t.Errorf("Perfil esperado 'admin', recebido '%s'", claims.Perfil)
	}
	if claims.Nome != "Admin User" {
		t.Errorf("Nome esperado 'Admin User', recebido '%s'", claims.Nome)
	}
	if claims.Login != "admin" {
		t.Errorf("Login esperado 'admin', recebido '%s'", claims.Login)
	}
}

func TestInvalidToken(t *testing.T) {
	SetJWTSecret("test-secret-key-12345")

	_, err := ValidateToken("token-invalido")
	if err == nil {
		t.Fatal("Deveria retornar erro para token inválido")
	}
}

func TestTokenWrongSecret(t *testing.T) {
	SetJWTSecret("secret-1")
	token, _ := GenerateToken(1, 1, "caixa", "User", "user01", 8)

	SetJWTSecret("secret-2")
	_, err := ValidateToken(token)
	if err == nil {
		t.Fatal("Deveria retornar erro quando o secret é diferente")
	}
}

func TestTokenDifferentProfiles(t *testing.T) {
	SetJWTSecret("test-secret")

	profiles := []string{"caixa", "supervisor", "gerente", "admin"}
	for _, perfil := range profiles {
		token, err := GenerateToken(1, 1, perfil, "User", "user", 8)
		if err != nil {
			t.Fatalf("Erro ao gerar token para perfil %s: %v", perfil, err)
		}

		claims, err := ValidateToken(token)
		if err != nil {
			t.Fatalf("Erro ao validar token para perfil %s: %v", perfil, err)
		}

		if claims.Perfil != perfil {
			t.Errorf("Perfil esperado '%s', recebido '%s'", perfil, claims.Perfil)
		}
	}
}
