package serverstorage

// Storager is for DB
type Storager interface {
	RegUser() error
	AuthUser() error
	// UpdateItems(context.Context)
	// UpdateItemsImageLinks()
	// DeleteItem(context.Context)
	// CreateUser(context.Context)
	// LoginUser(context.Context)
	// Connect(context.Context) error
	// Disconnect()
}
