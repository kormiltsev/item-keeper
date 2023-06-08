package serverstorage

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	pgx "github.com/jackc/pgx/v5"

	"github.com/jackc/pgx/v5/pgxpool"
	configs "github.com/kormiltsev/item-keeper/internal/configs"
)

var db *pgxpool.Pool

// ToPostgres is interface
type ToPostgres struct {
	Data *ToStorage
}

const (
	useridkey = `usersids`
)

const (
	registerChanges = `
INSERT INTO itemkeeper_changes(userid, itemid, fileid, updateded_at)
VALUES ($1, $2, $3, NOW())
RETURNING id
;`
	// get last update
	getLastUpdate = `
SELECT COALESCE(MAX(id),0) FROM itemkeeper_changes
;`

	// rewrite body
	updateItem = `
UPDATE itemkeeper_items
SET body = $1, deleted = $2
WHERE id = $3
		;`
)

func shifu(a int) (string, error) {
	key := sha256.Sum256([]byte(useridkey))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := []byte("awsome_nonce")

	ciphertext := aesgcm.Seal(nil, nonce, []byte(strconv.Itoa(a)), nil)

	export := hex.EncodeToString(ciphertext)
	return export, nil
}

func (postg *ToPostgres) RegUser(ctx context.Context) error {
	if postg.Data.User.Login == "" || postg.Data.User.Password == "" {
		return ErrEmptyRequest
	}

	// write to postgres
	regUser := `
		INSERT INTO itemkeeper_users(login, password, created_at)
		VALUES ($1, $2, NOW())
		RETURNING id
	;`

	// ON CONFLICT (login)
	// DO NOTHING

	regUserSetID := `
	UPDATE itemkeeper_users
	SET userid = $1
	WHERE id = $2;
`
	tx, err := db.Begin(ctx)
	if err != nil {
		log.Println("1: begin error:", err)
	}

	var id int
	er := db.QueryRow(ctx, regUser, postg.Data.User.Login, postg.Data.User.Password).Scan(&id)

	// _, er := tx.Exec(ctx, regUser, postg.Data.User.Login, postg.Data.User.Password, postg.Data.User.UserID)
	if er != nil {
		errortext := er.Error()

		if errortext[len(errortext)-6:len(errortext)-1] == "23505" {
			log.Println("user exists")
			return ErrUserExists
		}

		// no rows returned, its OK, But if othr error then rollback
		if !errors.Is(er, pgx.ErrNoRows) {
			log.Println("PG reg err:", er)

			err := tx.Rollback(ctx)
			if err != nil {
				log.Println("regUser: Rollback err:", err)
				return err
			}

			return er
		}
	}

	// generate UserID
	postg.Data.User.UserID, err = shifu(id)
	if err != nil {
		err := tx.Rollback(ctx)
		if err != nil {
			log.Println("regUser: Rollback err:", err)
			return err
		}
		return err
	}

	// add userid string (generated from id)
	_, er = tx.Exec(ctx, regUserSetID, postg.Data.User.UserID, id)
	if er != nil {
		err := tx.Rollback(ctx)
		if err != nil {
			log.Println("regUser update: Rollback err:", err)
			return err
		}
		log.Println("PG reg EXEC:", er)
		return er
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Println("regUser: commit err: ", err)
		return err
	}

	// get last update
	err = db.QueryRow(ctx, getLastUpdate).Scan(&postg.Data.User.LastUpdate)
	if err != nil {
		log.Println("postgres Last update err: ", err)
		return fmt.Errorf("storage BD err:%v", err)
	}

	// // send to client current server time
	// postg.Data.User.LastUpdate = lu

	return nil
}

