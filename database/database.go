package database

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/joho/godotenv"
	"log/slog"
	"os"
	"testBanyaRu/graph/model"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	psql *pgxpool.Pool
}

func Connect() *Database {
	ctx := context.Background()

	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(".env"); err != nil {
			slog.Error("Error loading .env file:", err)
		}
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		slog.Error("DATABASE_URL is not set in environment")
		os.Exit(1)
	}

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		slog.Error("Error connecting to postgres:", err)
		os.Exit(1)
	}

	if err := db.Ping(ctx); err != nil {
		slog.Error("Postgres is not available:", err)
		os.Exit(1)
	}

	fmt.Println("Connected to database")

	return &Database{
		psql: db,
	}
}

func (db *Database) Create(ctx context.Context, input model.MainInput) (*model.Main, error) {
	id := uuid.New().String()
	now := time.Now()

	subObj := map[string]interface{}{"tools": input.Tools, "tables": input.Tables, "chairs": input.Chairs}
	subObjJSON, _ := json.Marshal(subObj)

	var m model.Main
	var subObjRaw []byte
	var createdAt time.Time
	var updatedAt time.Time
	var deletedAt *time.Time

	err := db.psql.QueryRow(ctx, `
        INSERT INTO main (id, title, sub_id, sub_obj, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $5)
        RETURNING id, title, sub_id, sub_obj, created_at, updated_at, deleted_at
    `, id, input.Title, input.SubID, subObjJSON, now).Scan(
		&m.ID, &m.Title, &m.SubID, &subObjRaw, &createdAt, &updatedAt, &deletedAt,
	)
	if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
		return nil, fmt.Errorf("ID %s already exists", id)
	}
	if err != nil {
		return nil, err
	}

	m.CreatedAt = createdAt.Format(time.RFC3339)
	m.UpdatedAt = updatedAt.Format(time.RFC3339)
	if deletedAt != nil {
		dt := deletedAt.Format(time.RFC3339)
		m.DeletedAt = &dt
	}

	var subObjMap map[string]json.RawMessage
	_ = json.Unmarshal(subObjRaw, &subObjMap)
	_ = json.Unmarshal(subObjMap["tools"], &m.Tools)
	_ = json.Unmarshal(subObjMap["tables"], &m.Tables)
	_ = json.Unmarshal(subObjMap["chairs"], &m.Chairs)

	return &m, nil
}

func (db *Database) Get(ctx context.Context, id string) (*model.Main, error) {
	var m model.Main
	var subObjRaw []byte
	var createdAt time.Time
	var updatedAt time.Time
	var deletedAt *time.Time

	err := db.psql.QueryRow(ctx, `
        SELECT id, title, sub_id, sub_obj, created_at, updated_at, deleted_at
        FROM main WHERE id = $1 AND deleted_at IS NULL
    `, id).Scan(
		&m.ID, &m.Title, &m.SubID, &subObjRaw, &createdAt, &updatedAt, &deletedAt,
	)
	if err != nil {
		return nil, err
	}

	m.CreatedAt = createdAt.Format(time.RFC3339)
	m.UpdatedAt = updatedAt.Format(time.RFC3339)
	if deletedAt != nil {
		dt := deletedAt.Format(time.RFC3339)
		m.DeletedAt = &dt
	}

	var subObjMap map[string]json.RawMessage
	_ = json.Unmarshal(subObjRaw, &subObjMap)
	_ = json.Unmarshal(subObjMap["tools"], &m.Tools)
	_ = json.Unmarshal(subObjMap["tables"], &m.Tables)
	_ = json.Unmarshal(subObjMap["chairs"], &m.Chairs)

	return &m, nil
}

func (db *Database) Update(ctx context.Context, id string, input model.MainInput) (*model.Main, error) {
	now := time.Now()
	subObj := map[string]interface{}{"tools": input.Tools, "tables": input.Tables, "chairs": input.Chairs}
	subObjJSON, _ := json.Marshal(subObj)

	var m model.Main
	var subObjRaw []byte
	var createdAt time.Time
	var updatedAt time.Time
	var deletedAt *time.Time

	err := db.psql.QueryRow(ctx, `
        UPDATE main SET title = $1, sub_id = $2, sub_obj = $3, updated_at = $4
        WHERE id = $5
        RETURNING id, title, sub_id, sub_obj, created_at, updated_at, deleted_at
    `, input.Title, input.SubID, subObjJSON, now, id).Scan(
		&m.ID, &m.Title, &m.SubID, &subObjRaw, &createdAt, &updatedAt, &deletedAt,
	)
	if err != nil {
		return nil, err
	}

	m.CreatedAt = createdAt.Format(time.RFC3339)
	m.UpdatedAt = updatedAt.Format(time.RFC3339)
	if deletedAt != nil {
		dt := deletedAt.Format(time.RFC3339)
		m.DeletedAt = &dt
	}

	var subObjMap map[string]json.RawMessage
	_ = json.Unmarshal(subObjRaw, &subObjMap)
	_ = json.Unmarshal(subObjMap["tools"], &m.Tools)
	_ = json.Unmarshal(subObjMap["tables"], &m.Tables)
	_ = json.Unmarshal(subObjMap["chairs"], &m.Chairs)

	return &m, nil
}

func (db *Database) Delete(ctx context.Context, id string) (*model.MainDelete, error) {
	now := time.Now()
	cmdTag, err := db.psql.Exec(ctx, `UPDATE main SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`, now, id)
	if err != nil {
		return nil, err
	}
	if cmdTag.RowsAffected() == 0 {
		return nil, fmt.Errorf("nothing to delete for id %s", id)
	}
	return &model.MainDelete{DeleteID: id}, nil
}
