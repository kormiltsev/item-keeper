package client

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"sync"
// 	"time"

// 	pb "github.com/kormiltsev/item-keeper/internal/server/proto"

// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/credentials/insecure"
// 	"google.golang.org/grpc/metadata"
// 	"google.golang.org/grpc/status"
// )

// type ClientSettings struct {
// 	port       string
// 	tokenName  string
// 	Connection *grpc.ClientConn
// }

// var ClientSet = ClientSettings{
// 	port:      ":3333",
// 	tokenName: "cli_user_token",
// }

// // NewClientConnector make connection once and then returns interface with client
// func NewClientConnector(userid string) *ClientConnector {
// 	var err error
// 	var cli pb.ItemKeeperClient
// 	var once sync.Once
// 	once.Do(func() {
// 		ClientSet.Connection, err = grpc.Dial(ClientSet.port, grpc.WithTransportCredentials(insecure.NewCredentials()))
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		cli = pb.NewItemKeeperClient(ClientSet.Connection)
// 	})
// 	return &ClientConnector{UserID: userid, Items: make([]*pb.Uitem, 0), User: pb.LogUserRequest{}, Client: &cli}
// }

// // CloseConnection before stop application
// func CloseConnection() {
// 	ClientSet.Connection.Close()
// }

// func (cc *ClientConnector) ListOfItems(ctx context.Context) {

// 	ctxto, cancel := context.WithTimeout(ctx, 5000*time.Millisecond)
// 	defer cancel()

// 	md := metadata.New(map[string]string{ClientSet.tokenName: cc.UserID})
// 	ctx = metadata.NewOutgoingContext(ctxto, md)

// 	req := pb.GetCatalogRequest{Userid: cc.UserID, LastUpdate: cc.LastUpdate}

// 	cl := *cc.Client
// 	uit, err := cl.GetCatalog(ctx, &req)
// 	if err != nil {
// 		log.Printf("add item error from server: %v\n", err)
// 		cc.Error = fmt.Errorf("get catalog from server error: %v", err)
// 		return
// 	}

// 	// copy result
// 	cc.Items = make([]*pb.Uitem, len(uit.Uitem))
// 	copy(cc.Items, uit.Uitem)

// 	cc.LastUpdate = uit.LastUpdate
// }

// // func RunTestClient(port string) {
// // 	ClientSet.port = port
// // 	conn, err := grpc.Dial(ClientSet.port, grpc.WithTransportCredentials(insecure.NewCredentials()))
// // 	if err != nil {
// // 		log.Fatal(err)
// // 	}
// // 	defer conn.Close()

// // 	c := pb.NewItemKeeperClient(conn)

// // 	ctxto, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
// // 	defer cancel()

// // 	md := metadata.New(map[string]string{ClientSet.tokenName: ClientConnector.UserID})
// // 	ctx := metadata.NewOutgoingContext(ctxto, md)

// // 	// tc1
// // 	PushUser := storagecli.User{
// // 		Login: "ErrUserExists",
// // 		Pass:  "password",
// // 	}
// // 	RegUser(ctx, PushUser, c)

// // 	// tc2
// // 	ClientSet.token = "TokenNotSet"
// // 	md = metadata.New(map[string]string{ClientSet.tokenName: ClientSet.token})
// // 	ctx = metadata.NewOutgoingContext(ctxto, md)
// // 	PushUser = storagecli.User{
// // 		Login: "NewLogin",
// // 		Pass:  "password",
// // 	}
// // 	RegUser(ctx, PushUser, c)
// // }

// // AddNewItem
// func (cc *ClientConnector) AddNewItem(ctx context.Context) error {

// 	ctxto, cancel := context.WithTimeout(ctx, 5000*time.Millisecond)
// 	defer cancel()

// 	md := metadata.New(map[string]string{ClientSet.tokenName: cc.UserID})
// 	ctx = metadata.NewOutgoingContext(ctxto, md)

// 	req := pb.AddItemRequest{
// 		Uitem: cc.Items[0],
// 	}

