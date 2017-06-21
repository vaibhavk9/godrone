package main

import (
	"errors"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	pb "github.com/zang-cloud/micro-registration-auth/protos"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func grpcServiceAuthClient() error {
	log.Infoln("Registering GRPC service auth client.. ", grpc_addr)

	conn, err := grpc.Dial(grpc_addr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithBackoffMaxDelay(30*time.Second))
	if err != nil {
		return err
	}

	AuthClient = &ServiceAuth{
		ClientConn:    conn,
		ServiceClient: pb.NewServiceAuthClient(conn),
	}

	return nil
}

func CreateNewAppClient(cl *Client, ttlVal string) error {

	log.Infoln("Create New App client grpc call... ")

	ttl, _ := strconv.Atoi(ttlVal)

	request := pb.ClientRegisterRequest{}
	request.Client = &pb.Client{}
	request.Client.AccountSid = cl.AccountSid
	request.Client.ApplicationSid = cl.ApplicationSid
	request.Client.Nickname = cl.Nickname
	request.Client.PresenceStatus = "init"
	request.Client.Ttl = int64(ttl)
	request.Client.RemoteIp = cl.RemoteIp

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := AuthClient.ServiceClient.Create(ctx, &request)

	if err != nil {
		return err
	}

	if response.Status != pb.ResponseCode_OK {
		log.Errorln("Failure Creating Application Client", response.Error)
		return errors.New("Failure Creating Application Client" + response.Error)
	} else {
		cl.ClientPassword = response.Client.ClientToken
		cl.DateCreated = response.Client.DateCreated.Format(time.ANSIC)
		cl.DateUpdated = response.Client.DateUpdated.Format(time.ANSIC)
	}

	return nil

}

func GetClientBySID(csid string) (Client, error) {

	log.Infoln("Get App client by ClienSid grpc call... ")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := AuthClient.ServiceClient.GetClientByClientSid(ctx, &pb.ClientId{ClientSid: csid})

	if err != nil {
		return Client{}, err
	}

	var c Client

	if resp.Status != pb.ResponseCode_OK {

		return Client{}, errors.New("Application Error " + resp.Err)
	} else {

		for _, cl := range resp.Clients {
			c.AccountSid = cl.AccountSid
			c.ApplicationSid = cl.ApplicationSid
			c.ClientSid = cl.Sid
			c.ClientPassword = cl.ClientToken
			c.DateCreated = cl.DateCreated.Format(time.ANSIC)
			c.DateUpdated = cl.DateUpdated.Format(time.ANSIC)
			c.Nickname = cl.Nickname
			c.PresenceStatus = cl.PresenceStatus
		}
	}

	return c, nil
}

/*
	Args: preinitialized Client struct ,page , pageSize
	Returns : Client slice,total record count , grpc service/client error
*/
func ListAppClients(c Client, page int32, pageSize int32) ([]Client, int64, error) {
	log.Infoln("List All Application Clients grpc call... ")

	var offset int32

	if page < 1 {
		offset = 0

	} else {
		offset = page * pageSize
	}

	in := &pb.FetchInputFields{
		AccountSid:     c.AccountSid,
		ApplicationSid: c.ApplicationSid,
		Offset:         offset,
		Limit:          pageSize,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := AuthClient.ServiceClient.GetClientListByFetchFields(ctx, in)

	if err != nil {
		return nil, 0, err
	}

	var ClientArr []Client

	if resp.Status != pb.ResponseCode_OK {

		return nil, 0, errors.New("Application Error " + resp.Err)

	} else {

		for _, client := range resp.Clients {
			c.ClientSid = client.Sid
			c.ClientPassword = client.ClientToken
			c.DateCreated = client.DateCreated.Format(time.ANSIC)
			c.DateUpdated = client.DateUpdated.Format(time.ANSIC)
			c.Nickname = client.Nickname
			c.PresenceStatus = client.PresenceStatus
			ClientArr = append(ClientArr, c)
		}
	}

	return ClientArr, resp.TotalCount, nil

}

func DeleteClients(csid string) error {
	log.Infoln("DeleteClients grpc call... ")

	id := &pb.ClientId{ClientSid: csid}

	var cids pb.ClientIds

	cids.ClientSids = append(cids.ClientSids, id)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := AuthClient.ServiceClient.DeleteClientsWithCheck(ctx, &cids)

	if err != nil {
		return err
	}

	if resp.Status != pb.ResponseCode_OK {
		return errors.New(resp.Err)
	}

	return nil
}
