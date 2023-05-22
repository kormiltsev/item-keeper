package serverstorage

import "context"

// Storager is for DB
type Storager interface {
	RegUser() error
	AuthUser() error
	PutItems() error
	UploadFile() error
	UpdateByLastUpdate() error
	GetFileByFileID() error
	DeleteItems() error
	DeleteFile() error
	// UpdateItemsImageLinks()
	// DeleteItem(context.Context)
	// CreateUser(context.Context)
	// LoginUser(context.Context)
	Connect(context.Context) error
	Disconnect()
}
