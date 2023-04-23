package storage

import (
	configs "github.com/kormiltsev/item-keeper/internal/configs"
)

func NewToStorage(uitem *Uitem) Storager {

	if configs.ServiceConfig.DBlink == "mock" || configs.ServiceConfig.DBlink == "" {
		stora := ToMock{}
		uitem.DB = &stora
		stora.Data = uitem
		return &stora
	}

	postgra := ToPostgres{
		db: db,
	}
	uitem.DB = &postgra
	postgra.Data = uitem
	return &postgra
}

func NewItem() *Uitem {
	return &Uitem{
		List: make([]Item, 0),
	}
}
