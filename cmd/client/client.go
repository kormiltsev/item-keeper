package main

// import client "github.com/kormiltsev/item-keeper/internal/client"
import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"strconv"
	"time"

	app "github.com/kormiltsev/item-keeper/internal/app"
	appstorage "github.com/kormiltsev/item-keeper/internal/app/appstorage"
	// logger "github.com/kormiltsev/item-keeper/internal/logger"
	// "go.uber.org/zap"
)

func main() {
	//redirect logger
	// blog := logger.NewLog("./configs/logger.json")
	// defer blog.Logger.Sync()
	// undo := zap.RedirectStdLog(blog.Logger)
	// defer undo()

	ctx := context.Background()
	testreqs(ctx)

	// testParameters(ctx)
}

func testreqs(ctx context.Context) {
	// generate random login/password
	sum := sha256.Sum256([]byte(strconv.FormatInt(time.Now().UnixNano(), 16)))
	login1 := "login_" + hex.EncodeToString(sum[:])
	password1 := "password_" + hex.EncodeToString(sum[:])

	log.Println("reg user 1")

	// create new user
	err := app.RegUser(ctx, login1, password1)
	if err != nil {
		log.Println("FAIL create user:", err)
		return
	}

	// create user exists
	err = app.RegUser(ctx, login1, "lll")
	if err == nil {
		log.Println(`FAIL create user: expect error "user exists", but resieved nil`)
		return
	}

	// add item 1 to user 1
	err = app.AddNewItem(ctx, presetItem(1))
	if err != nil {
		log.Println("FAIL create user:", err)
		return
	}

	// search by parameters
	search := app.NewSearchByParameter()
	search.Mapa["Name"] = []string{"1"}
	err = search.SearchItemByParameters()
	if err != nil {
		log.Println("FAIL search:", err)
		return
	}
	if len(search.Answer) == 0 {
		log.Println("Error should find at least 1 item:", search.Answer)
	}

	// set up for next user
	login2 := "2_" + login1
	password2 := "2_" + password1

	log.Println("reg user 2")

	// create new user2
	err = app.RegUser(ctx, login2, password2)
	if err != nil {
		log.Println("FAIL create user:", err)
		return
	}

	// add item 3 to user 2
	err = app.AddNewItem(ctx, presetItem(3))
	if err != nil {
		log.Println("FAIL add item 3:", err)
		return
	}

	// search by parameters
	search = app.NewSearchByParameter()
	search.Mapa["Color"] = []string{"red, Green"}
	search.Mapa["Manufacture"] = []string{"Mercedes"}
	err = search.SearchItemByParameters()
	if err != nil {
		log.Println("FAIL search:", err)
		return
	}
	if len(search.Answer) != 1 {
		log.Println("ERROR: len(answer) != 1")
	}

	log.Println("auth user 1")

	// valid authorization
	err = app.AuthUser(ctx, login1, password1)
	if err != nil {
		log.Println("FAIL authorization user:", err)
		return
	}

	// invalid authorization
	err = app.AuthUser(ctx, login2, "wrong")
	if err == nil {
		log.Println(`FAIL create user: expect error "wrong password", but resieved nil`)
		return
	}

	// add item 2 to user 1
	err = app.AddNewItem(ctx, presetItem(2))
	if err != nil {
		log.Println("FAIL add item 2:", err)
		return
	}

	// search by parameters
	search = app.NewSearchByParameter()
	search.Mapa["Name"] = []string{"1"}               // should find item 1 with 1 txt file
	search.Mapa["Manufacture"] = []string{"Mercedes"} // will not find
	search.Mapa["Has file"] = []string{"jpeg"}        // should find item 2 with 2 files
	err = search.SearchItemByParameters()
	if err != nil {
		log.Println("FAIL search:", err)
		return
	}
	if len(search.Answer) != 2 {
		log.Println("ERROR: len(answer) != 2")
	}
	var itemidfornext string
	for key, it := range search.Answer {
		files, errfiles, errf := app.RequestFilesByFileID(context.Background(), it.FileIDs...)
		if errf != nil {
			log.Printf("ERROR: %v,\nproblem file ids:%s", errf, errfiles)
			return
		}
		if len(files) == 0 {
			log.Printf("ERROR: 0 files returnd from request")
			return
		}
		itemidfornext = key
	}
	// check file saved
	//
	// ===============

	// delete item from prevous check
	todeleteitems := []string{itemidfornext}
	erroritems, errdel := app.DeleteItems(ctx, todeleteitems)
	if errdel != nil {
		log.Println("ERROR: delete item error with itemids:", erroritems)
	}

	// // search deleted item by parameters
	// search = app.NewSearchByParameter()
	// search.Mapa["Name"] = []string{"1"}
	// search.Mapa["Manufacture"] = []string{"Mercedes"} // will not find
	// err = search.SearchItemByParameters()
	// if err != nil {
	// 	log.Println("FAIL search:", err)
	// 	return
	// }

	log.Println("done testreqs")
}

