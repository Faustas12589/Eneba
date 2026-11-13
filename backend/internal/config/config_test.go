package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	content := "# comment\n\nTEST_A=from_file\nTEST_B=\"quoted\"\nTEST_EMPTY=\nTEST_KEEP=from_file\nnot a pair\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TEST_KEEP", "from_env") // real env must win over the file
	for _, k := range []string{"TEST_A", "TEST_B", "TEST_EMPTY"} {
		os.Unsetenv(k)
		t.Cleanup(func() { os.Unsetenv(k) })
	}

	loadDotEnv(path)

	want := map[string]string{"TEST_A": "from_file", "TEST_B": "quoted", "TEST_KEEP": "from_env"}
	for k, v := range want {
		if got := os.Getenv(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
	if _, set := os.LookupEnv("TEST_EMPTY"); set {
		t.Error("empty values should be skipped")
	}

	loadDotEnv(filepath.Join(t.TempDir(), "missing")) // must not panic
}
