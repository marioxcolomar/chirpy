# Chirpy

# Motivation

# Quick start

# Usage

## Database queries
When adding queries in `sql/queries` the necessary functionaly would be created inside `internal/database` directory. The functionality can be updated by running `sqlc generate` after adding new queries.

## Migrations
For creating migrations we are using Goose (link to goose). Adding new migration should happen in `sql/schema` following the naming convention of number plus schema or migration context in the file name.
Running a migration can be done through the terminal using the goose command the postgres protocal and providing the connection URL to the database instance.
For example an up migration running locally would be:
`goose postgres postgres://postgres:@localhost/5432/chirpy up`
If you're username, password, port or database name differ make sure to change the command accordingly.
