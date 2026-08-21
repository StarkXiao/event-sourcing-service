package main

import (
	"context"
	"event-sourcing-service-test/internal/application"
	"event-sourcing-service-test/internal/config"
	"event-sourcing-service-test/internal/consistency"
	"event-sourcing-service-test/internal/eventstore"
	"event-sourcing-service-test/internal/platform"
	"event-sourcing-service-test/internal/projection"
	"event-sourcing-service-test/internal/replay"
	httpapi "event-sourcing-service-test/internal/transport/http"
	"log"
	"net/http"
)

func main() {
	c := config.Load()
	ctx := context.Background()
	db, e := platform.Open(ctx, c.DatabaseURL)
	if e != nil {
		log.Fatal(e)
	}
	defer db.Close()
	if c.AutoMigrate {
		if e = platform.Migrate(ctx, db); e != nil {
			log.Fatal(e)
		}
	}
	s := eventstore.Postgres{DB: db}
	w := projection.Worker{DB: db, Store: s, Orders: projection.Orders{DB: db}, Accounts: projection.Accounts{DB: db}}
	for i := 0; i < c.WorkerConcurrency; i++ {
		go w.Run(ctx)
	}
	log.Printf("listening on %s", c.HTTPAddr)
	rj := replay.Jobs{DB: db}
	ck := consistency.Checker{Store: s, DB: db}
	log.Fatal(http.ListenAndServe(c.HTTPAddr, httpapi.Router(application.Commands{Store: s}, application.Queries{Store: s}, db, application.Replay{Jobs: rj}, application.Consistency{Checker: ck}, c.AdminToken)))
}
