package storage

import (
	"context"
)

type ToMock struct {
	Data *Uitem
}

func (stormock *ToMock) GetCatalogByUser(ctx context.Context) {
}

func (stormock *ToMock) NewItems(ctx context.Context) {
}

func (stormock *ToMock) UpdateItems(ctx context.Context) {
}

func (stormock *ToMock) LoginUser(ctx context.Context) {
	switch stormock.Data.User.Login {
	case "correct":
		stormock.Data.User.Error = nil
	case "wrong":
		stormock.Data.User.Error = ErrLoginNotFound
		return
	default:
	}

	switch stormock.Data.User.Pass {
	case "correct":
		stormock.Data.User.Error = nil
		return
	case "wrong":
		stormock.Data.User.Error = ErrPasswordWrong
		return
	default:
	}
}

func (stormock *ToMock) CreateUser(ctx context.Context) {
	switch stormock.Data.User.Login {
	case "ErrUserExists":
		stormock.Data.Err = ErrUserExists
		return
	default:
		stormock.Data.Err = nil
		stormock.Data.User.UserID = "CorrectUserID"
		return
	}
}

func (stormock *ToMock) Connect(ctx context.Context) error {
	return nil
}

func (stormock *ToMock) Disconnect() {
}
