package storage

import "context"

// Storager is for DB
type Storager interface {
	GetCatalogByUser(context.Context)
	NewItems(context.Context)
	UpdateItems(context.Context)
	DeleteItem(context.Context)
	CreateUser(context.Context)
	LoginUser(context.Context)
	Connect(context.Context) error
	Disconnect()
}
