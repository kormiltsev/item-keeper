package serverstorage

import "context"

// Storager is for DB
type Storager interface {
	RegUser(context.Context) error
	AuthUser(context.Context) error
	PutItems(context.Context) error
	UploadFile(context.Context) error
	UpdateByLastUpdate(context.Context) error
	GetFileByFileID(context.Context) error
	DeleteItems(context.Context) error
	DeleteFile(context.Context) error
	Connect(context.Context) error
	Disconnect()
}
