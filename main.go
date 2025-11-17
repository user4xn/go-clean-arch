// @title Clean Arch API
// @version 1.0
// @description API for authentication & user service

// @contact.name API Support
// @contact.email support@example.com

// @host localhost:8080
// @BasePath /api/v1
package main

import (
	"clean-arch/database"
	"clean-arch/database/migration"
	"clean-arch/database/seeder"
	"clean-arch/internal/factory"
	"clean-arch/internal/http"
	"clean-arch/pkg/config"
	"clean-arch/pkg/genx"
	"flag"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func main() {
	config.LoadEnv(".env")

	var (
		m   string
		s   string
		i   bool
		mmf string
		gen string
	)

	database.CreateConnection()

	flag.StringVar(
		&m,
		"m",
		"",
		`This flag is used for migration`,
	)

	flag.StringVar(
		&s,
		"s",
		"none",
		`This flag is used for seeder`,
	)

	flag.BoolVar(
		&i,
		"i",
		false,
		`This flag is used for first migration`,
	)

	flag.StringVar(
		&mmf,
		"mmf",
		"",
		`This flag is used for creating migration file name`,
	)

	flag.StringVar(
		&gen,
		"gen",
		"",
		`This flag is used for generating app file`,
	)

	flag.Parse()

	if i {
		migration.FirstMigrate()
		return
	}

	if m != "" {
		if m == "all" {
			err := migration.MigrateAll()
			if err != nil {
				fmt.Println(cases.Title(language.Indonesian).String(err.Error()))
			}
			return
		}

		err := migration.Migrate(m)
		if err != nil {
			fmt.Println(cases.Title(language.Indonesian).String(err.Error()))
		}
		return
	}

	if s == "seed" {
		seeder.Seed()
		return
	}

	if mmf != "" {
		err := migration.CreateMigrationFile(mmf)
		if err != nil {
			fmt.Println(cases.Title(language.Indonesian).String(err.Error()))
		}
		return
	}

	if gen != "" {
		if gen == "all" {
			tmplData := genx.GetData()
			err := genx.GenerateAll(tmplData)
			if err != nil {
				log.Fatalf("failed to generate all: %v", err)
			}
		}
		return
	}

	f := factory.NewFactory() // Database instance initialization
	g := gin.New()

	http.NewHttp(g, f)

	if err := g.Run(fmt.Sprintf(":%d", config.AppPort())); err != nil {
		log.Fatal("Can't start server.")
	}
}