func (postg *ToPostgres) AuthUser(ctx context.Context) error {
	if postg.Data.User.Login == "" || postg.Data.User.Password == "" {
		return ErrEmptyRequest
	}

	// write to postgres
	regUser := `
		SELECT id, password, userid FROM itemkeeper_users WHERE login=$1
	;`

	var id int
	var passwd, userid string
	err := db.QueryRow(ctx, regUser, postg.Data.User.Login).Scan(&id, &passwd, &userid)
	switch err {
	case nil:
		// check password
		if passwd == postg.Data.User.Password {
			// return userID
			postg.Data.User.UserID = userid

			// send to client current server time
			postg.Data.User.LastUpdate = time.Now().UnixMilli()

			return nil
		}
		return ErrPasswordWrong
	case pgx.ErrNoRows:
		return ErrLoginNotFound
	default:
		log.Println("postgres GET err: ", err)
		return fmt.Errorf("storage BD err:%v", err)
	}
}

func (postg *ToPostgres) PutItems(ctx context.Context) error {
	// error if empty
	if len(postg.Data.List) == 0 {
		return ErrEmptyRequest
	}
	// id serial primary key,
	// userid TEXT not null,
	// body TEXT,
	// files TEXT[],
	// deleted BOOLEAN,
	// uploaded_at TIMESTAMPTZ DEFAULT Now()

	// write to postgres
	putItem := `
			INSERT INTO itemkeeper_items(userid, body, deleted)
			VALUES ($1, $2, FALSE)
			RETURNING id
		;`

	tx, err := db.Begin(ctx)
	if err != nil {
		log.Println("1: begin error:", err)
	}

	// for all items
	for i, item := range postg.Data.List {
		// Update != 0 means update
		if item.ItemID != 0 {
			_, er := tx.Exec(ctx, updateItem, item.Body, item.Deleted, item.ItemID)
			if er == nil {
				continue // updated and go to next item
			}
			log.Println("PG update error:", er)
			// if err we try to save item as new, coz may be wrong itemid (or bad idea?)
		}

		// add new item
		var id int64
		err := tx.QueryRow(ctx, putItem, item.UserID, item.Body).Scan(&id)
		if err != nil {
			// erase ItemID to mark itemn as error one
			postg.Data.List[i].ItemID = 0
			log.Println("PG put item error:", err)
			continue
		}
		postg.Data.List[i].ItemID = id

		// register changes
		var changesid int64
		err = tx.QueryRow(ctx, registerChanges, item.UserID, id, 0).Scan(&changesid)
		if err != nil {
			// erase ItemID to mark itemn as error one
			postg.Data.List[i].ItemID = 0
			log.Println("PG put item (changes log) error:", err)
			continue
		}
		logNewChange(item.UserID, id, 0, changesid) // go?
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Println("PG commit err: ", err)
		for i := range postg.Data.List {
			postg.Data.List[i].ItemID = 0
			// nothing were saved
			return fmt.Errorf("internal database error:%v", err)
		}
	}
	return nil
}

// UploadFile register file in table, returns go to save file somewhere, then update itemtable and changes
func (postg *ToPostgres) UploadFile(ctx context.Context) error {
	log.Println("file upload:", postg.Data.File.ItemID)
	// check body
	if len(postg.Data.File.Body) == 0 {
		return fmt.Errorf("empty file.Body in request uploadFile: %d", postg.Data.File.FileID)
	}

	// write to postgres
	putFile := `
			INSERT INTO itemkeeper_files(userid, itemid, deleted)
			VALUES ($1, $2, FALSE)
			RETURNING id
		;`

	tx, err := db.Begin(ctx)
	if err != nil {
		log.Println("1: begin error:", err)
	}

	var id int64
	err = tx.QueryRow(ctx, putFile, postg.Data.File.UserID, postg.Data.File.ItemID).Scan(&id)
	// _, er := tx.Exec(ctx, putFile, postg.Data.File.UserID, postg.Data.File.ItemID, false)
	if err != nil {
		log.Println("PG putFile error:", err)
		er := tx.Rollback(ctx)
		if er != nil {
			log.Println("PG putFile Rollback err:", err)
			return er
		}
		return err
	}

	// register changes
	var changesid int64
	err = tx.QueryRow(ctx, registerChanges, postg.Data.File.UserID, postg.Data.File.ItemID, id).Scan(&changesid)
	if err != nil {
		log.Println("PG upload file (changes log) error:", err)
		er := tx.Rollback(ctx)
		if er != nil {
			log.Println("PG putFile Rollback err:", err)
			return er
		}
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Println("PG commit err: ", err)
		for i := range postg.Data.List {
			postg.Data.List[i].ItemID = 0
			// nothing were saved
			return fmt.Errorf("internal database error:%v", err)
		}
	}

	// add logs to RAM
	logNewChange(postg.Data.File.UserID, postg.Data.File.ItemID, id, changesid) // go?

	// send file to storage
	go fileUploadToFileStorage(id, &postg.Data.File)

	return nil
}

