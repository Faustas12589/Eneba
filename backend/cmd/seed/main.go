// Seeds Firestore from the embedded JSON files plus a generated key pool.
// Idempotent: documents are keyed by id, so re-running overwrites instead of
// duplicating. It is a full demo reset: carts and orders are deleted too.
//
// Cost: about 2 x 5600 writes/deletes per run. The free (Spark) plan allows
// 20k writes and 20k deletes per day, so only seed a few times a day.
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
	keys := seeds.Keys(offers)

	// Step 1: delete. Re-seeding resets offer stock and the key pool, so carts
	// and orders that point at old keys must go too. Old "reservations" docs
	// are from before carts became orders.
	del := newBatch(fs, ctx)
	for _, col := range []string{repository.ColKeys, repository.ColOrders, repository.ColReservations} {
		docs, err := fs.Collection(col).Documents(ctx).GetAll()
		must(err)
		for _, d := range docs {
			del.delete(d.Ref)
		}
	}
	deleted := del.finish("delete") // waits, so no delete can land after a write below

	// Step 2: write.
	bw := newBatch(fs, ctx)
	for _, p := range products {
		bw.set(fs.Collection(repository.ColProducts).Doc(p.ID), p)
	}
	for _, o := range offers {
		bw.set(fs.Collection(repository.ColOffers).Doc(o.ID), o)
	}
	for _, b := range banners {
		bw.set(fs.Collection(repository.ColBanners).Doc(b.ID), b)
	}
	for _, k := range keys {
		bw.set(fs.Collection(repository.ColKeys).Doc(k.ID), k)
	}
	bw.finish("write")
	log.Printf("seeded %d products, %d offers, %d banners, %d keys; deleted %d old keys/orders/reservations",
		len(products), len(offers), len(banners), len(keys), deleted)
}

// batch wraps a BulkWriter and keeps every job, because a BulkWriter does not
// report failed writes by itself: each job's result has to be checked.
// (Before this, a seed that hit the daily write quota still printed success.)
type batch struct {
	bw   *firestore.BulkWriter
	jobs []*firestore.BulkWriterJob
}

func newBatch(fs *firestore.Client, ctx context.Context) *batch {
	return &batch{bw: fs.BulkWriter(ctx)}
}

func (b *batch) set(ref *firestore.DocumentRef, v any) {
	j, err := b.bw.Set(ref, v)
	must(err)
	b.jobs = append(b.jobs, j)
}

func (b *batch) delete(ref *firestore.DocumentRef) {
	j, err := b.bw.Delete(ref)
	must(err)
	b.jobs = append(b.jobs, j)
}

// finish flushes, then stops the program if any write failed. Returns the count.
func (b *batch) finish(what string) int {
	b.bw.End()
	failed := 0
	var first error
	for _, j := range b.jobs {
		if _, err := j.Results(); err != nil {
			failed++
			if first == nil {
				first = err
			}
		}
	}
	if failed > 0 {
		log.Fatalf("%d of %d %ss failed (first error: %v). Data is half-seeded: fix the cause "+
			"(often the free plan's daily write quota) and run the seed again.", failed, len(b.jobs), what, first)
	}
	return len(b.jobs)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
