package seeds

import (
	"crypto/sha256"
	"fmt"

	"eneba-marketplace/backend/internal/domain"
)

// Keys builds the demo key pool: exactly offer.Stock available keys per offer,
// so the stock counter and the pool agree from the start (DECISIONS #19).
// Codes are derived from the key id, so re-seeding gives the same keys.
// They are fake: anyone with the source can compute them.
func Keys(offers []domain.Offer) []domain.Key {
	var keys []domain.Key
	for _, o := range offers {
		for i := 1; i <= o.Stock; i++ {
			id := fmt.Sprintf("%s-k%04d", o.ID, i)
			keys = append(keys, domain.Key{ID: id, OfferID: o.ID, Code: fakeCode(id), Status: domain.KeyAvailable})
		}
	}
	return keys
}

// fakeCode turns an id into a key-looking code like "7KQ2M-XJ4PA-9HDTE".
func fakeCode(id string) string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // no 0/O, 1/I lookalikes
	sum := sha256.Sum256([]byte(id))
	code := make([]byte, 0, 17)
	for i := range 15 {
		if i > 0 && i%5 == 0 {
			code = append(code, '-')
		}
		code = append(code, alphabet[int(sum[i])%len(alphabet)])
	}
	return string(code)
}
