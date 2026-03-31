package utils

import (
	"testing"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "123456"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Erro ao gerar hash: %v", err)
	}

	if hash == "" {
		t.Fatal("Hash não deve ser vazio")
	}

	if hash == password {
		t.Fatal("Hash não deve ser igual à senha original")
	}

	// Verificar senha correta
	if !CheckPassword(hash, password) {
		t.Fatal("CheckPassword deveria retornar true para senha correta")
	}

	// Verificar senha incorreta
	if CheckPassword(hash, "senha_errada") {
		t.Fatal("CheckPassword deveria retornar false para senha incorreta")
	}
}

func TestHashPasswordDifferentHashes(t *testing.T) {
	password := "mesma_senha"

	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)

	if hash1 == hash2 {
		t.Fatal("Hashes devem ser diferentes (bcrypt usa salt aleatório)")
	}

	// Ambos devem validar com a mesma senha
	if !CheckPassword(hash1, password) {
		t.Fatal("hash1 deveria validar")
	}
	if !CheckPassword(hash2, password) {
		t.Fatal("hash2 deveria validar")
	}
}