func testParameters(ctx context.Context) {
	// generate random login/password
	sum := sha256.Sum256([]byte(strconv.FormatInt(time.Now().UnixNano(), 16)))
	loginPar := "login_" + hex.EncodeToString(sum[:])
	passwordPar := "password_" + hex.EncodeToString(sum[:])

	// create new user
	err := app.RegUser(ctx, loginPar, passwordPar)
	if err != nil {
		log.Println("FAIL create user:", err)
		return
	}

	// add item 10, 11, 12, 13, 14
	for i := 10; i < 15; i++ {
		err = app.AddNewItem(ctx, presetItem(i))
		if err != nil {
			log.Println("FAIL upload item", i, err)
			return
		}
	}

	// erase local data via switch user
	err = app.RegUser(ctx, "no", "no")
	if err != nil {
		err = app.AuthUser(ctx, loginPar, passwordPar)
		if err != nil {
			log.Println("FAIL authorization nono user:", err)
			return
		}
	}

	// valid authorization
	err = app.AuthUser(ctx, loginPar, passwordPar)
	if err != nil {
		log.Println("FAIL authorization user:", err)
		return
	}

	// app.UpdateDataFromServer(ctx)

	// search deleted item by parameters
	time.Sleep(1 * time.Second)
	search := app.NewSearchByParameter()
	search.Mapa["Name"] = []string{"Item"}
	// search.Mapa["Manufacture"] = []string{"Mercedes"} // will not find
	err = search.SearchItemByParameters()
	if err != nil {
		log.Println("FAIL search:", err)
		return
	}

	log.Println("done testParameters")
}

func presetItem(number int) *appstorage.Item {
	switch number {
	case 1:
		return &appstorage.Item{
			Parameters:    []appstorage.Parameter{{Name: "Name", Value: "Name Of Item 1"}, {Name: "Color", Value: "red"}, {Name: "Size", Value: "big one"}, {Name: "Has file", Value: "1 (txt)"}},
			UploadAddress: []string{"./data/sourceClient/test.txt"},
		}
	case 2:
		return &appstorage.Item{
			Parameters:    []appstorage.Parameter{{Name: "Name", Value: "Name Of Item 2"}, {Name: "Color", Value: "Green or red or what ever"}, {Name: "Has file", Value: "2 (txt and jpeg)"}},
			UploadAddress: []string{"./data/sourceClient/test.txt", "./data/sourceClient/Jocker.jpeg"},
		}
	case 3:
		return &appstorage.Item{
			Parameters:    []appstorage.Parameter{{Name: "Name", Value: "Name Of Item 3"}, {Name: "Manufacture", Value: "Mercedes"}, {Name: "Has file", Value: "2 (txt and jpeg)"}},
			UploadAddress: []string{"./data/sourceClient/test.txt", "./data/sourceClient/Jocker.jpeg"},
		}
	case 10:
		return &appstorage.Item{
			Parameters:    []appstorage.Parameter{},
			UploadAddress: []string{},
		}
	case 11:
		return &appstorage.Item{
			Parameters:    []appstorage.Parameter{{Name: "Name", Value: "Name Of Item 11"}},
			UploadAddress: []string{},
		}
	case 12:
		return &appstorage.Item{
			Parameters:    []appstorage.Parameter{{Name: "Name", Value: "Name Of Item 12"}, {Name: "Color", Value: "red"}},
			UploadAddress: []string{},
		}
	case 13:
		return &appstorage.Item{
			Parameters:    []appstorage.Parameter{{Name: "Name", Value: "Name Of Item 13"}, {Name: "Color", Value: "blue"}, {Name: "Size", Value: "big one"}},
			UploadAddress: []string{},
		}
	case 14:
		return &appstorage.Item{
			Parameters:    []appstorage.Parameter{{Name: "Name", Value: "Name Of Item 14"}, {Name: "Color", Value: "red"}, {Name: "Color", Value: "blue"}, {Name: "Type", Value: "Long one"}, {Name: "next", Value: "find me"}},
			UploadAddress: []string{},
		}
		return nil
	}
	return &appstorage.Item{
		Parameters:    []appstorage.Parameter{{Name: "Name", Value: "Default"}},
		UploadAddress: []string{},
	}
}