func (postg *ToPostgres) UpdateByLastUpdate(ctx context.Context) error {
	log.Println("lastupdate requested:", postg.Data.User.LastUpdate)
	postg.Data.List = postg.Data.List[:0]
	postg.Data.FilesNoBody = postg.Data.FilesNoBody[:0]

	// get items
	getItemsUpdated := `
	SELECT id, userid, body, deleted FROM itemkeeper_items
	WHERE id IN(
	  SELECT itemid FROM itemkeeper_changes 
	  WHERE id > $1
	)
		;`

	// get files
	getFilesUpdated := `
	SELECT id, userid, itemid, deleted FROM itemkeeper_files
	WHERE id IN(
	  SELECT itemid FROM itemkeeper_changes 
	  WHERE id > $1
	)
		;`

	// get last update
	var lu int64
	err := db.QueryRow(ctx, getLastUpdate).Scan(&lu)
	if err != nil {
		log.Println("postgres Last update err: ", err)
		return fmt.Errorf("storage BD err:%v", err)
	}

	// items query
	rows, err := db.Query(ctx, getItemsUpdated, postg.Data.User.LastUpdate)
	switch err {
	case nil:
	case pgx.ErrNoRows:
		log.Println("item list is up to date for user", postg.Data.User.UserID)
	default:
		log.Println("postgres getItemsUpdated err: ", err)
		return fmt.Errorf("storage BD err:%v", err)
	}

	for rows.Next() {
		var newitem = Item{}
		err := rows.Scan(&newitem.ItemID, &newitem.UserID, &newitem.Body, &newitem.Deleted)
		if err != nil {
			return fmt.Errorf("POSTGRES rows.Scan error: %v", err)
		}
		postg.Data.List = append(postg.Data.List, newitem)
	}

	// files query
	rows, err = db.Query(ctx, getFilesUpdated, postg.Data.User.LastUpdate)
	switch err {
	case nil:
	case pgx.ErrNoRows:
		log.Println("files list is up to date for user", postg.Data.User.UserID)
	default:
		log.Println("postgres getFilesUpdated err: ", err)
		return fmt.Errorf("storage BD err:%v", err)
	}

	for rows.Next() {
		var newfile = File{}
		err := rows.Scan(&newfile.FileID, &newfile.UserID, &newfile.ItemID, &newfile.Deleted)
		if err != nil {
			return fmt.Errorf("POSTGRES rows.Scan error: %v", err)
		}
		postg.Data.FilesNoBody = append(postg.Data.FilesNoBody, newfile)
	}

	postg.Data.User.LastUpdate = lu
	return nil
}

func (postg *ToPostgres) GetFileByFileID(ctx context.Context) error {
	getFile := `
	SELECT id, userid, itemid, deleted FROM itemkeeper_files
	WHERE id = $1
		;`

	// get last update
	var fileNoBody = File{}
	err := db.QueryRow(ctx, getFile, postg.Data.File.FileID).Scan(&fileNoBody.FileID, &fileNoBody.UserID, &fileNoBody.ItemID, &fileNoBody.Deleted)
	switch err {
	case nil:
	case pgx.ErrNoRows:
		log.Println("file not found in DB")
		return ErrItemNotFound
	default:
		log.Println("postgres GET file by id err: ", err)
		return fmt.Errorf("storage BD err:%v", err)
	}

	if postg.Data.File.FileID == fileNoBody.FileID && postg.Data.File.UserID == fileNoBody.UserID {
		postg.Data.File.Body, err = fileDownloadFromStorage(&fileNoBody)
		if err != nil {
			log.Println("fileDownloadFromStorage error: ", err)
			return ErrItemNotFound
		}
		// add itemId, coze itemId is not mandatory in request
		postg.Data.File.ItemID = fileNoBody.ItemID
		return nil
	}

	return ErrItemNotFound
}

