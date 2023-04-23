package storage

import (
	"context"
)

type ToMock struct {
	Data *Uitem
}

func (stormock *ToMock) GetCatalogByUser(ctx context.Context) {
	return
}

func (stormock *ToMock) NewItems(ctx context.Context) {
	return
}

func (stormock *ToMock) UpdateItems(ctx context.Context) {
	return
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
	return
}

func (stormock *ToMock) CreateUser(ctx context.Context) {
	switch stormock.Data.User.Pass {
	case "correct":
		return
	}
	return
}

func (stormock *ToMock) Connect(ctx context.Context) error {
	return nil
}

func (stormock *ToMock) Disconnect() {
	return
}
