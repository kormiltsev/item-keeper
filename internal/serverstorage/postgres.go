package serverstorage

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
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
		ON CONFLICT (login)
		DO NOTHING
		RETURNING id
	;`

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
			return ErrUserExists
		}
		err := tx.Rollback(ctx)
		if err != nil {
			log.Println("regUser: Rollback err:", err)
			return err
		}
		return er
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
		return er
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Println("regUser: commit err: ", err)
		return err
	}

	// send to client current server time
	postg.Data.User.LastUpdate = time.Now().UnixMilli()

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
	return err
}

func (postg *ToPostgres) PutItems(ctx context.Context) error {
	// 	// error if empty
	// 	if len(postg.Data.List) == 0 {
	// 		return ErrEmptyRequest
	// 	}
	// 	id serial primary key,
	// 	userid TEXT not null,
	// 	body TEXT,
	// 	files TEXT[],
	// 	deleted BOOLEAN,
	// 	uploaded_at TIMESTAMPTZ DEFAULT Now()

	// 	// write to postgres
	// 	putItem := `
	// 		INSERT INTO itemkeeper_items(userid, body, deleted)
	// 		VALUES ($1, $2, FALSE)
	// 	;`

	// 	updateItem := `
	// 		UPDATE itemkeeper_items
	// 		SET body = $1, deleted = $3
	// 		WHERE id = $4
	// 	;`

	// 	tx, err := db.Begin(ctx)
	// 	if err != nil {
	// 		log.Println("1: begin error:", err)
	// 	}

	// 	// for all items
	// for i, item := range postg.Data.List {
	// 	// != 0 means update
	// 	if item.ItemID != 0 {
	// 		_, er := tx.Exec(ctx, updateItem, item.Body, item.Deleted, item.ItemID)
	// 		if er != nil {
	// 			// if error (probably id not existed) try to save as new
	// 			_, er := tx.Exec(ctx, putItem, item.UserID, item.Body)
	// 			if er != nil {
	// 				// erase ItemID to mark itemn as error one
	// 				postg.Data.List[i] = 0
	// 			}
	// 		}
	// 	}
	// }
	// 	_, er := tx.Exec(ctx, sqlPost, postg.Link.UserID, postg.Link.Alias, postg.Link.Original)
	// 	if er != nil {
	// 		errortext := er.Error()
	// 		if errortext[len(errortext)-6:len(errortext)-1] == "23505" {
	// 			postg.Link.Err = ErrConflictAlias
	// 			log.Println("PG POST alias exists: error 23505; new alias again..")
	// 			return
	// 		}
	// 		err := tx.Rollback(ctx)
	// 		if err != nil {
	// 			log.Println("error witn Rollback:", err)
	// 		}
	// 	}

	// 	err = tx.Commit(ctx)
	// 	if err != nil {
	// 		log.Println("3: commit err: ", err)
	// 	}
	return nil
}

func (postg *ToPostgres) UploadFile(ctx context.Context) error {
	return nil
}

func (postg *ToPostgres) UpdateByLastUpdate(ctx context.Context) error {
	return nil
}

func (postg *ToPostgres) GetFileByFileID(ctx context.Context) error {
	return nil
}

func (postg *ToPostgres) DeleteItems(ctx context.Context) error {
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
			  itemid TEXT not null,
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
			itemid TEXT not null,
			fileid TEXT not null,
			deleted BOOLEAN,
			uploaded_at TIMESTAMPTZ DEFAULT Now()
		);
	  `
	_, err = db.Exec(ctx, files)
	if err != nil {
		log.Println("error in create table users:", err)
	}
	return err

	// files table
	var changes = `
		CREATE TABLE IF NOT EXISTS itemkeeper_changes(
			id serial primary key,
			itemid TEXT not null,
			updateded_at BIGINT not nill,
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
