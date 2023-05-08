package main

import (
	"context"
	"log"

	client "github.com/kormiltsev/item-keeper/internal/client"
	pb "github.com/kormiltsev/item-keeper/internal/server/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	// additem
	addTest("mockuser1", "Parameter1", "val11", "")
	addTest("mockuser1", "Parameter2", "val12", "./cmd/mockclient/testfile")
	addTest("mockuser2", "Parameter1", "val12", "./cmd/mockclient/testfile")
	idToDelete := catalogTest("mockuser1")
	// catalogFileTest("mockuser2")
	deleteTest("mockuser1", idToDelete)

	log.Println("mock client completed")
}

func addTest(userid, paraneter, pvalue, fileaddress string) {
	ctx := context.Background()
	cc := client.NewClientConnector(userid)
	defer client.CloseConnection()

	uit := pb.Uitem{Params: make([]*pb.Parameter, 0), Images: make([]*pb.Image, 0)}
	uit.Userid = userid
	uit.Params = append(uit.Params, &pb.Parameter{Name: paraneter, Value: pvalue})
	uit.Params = append(uit.Params, &pb.Parameter{Name: "Parameter2", Value: "val2"})
	uit.Images = append(uit.Images, &pb.Image{Title: fileaddress})

	// set new cc
	cc.Items = make([]*pb.Uitem, 1)
	cc.Items[0] = &uit
	cc.UserID = userid
	cc.LastUpdate = 0
	cc.Items[0].Userid = cc.UserID
	err := cc.AddNewItem(ctx)
	if err != nil {
		log.Println("FAIL: ADD: error in additind item to server:", err)
		return
	}

	if len(cc.Items) > 1 || len(cc.Items) == 0 {
		log.Println("FAIL: ADD: returns not only one item, quantity=", len(cc.Items))
	}

	if cc.Items[0].Lastupdate == 0 {
		log.Println("FAIL: ADD: new last update not set from server, cc.Items[0].Lastupdate=", cc.Items[0].Lastupdate)
	}

	if cc.Items[0].Id == "" {
		log.Println("FAIL: ADD: item id not set from server, cc.Items[0].Id=", cc.Items[0].Id)
	}
}

func catalogTest(userid string) string {
	ctx := context.Background()
	cc := client.NewClientConnector(userid)
	defer client.CloseConnection()

	cc.LastUpdate = 0

	// send lastupdate date from client to server and request catalog if LUclient != LUserver
	cc.ListOfItems(ctx)

	// if there new data on server, upload it to ram
	if cc.LastUpdate == 0 {
		log.Println("FAIL: CATALOG: last update = 0")
	}

	if len(cc.Items) < 2 {
		log.Println("FAIL: CATALOG: returns less than 2 items, quantity=", len(cc.Items))
	}
	idToDelete := cc.Items[0].Id

	//repeat
	//cc.LastUpdate = server's LU
	localLU := cc.LastUpdate

	// send lastupdate date from client to server and request catalog if LUclient != LUserver
	cc.ListOfItems(context.Background())

	// if there new data on server, upload it to ram
	if cc.LastUpdate != localLU {
		log.Println("FAIL: CATALOG: last update changed, =", cc.LastUpdate)
	}

	if len(cc.Items) != 0 {
		log.Println("FAIL: CATALOG: returns some items, quantity=", len(cc.Items))
	}

	return idToDelete
}

func catalogFileTest(userid string) {
	ctx := context.Background()
	cc := client.NewClientConnector(userid)
	defer client.CloseConnection()

	cc.LastUpdate = 0

	// send lastupdate date from client to server and request catalog if LUclient != LUserver
	cc.ListOfItems(ctx)

	// if there new data on server, upload it to ram
	if cc.LastUpdate == 0 {
		log.Println("FAIL: CATALOG: last update = 0")
	}

	if len(cc.Items) == 0 {
		log.Println("FAIL: CATALOG: returns zero items for user 2")
		return
	}

	if len(cc.Items[0].Images) != 1 {
		log.Println("FAIL: CATALOG: returns not 1 Image for user 2, quantity =", len(cc.Items[0].Images))
		return
	}

	if len(cc.Items[0].Images[0].Body) == 0 {
		log.Println("FAIL: CATALOG: returns empty image.body")
	}
}

func deleteTest(userid, itemid string) {
	ctx := context.Background()
	cc := client.NewClientConnector(userid)
	defer client.CloseConnection()

	// check not exists
	uit := pb.Uitem{Id: "not Existet ITEM"}
	cc.Items = make([]*pb.Uitem, 1)
	cc.Items[0] = &uit

	err := cc.DeleteItem(ctx)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() != codes.NotFound {
				log.Println("FAIL: DELETE: wrong response code:", err)
			}
		} else {
			log.Println("FAIL: DELETE: returns err:", err)
		}
	}

	//delete for sure
	uit = pb.Uitem{Id: itemid}
	cc.Items = make([]*pb.Uitem, 1)
	cc.Items[0] = &uit

	err = cc.DeleteItem(ctx)
	if err != nil {
		log.Println("FAIL: DELETE: returns err:", err)
	}

	if cc.LastUpdate == cc.Items[0].Lastupdate {
		log.Println("FAIL: DELETE: last Update not changed:", cc.LastUpdate)
	}

	// check deleted
	uit = pb.Uitem{Id: itemid}
	cc.Items = make([]*pb.Uitem, 1)
	cc.Items[0] = &uit

	err = cc.DeleteItem(ctx)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() != codes.NotFound {
				log.Println("FAIL: DELETE: wrong response code:", err)
			}
		} else {
			log.Println("FAIL: DELETE: returns err:", err)
		}

	}
}
