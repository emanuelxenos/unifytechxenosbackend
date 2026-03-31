package middleware

import (
	"testing"
)

func TestIsLocalIP(t *testing.T) {
	localIPs := []string{
		"127.0.0.1",
		"192.168.1.1",
		"192.168.0.100",
		"10.0.0.1",
		"10.255.255.255",
		"172.16.0.1",
		"172.31.255.255",
	}

	for _, ip := range localIPs {
		if !isLocalIP(ip) {
			t.Errorf("IP %s deveria ser reconhecido como local", ip)
		}
	}

	externalIPs := []string{
		"8.8.8.8",
		"1.1.1.1",
		"200.100.50.25",
		"172.32.0.1",
		"11.0.0.1",
	}

	for _, ip := range externalIPs {
		if isLocalIP(ip) {
			t.Errorf("IP %s NÃO deveria ser reconhecido como local", ip)
		}
	}
}

func TestIsLocalIP_InvalidIP(t *testing.T) {
	if isLocalIP("invalid-ip") {
		t.Error("IP inválido não deveria retornar true")
	}
	if isLocalIP("") {
		t.Error("IP vazio não deveria retornar true")
	}
}
