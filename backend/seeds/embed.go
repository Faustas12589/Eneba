// Package seeds embeds the demo data so both the in-memory repository
// and the Firestore seeder can use it without file-path headaches.
package seeds

import (
	"embed"
	"encoding/json"
)

//go:embed *.json
var files embed.FS

// Load decodes seeds/<name>.json into out.
func Load(name string, out any) error {
	b, err := files.ReadFile(name + ".json")
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}
