package config

import (
	"os"
	"strings"
)

type Config struct {
	Port         string
	Store        string // "memory" or "firestore"
	CORSOrigins  []string
	FirebaseProj string
	// Credentials: JSON content in FIREBASE_CREDENTIALS_JSON (Render), or
	// a file path in GOOGLE_APPLICATION_CREDENTIALS (local).
	FirebaseCredsJSON string
}

func Load() Config {
	return Config{
		Port:              get("PORT", "8080"),
		Store:             get("STORE", "memory"),
		CORSOrigins:       split(get("CORS_ORIGINS", "http://localhost:5173")),
		FirebaseProj:      os.Getenv("FIREBASE_PROJECT_ID"),
		FirebaseCredsJSON: os.Getenv("FIREBASE_CREDENTIALS_JSON"),
	}
}

func get(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func split(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