// 	cl := *cc.Client
// 	uit, err := cl.AddItem(ctx, &req)
// 	if err != nil {
// 		log.Printf("add item error from server: %v\n", err)
// 		return err
// 	}

// 	cc.Items[0].Id = uit.Uitem.Id
// 	cc.Items[0].Lastupdate = uit.Uitem.Lastupdate
// 	cc.LastUpdate = uit.OldLastUpdate

// 	// log.Printf("cc LU:%d, uit OldLU:%d", cc.LastUpdate, uit.OldLastUpdate)
// 	return nil
// }

// // DeleteItem gets id and returs Item not found error
// func (cc *ClientConnector) DeleteItem(ctx context.Context) error {

// 	ctxto, cancel := context.WithTimeout(ctx, 5000*time.Millisecond)
// 	defer cancel()

// 	md := metadata.New(map[string]string{ClientSet.tokenName: cc.UserID})
// 	ctx = metadata.NewOutgoingContext(ctxto, md)

// 	req := pb.DeleteItemRequest{
// 		Userid: cc.UserID,
// 		Itemid: cc.Items[0].Id,
// 	}

// 	cl := *cc.Client
// 	response, err := cl.DeleteItem(ctx, &req)
// 	if err != nil {
// 		if e, ok := status.FromError(err); ok {
// 			if e.Code() == codes.NotFound {
// 				fmt.Println(`item not found:`, e.Message())
// 			} else {
// 				fmt.Println(e.Code(), e.Message())
// 			}
// 		} else {
// 			log.Printf("add item error from server: %v\n", err)
// 		}
// 		cc.Error = err
// 		return err
// 	}

// 	cc.LastUpdate = response.OldLastUpdate
// 	cc.Items[0].Lastupdate = response.LastUpdate

// 	return nil
// }

// // EditItem gets new item and replase by id
// func (cc *ClientConnector) EditItem(ctx context.Context) error {

// 	ctxto, cancel := context.WithTimeout(ctx, 5000*time.Millisecond)
// 	defer cancel()

// 	md := metadata.New(map[string]string{ClientSet.tokenName: cc.UserID})
// 	ctx = metadata.NewOutgoingContext(ctxto, md)

// 	req := pb.UpdateItemRequest{
// 		Uitem: cc.Items[0],
// 	}

// 	cl := *cc.Client
// 	resp, err := cl.UpdateItem(ctx, &req)
// 	if err != nil {
// 		if e, ok := status.FromError(err); ok {
// 			if e.Code() == codes.NotFound {
// 				fmt.Println(`item not found:`, e.Message())
// 			} else {
// 				fmt.Println(e.Code(), e.Message())
// 			}
// 		} else {
// 			log.Printf("add item error from server: %v\n", err)
// 		}
// 		cc.Error = err
// 		return err
// 	}

// 	cc.LastUpdate = resp.OldLastUpdate

// 	return nil
// }

// // RegUser set up userid in client token (metadata). Error if user exists.
// // func RegUser(ctx context.Context, user storagecli.User, c pb.ItemKeeperClient) {

// // 	newuserid, err := c.RegUser(ctx, &pb.RegUserRequest{
// // 		Login: user.Login,
// // 		Pass:  user.Pass,
// // 	})
// // 	if err != nil {
// // 		if e, ok := status.FromError(err); ok {
// // 			if e.Code() == codes.AlreadyExists {
// // 				log.Println(`user already exists:`, e.Message())
// // 			} else {
// // 				log.Println(e.Code(), e.Message())
// // 			}
// // 		} else {
// // 			log.Printf("Reg user error from server: %v\n", err)
// // 		}
// // 	} else {
// // 		log.Println("user id:", newuserid.Userid)
// // 		ClientSet.token = newuserid.Userid
// // 		log.Println("client settings:", ClientSet)
// // 	}
// // }

// // DELETE
// // func  (cc *ClientConnector) SendFile(ctx context.Context) error {
// // for _, item := range cc.Items {
// // 	for _, file := range item.Images {
// // 	}
// // }
// // }
