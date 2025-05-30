# Atlas configuration
# Docs:
# - Syntax: https://atlasgo.io/atlas-schema/hcl
# - ORM support: https://atlasgo.io/guides/orms/gorm
# - Standalone mode: https://atlasgo.io/guides/orms/gorm/standalone

# Run this program, capture the output, use that as the schema.
data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "cmd/atlas-loader/main.go",
  ]
}

variable "database_url" {
  type    = string
  default = getenv("DATABASE_URL")
}

# data.external_schema.gorm -> Run the program above to get schema info
# data.external_schema.gorm.url -> The schema definition the program produced
env "local" {
  src = data.external_schema.gorm.url
  # Use docker for temp dev database (Atlas manages this automatically)
  # doc: https://atlasgo.io/concepts/dev-database
  dev = "docker://postgres/13/dev?search_path=public"
  # URL to local, actual DB
  url = var.database_url
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \" \" }}"
    }
  }
}

env "production" {
  src = data.external_schema.gorm.url
  url = var.database_url
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
