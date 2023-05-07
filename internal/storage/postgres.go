package storage

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	configs "github.com/kormiltsev/item-keeper/internal/configs"
)

var db *pgxpool.Pool

// ToPostgres is Storager interface
type ToPostgres struct {
	db   *pgxpool.Pool
	Data *Uitem
}

// GetCatalogByUser(context.Context)
// NewItems(context.Context)
// UpdateItems(context.Context)
// NewStorager() *Storager
// FindOrCreateUser(context.Context)
// Connect() error
// Disconnect()

func (postg *ToPostgres) GetCatalogByUser(ctx context.Context) {
}

func (postg *ToPostgres) NewItems(ctx context.Context) {
}

func (postg *ToPostgres) UpdateItems(ctx context.Context) {
}

func (postg *ToPostgres) DeleteItem(ctx context.Context) {
}

func (postg *ToPostgres) CreateUser(ctx context.Context) {
}
func (postg *ToPostgres) LoginUser(ctx context.Context) {
}

// Connect make connection with DB or panic
func (postg *ToPostgres) Connect(ctx context.Context) error {
	// connect to DB
	poolConfig, err := pgxpool.ParseConfig(configs.ServiceConfig.DBlink)
	if err != nil {
		log.Println("Unable to parse database_url:", err)
		return err
	}
	log.Println(poolConfig)

	db, err = pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Println("Unable to create connection pool:", err)
		return err
	}

	var items = `
			CREATE TABLE IF NOT EXISTS keeper_items(
				id serial primary key,
			  userid INTEGER not null,
			  name TEXT not null,
			  tags TEXT[],
			  parameters TEXT[][],
			  picturelink TEXT[],
			  deleted BOOLEAN,
			  uploaded_at TIMESTAMPTZ DEFAULT Now()
			);
		  `
	_, err = db.Exec(ctx, items)
	if err != nil {
		log.Println("error in create table items:", err)
	}

	// users table
	var users = `
	CREATE TABLE IF NOT EXISTS keeper_users(
		id serial primary key,
	  login VARCHAR(128) not null unique,
	  pass VARCHAR(128) not null,
	  last_update_at TIMESTAMPTZ DEFAULT Now(),
	  created_at TIMESTAMPTZ DEFAULT Now()
	);
  `
	_, err = db.Exec(ctx, users)
	if err != nil {
		log.Println("error in create table users:", err)
	}
	return err
}

// Disconnect close all connections
func (postg *ToPostgres) Disconnect() {
	db.Close()
}
