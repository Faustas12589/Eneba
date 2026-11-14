// Seeds Firestore from the embedded JSON files. Idempotent: documents are
// keyed by id, so re-running overwrites instead of duplicating.
//
//	FIREBASE_PROJECT_ID=... GOOGLE_APPLICATION_CREDENTIALS=./sa.json go run ./cmd/seed
package main

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"

	"eneba-marketplace/backend/internal/config"
	"eneba-marketplace/backend/internal/domain"
	"eneba-marketplace/backend/internal/repository"
	"eneba-marketplace/backend/seeds"
)

func main() {
	cfg := config.Load()
	if cfg.FirebaseProj == "" {
		log.Fatal("FIREBASE_PROJECT_ID is required")
	}
	ctx := context.Background()
	var opts []option.ClientOption
	if cfg.FirebaseCredsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(cfg.FirebaseCredsJSON)))
	}
	fs, err := firestore.NewClient(ctx, cfg.FirebaseProj, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer fs.Close()

	var products []domain.Product
	var offers []domain.Offer
	var banners []domain.Banner
	must(seeds.Load("products", &products))
	must(seeds.Load("offers", &offers))
	must(seeds.Load("banners", &banners))

	bw := fs.BulkWriter(ctx)
	// Re-seeding resets offer stock, so old cart reservations must go too,
	// or carts would "hold" units that are already back in stock.
	oldRes, err := fs.Collection(repository.ColReservations).Documents(ctx).GetAll()
	must(err)
	for _, d := range oldRes {
		_, err := bw.Delete(d.Ref)
		must(err)
	}
	for _, p := range products {
		_, err := bw.Set(fs.Collection(repository.ColProducts).Doc(p.ID), p)
		must(err)
	}
	for _, o := range offers {
		_, err := bw.Set(fs.Collection(repository.ColOffers).Doc(o.ID), o)
		must(err)
	}
	for _, b := range banners {
		_, err := bw.Set(fs.Collection(repository.ColBanners).Doc(b.ID), b)
		must(err)
	}
	bw.End()
	log.Printf("seeded %d products, %d offers, %d banners; cleared %d reservations",
		len(products), len(offers), len(banners), len(oldRes))
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
