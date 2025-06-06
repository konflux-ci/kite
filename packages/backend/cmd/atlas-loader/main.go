package main

import (
	"io"
	"log"
	"os"

	"ariga.io/atlas-provider-gorm/gormschema"
	"github.com/konflux-ci/kite/internal/models"
)

func main() {
	// Load all the models, generate SQL statements for them.
	stmts, err := gormschema.New("postgres").Load(
		&models.IssueScope{},
		&models.Issue{},
		&models.Link{},
		&models.RelatedIssue{},
	)

	if err != nil {
		log.Fatalf("failed to load gorm schema: %v", err)
	}

	// Output statements to stdout
	_, err = io.WriteString(os.Stdout, stmts)
	if err != nil {
		log.Fatalf("Unexpected error, got: %v", err)
	}
}