func (postg *ToPostgres) DeleteItems(ctx context.Context) error {
	// delete items
	if len(postg.Data.List) == 0 {
		return ErrEmptyRequest
	}

	// delete if contain some
	postg.Data.FilesNoBody = postg.Data.FilesNoBody[:0]

	// delete item
	deleteItem := `
	UPDATE itemkeeper_items
	SET deleted = TRUE
	WHERE id = $1 AND userid = $2
			;`

	// delete file
	deleteFile := `
			SELECT id, userid, itemid FROM itemkeeper_files
			WHERE itemid = $1 AND userid = $2
			;
	UPDATE itemkeeper_files
	SET deleted = TRUE
	WHERE itemid = $3 AND userid = $4
			;`

	// delete file by fileid
	deleteFileByID := `
	SELECT id, userid, itemid FROM itemkeeper_files
	WHERE id = $1 AND userid = $2
	;
UPDATE itemkeeper_files
SET deleted = TRUE
WHERE id = $3 AND userid = $4
	;`

	getItem := `
	SELECT deleted FROM itemkeeper_items
	WHERE id = $1 AND userid = $2
		;`

	getFile := `
		SELECT deleted FROM itemkeeper_files
		WHERE itemid = $1 AND userid = $2
			;`

	getFileByID := `
			SELECT deleted FROM itemkeeper_files
			WHERE id = $1 AND userid = $2
				;`

	tx, err := db.Begin(ctx)
	if err != nil {
		log.Println("tx delete: begin error:", err)
	}

	for _, itemtodelete := range postg.Data.List {
		if itemtodelete.ItemID == 0 && len(itemtodelete.FilesID) == 0 {
			// empty request, doing next item
			continue
		}

		// if there is file id so need to delete onli this file (don't delete item)
		for _, fileIDToDelete := range itemtodelete.FilesID {

			// delete item
			rows, err := tx.Query(ctx, deleteFileByID, fileIDToDelete, itemtodelete.UserID, fileIDToDelete, itemtodelete.UserID)
			switch err {
			case nil:
			case pgx.ErrNoRows:
				log.Println("deleteItem item not found to delete", itemtodelete.ItemID)
			default:
				// try to get file with error
				var deleted bool
				err = tx.QueryRow(ctx, getFileByID, fileIDToDelete, itemtodelete.UserID).Scan(&deleted)
				if err == nil && !deleted {
					log.Println("deletion error but file exists")
					er := tx.Rollback(ctx)
					if er != nil {
						log.Println("deleteFile and query Rollback err:", er)
						return er
					}
					log.Println("PG deletion file error:", err)
					return err
				}
				// if file not found of mark 'deleted' so ok, move next
			}

			for rows.Next() {
				var newfile = File{}
				err := rows.Scan(&newfile.FileID, &newfile.UserID, &newfile.ItemID)
				if err != nil {
					return fmt.Errorf("POSTGRES rows.Scan error: %v", err)
				}
				postg.Data.FilesNoBody = append(postg.Data.FilesNoBody, newfile)
			}

		}
		// delete item
		_, err := tx.Exec(ctx, deleteItem, itemtodelete.ItemID, itemtodelete.UserID)
		if err != nil {

			// try to get item with error
			var deleted bool
			err := tx.QueryRow(ctx, getItem, itemtodelete.ItemID, itemtodelete.UserID).Scan(&deleted)
			if err == nil && !deleted {
				log.Println("deletion error but item exists")
				er := tx.Rollback(ctx)
				if er != nil {
					log.Println("deleteItem and query Rollback err:", er)
					return er
				}
				log.Println("PG deletion error:", err)
				return err
			}
		}

		// delete item
		rows, err := tx.Query(ctx, deleteFile, itemtodelete.ItemID, itemtodelete.UserID, itemtodelete.ItemID, itemtodelete.UserID)
		switch err {
		case nil:
		case pgx.ErrNoRows:
			log.Println("deleteItem item not found to delete", itemtodelete.ItemID)
		default:
			// try to get file with error
			var deleted bool
			err = tx.QueryRow(ctx, getFile, itemtodelete.ItemID, itemtodelete.UserID).Scan(&deleted)
			if err == nil && !deleted {
				log.Println("deletion error but file exists")
				er := tx.Rollback(ctx)
				if er != nil {
					log.Println("deleteFile and query Rollback err:", er)
					return er
				}
				log.Println("PG deletion file error:", err)
				return err
			}
			// if file not found of mark 'deleted' so ok, move next
		}

		for rows.Next() {
			var newfile = File{}
			err := rows.Scan(&newfile.FileID, &newfile.UserID, &newfile.ItemID)
			if err != nil {
				return fmt.Errorf("POSTGRES rows.Scan error: %v", err)
			}
			postg.Data.FilesNoBody = append(postg.Data.FilesNoBody, newfile)
		}

	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Println("regUser: commit err: ", err)
		return err
	}

	go deleteFilesByID(postg.Data)

	return nil
}

