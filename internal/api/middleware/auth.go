package middleware

import (
	"context"
	"net/http"
	"strings"

	"erp-backend/pkg/utils"
)

type contextKey string

const (
	UserClaimsKey contextKey = "user_claims"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.Error(w, http.StatusUnauthorized, "Token não fornecido")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.Error(w, http.StatusUnauthorized, "Formato de token inválido")
			return
		}

		claims, err := utils.ValidateToken(parts[1])
		if err != nil {
			utils.Error(w, http.StatusUnauthorized, "Token inválido ou expirado")
			return
		}

		ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserClaims(r *http.Request) *utils.JWTClaims {
	claims, ok := r.Context().Value(UserClaimsKey).(*utils.JWTClaims)
	if !ok {
		return nil
	}
	return claims
}

func RequireProfile(profiles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserClaims(r)
			if claims == nil {
				utils.Error(w, http.StatusUnauthorized, "Não autenticado")
				return
			}

			// Admin tem acesso a tudo
			if claims.Perfil == "admin" {
				next.ServeHTTP(w, r)
				return
			}

			// Hierarquia: admin > gerente > supervisor > caixa
			userLevel := profileLevel(claims.Perfil)
			authorized := false
			for _, p := range profiles {
				if userLevel >= profileLevel(p) {
					authorized = true
					break
				}
			}

			if !authorized {
				utils.Error(w, http.StatusForbidden, "Sem permissão para acessar este recurso")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func profileLevel(perfil string) int {
	switch perfil {
	case "admin":
		return 4
	case "gerente":
		return 3
	case "supervisor":
		return 2
	case "caixa":
		return 1
	default:
		return 0
	}
}
