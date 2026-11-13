package config

import (
	"bufio"
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

// Load reads config from the environment. A local `.env` file (in the directory
// the binary runs from) is loaded first, but never overrides real env vars, so
// Render's dashboard values always win.
func Load() Config {
	loadDotEnv(".env")
	return Config{
		Port:              get("PORT", "8080"),
		Store:             get("STORE", "memory"),
		CORSOrigins:       split(get("CORS_ORIGINS", "http://localhost:5173")),
		FirebaseProj:      os.Getenv("FIREBASE_PROJECT_ID"),
		FirebaseCredsJSON: os.Getenv("FIREBASE_CREDENTIALS_JSON"),
	}
}

// loadDotEnv sets KEY=VALUE lines from path as env vars. Missing file = no-op.
// Skips blank lines, # comments, empty values, and keys already set.
// Written by hand (~20 lines) instead of adding a dependency (CLAUDE.md rule 8).
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.Trim(strings.TrimSpace(v), `"'`)
		if _, set := os.LookupEnv(k); set || v == "" {
			continue
		}
		os.Setenv(k, v)
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
