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
	Connect(context.Context) error
	Disconnect()
}
