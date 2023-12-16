package config

import (
	_ "fmt"
	_ "strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Root struct {
	App        App
	RWPostgres Postgres
	ROPostgres Postgres
}

// Constructs the root configuration by loading variables
// from the environment, plus the filenames provided.
func Load(filenames ...string) Root {
	// we do not care if there is no .env file.
	_ = godotenv.Overload(filenames...)

	r := Root{
		App:        App{},
		RWPostgres: Postgres{},
		ROPostgres: Postgres{},
	}

	mustLoad("APP", &r.App)
	mustLoad("POSTGRES_RW", &r.RWPostgres)

	return r
}

// mustLoad require env vars to satisfy spec interface rules.
func mustLoad(prefix string, spec interface{}) {
	err := envconfig.Process(prefix, spec)
	if err != nil {
		panic(err)
	}
}

// mustLoad assume env vars can to satisfy spec interface rules.
func mayLoad(prefix string, spec interface{}) {
	_ = envconfig.Process(prefix, spec)
}