func (postg *ToPostgres) DeleteFile(ctx context.Context) error {
	return nil
}

// Connect make connection with DB or panic
func (postg *ToPostgres) Connect(ctx context.Context) error {
	// connect to DB
	poolConfig, err := pgxpool.ParseConfig(configs.ServiceConfig.DBlink)
	if err != nil {
		log.Println("Unable to parse database_url:", err)
		return err
	}
	log.Println(poolConfig)

	db, err = pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Println("Unable to create connection pool:", err)
		return err
	}

	// users table
	var users = `
		CREATE TABLE IF NOT EXISTS itemkeeper_users(
			id serial primary key,
		  login VARCHAR(128) not null unique,
		  password VARCHAR(128) not null,
		  userid TEXT,
		  created_at TIMESTAMPTZ DEFAULT Now()
		);
	  `
	_, err = db.Exec(ctx, users)
	if err != nil {
		log.Println("error in create table users:", err)
	}

	var items = `
			CREATE TABLE IF NOT EXISTS itemkeeper_items(
				id serial primary key,
			  userid TEXT not null,
			  body TEXT,
			  files TEXT[],
			  deleted BOOLEAN,
			  uploaded_at TIMESTAMPTZ DEFAULT Now()
			);
		  `
	_, err = db.Exec(ctx, items)
	if err != nil {
		log.Println("error in create table items:", err)
	}

	// files table
	var files = `
		CREATE TABLE IF NOT EXISTS itemkeeper_files(
			id serial primary key,
			userid TEXT not null,
			itemid BIGINT not null,
			fileid BIGINT,
			deleted BOOLEAN,
			uploaded_at TIMESTAMPTZ DEFAULT Now()
		);
	  `
	_, err = db.Exec(ctx, files)
	if err != nil {
		log.Println("error in create table users:", err)
	}

	// changes table
	var changes = `
		CREATE TABLE IF NOT EXISTS itemkeeper_changes(
			id serial primary key,
			userid TEXT not null,
			itemid BIGINT not null,
			fileid BIGINT,
			updateded_at TIMESTAMPTZ DEFAULT Now()
		);
	  `
	_, err = db.Exec(ctx, changes)
	if err != nil {
		log.Println("error in create table users:", err)
	}
	return err
}

// Disconnect close all connections
func (postg *ToPostgres) Disconnect() {
	db.Close()
}
