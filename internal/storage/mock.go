package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	configs "github.com/kormiltsev/item-keeper/internal/configs"
)

type ToMock struct {
	Data *Uitem
}

var lastupdate int64 = time.Now().UnixMilli()

var ram = []Item{Item{
	ID:         "PresetItemID",
	UserID:     "userPreset",
	Name:       "Pre-uploaded in mock",
	Tags:       []string{},
	Parameters: map[string]string{"parameter preset": "preset value"},
	ImageLink:  make([]string, 0),
	TitleImage: make([]byte, 0),
	LastUpdate: lastupdate,
	Deleted:    false,
}}

func (stormock *ToMock) GetCatalogByUser(ctx context.Context) {
	if stormock.Data.User.LastUpdate == lastupdate {
		log.Println("same lastUpdate:", lastupdate)
		return
	}
	stormock.Data.User.LastUpdate = lastupdate

	//copy(stormock.Data.List, ram)
	stormock.Data.List = stormock.Data.List[:0]
	for _, item := range ram {
		if !item.Deleted {
			stormock.Data.List = append(stormock.Data.List, item)
		}
	}
}

func (stormock *ToMock) NewItems(ctx context.Context) {
	//add to storage
	for i := range stormock.Data.List {
		stormock.Data.List[i].ID = fmt.Sprintf("newupload_%d", len(ram))
		// ram = append(ram, v)
	}
	ram = append(ram, stormock.Data.List...)

	// send client last update time
	stormock.Data.User.LastUpdate = lastupdate

	// users last update date
	lastupdate = time.Now().UnixMilli()

	// send new client last update time (with new item added)
	stormock.Data.List[0].LastUpdate = lastupdate

	stormock.Data.Err = nil
}

func (stormock *ToMock) UpdateItems(ctx context.Context) {
	log.Println("edited id:", stormock.Data.List[0].ID)

	// mark item as deleted
	for i, it := range ram {
		if it.ID == stormock.Data.List[0].ID {
			ram[i] = stormock.Data.List[0]

			// send client last update time
			stormock.Data.User.LastUpdate = lastupdate

			// users last update date
			lastupdate = time.Now().UnixMilli()

			// send new client last update time (with new item added)
			stormock.Data.List[0].LastUpdate = lastupdate

			stormock.Data.Err = nil

			log.Println("edited:", ram[i])
			return
		}
	}
	stormock.Data.Err = ErrItemNotFound
}

func (stormock *ToMock) DeleteItem(ctx context.Context) {
	log.Println("stormock id:", stormock.Data.List[0].ID)

	// mark item as deleted
	for i, it := range ram {
		if it.ID == stormock.Data.List[0].ID {
			ram[i].Deleted = true

			// send client last update time
			stormock.Data.User.LastUpdate = lastupdate

			// users last update date
			lastupdate = time.Now().UnixMilli()

			// send new client last update time (with new item added)
			stormock.Data.List[0].LastUpdate = lastupdate

			stormock.Data.Err = nil
			return
		}
	}
	stormock.Data.Err = ErrItemNotFound

	log.Println("ram", ram)
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
	err := FileStoragePing(configs.ServiceConfig.FileServerAddress)
	if err != nil {
		log.Println("file storage fail:", err)
	}
	return nil
}

func (stormock *ToMock) Disconnect() {
}
