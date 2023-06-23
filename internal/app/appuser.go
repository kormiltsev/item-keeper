package app

import (
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	appstorage "github.com/kormiltsev/item-keeper/internal/app/appstorage"
	clientconnector "github.com/kormiltsev/item-keeper/internal/client"
	pb "github.com/kormiltsev/item-keeper/internal/server/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// variables define User and his password and last update date
var (
	currentuser string = "AppUser"
	// currentuserpassword    string
	currentuserencryptokey []byte = []byte("manualLocalKey")
	currentlastupdate      int64  = 0
)

// SaveUserCryptoPass save secret to var from UI.
func SaveUserCryptoPass(secretword []byte) {
	currentuserencryptokey = secretword
}

// RegUser request server to create new user. Returns error 'user exists'.
func RegUser(ctx context.Context, login, password string) error {

	// set context with time limit
	ctxto, cancel := context.WithTimeout(ctx, 5000*time.Millisecond)
	defer cancel()

	// encrypt login and password here
	login, password = encodeLoginPass(login, password)

	//buil request
	req := pb.RegUserRequest{
		Login:    login,
		Password: password,
	}

	// gRPC
	cc := clientconnector.NewClientConnector(ctxto)
	cl := *cc.Client

	// run request
	response, err := cl.RegUser(cc.Ctx, &req)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.AlreadyExists {
				return fmt.Errorf(`user exists:%v`, e.Message())
			} else {
				return fmt.Errorf(`regUser:%v:%v`, e.Code(), e.Message())
			}
		}
		return fmt.Errorf(`regUser error:%v`, err)
	}

	// save local
	appstorage.NewUser(response.Userid, 0)

	// save current user id
	currentuser = response.Userid
	currentlastupdate = response.Lastupdate

	// upload everithing
	err = UpdateDataFromServer(ctx)
	if err != nil {
		log.Println("can't update with new app user:", err)
	}

	return nil
}

// encodeLoginPass encode login and pawwsord to share with server.
func encodeLoginPass(login, password string) (string, string) {
	h := sha1.New()
	h.Write([]byte(login))
	login = hex.EncodeToString(h.Sum(nil))

	sum := sha256.Sum256([]byte(login + password))
	password = hex.EncodeToString(sum[:])

	return login, password
}

// AuthUser request server for User ID. Returns err 'wrong login/password'.
func AuthUser(ctx context.Context, login, password string) error {

	// set context with time limit
	ctxto, cancel := context.WithTimeout(ctx, 5000*time.Millisecond)
	defer cancel()

	// encrypt login and password here
	login, password = encodeLoginPass(login, password)

	// buil request
	req := pb.AuthUserRequest{
		Login:    login,
		Password: password,
	}

	// gRPC
	cc := clientconnector.NewClientConnector(ctxto)
	cl := *cc.Client

	// run request
	response, err := cl.AuthUser(cc.Ctx, &req)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.InvalidArgument {
				return fmt.Errorf(`wrong login/password:%v`, e.Message())
			} else {
				return fmt.Errorf(`AuthUser:%v:%v`, e.Code(), e.Message())
			}
		}
		return fmt.Errorf(`AuthUser error:%v`, err)
	}

	// save local
	// using NewUser() due to oneClient = oneUser rule (for now)
	appstorage.NewUser(response.Userid, 0)

	// save current user id
	currentuser = response.Userid
	currentlastupdate = 0

	// and go get catalog from server
	go UpdateDataFromServer(context.Background()) // used currentuser and currentlastupdate

	return nil
}
