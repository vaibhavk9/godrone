package main

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	//	rest "github.com/zang-cloud/rest-framework"
	"net/http"
	//	"fmt"
	"context"
	helpers "github.com/zang-cloud/micro-common/helpers"
	"net"
	"strings"
)

func CreateApplicationClient(w http.ResponseWriter, req *http.Request) {
	log.Infoln("CreateApplicationClient call :")

	log.Infoln("Checking GRPC Service Auth Connection...")

	if EmptyStructCheck(AuthClient) {
		RenderReponseErr(w, errors.New("Could not establish grpc link with Service Auth Client"))
		return
	}

	params := mux.Vars(req)

	user_name := req.FormValue("nickname")

	Ip, _, _ := net.SplitHostPort(req.RemoteAddr)

	Ip = net.ParseIP(Ip).String()

	reqClient := Client{
		Nickname:       user_name,
		AccountSid:     params["AccountSid"],
		ApplicationSid: params["ApplicationSid"],
		ClientSid:      params["ClientSid"],
		RemoteIp:       Ip,
	}

	respErr := CreateNewAppClient(&reqClient, req.FormValue("ttl"))

	if respErr != nil {
		RenderServiceAuthErr(w, "AppClient Creation", respErr)
		return
	}

	c := SimpleResponse{
		Client: []Client{

			{
				Nickname:       user_name,
				ClientPassword: reqClient.ClientPassword,
				Uri:            req.URL.EscapedPath(),
				SessionId:      "none",
				AccountSid:     params["AccountSid"],
				ApplicationSid: params["ApplicationSid"],
				ClientSid:      params["ClientSid"],
				ApiVersion:     params["APIVersion"],
				RemoteIp:       Ip},
		},
	}

	ext := ReqFormat(params["format"])

	var ClientArg interface{}

	if ext == "csv" {
		ClientArg = c.Client
	} else {
		ClientArg = c
	}

	err := HandleResponseEncoding(w, ext, ClientArg)

	if err != nil {
		RenderEncodingErr(w, ext, err)
		return
	}

}

func GetApplicationClient(w http.ResponseWriter, req *http.Request) {
	log.Infoln("GetApplicationClient :")

	log.Infoln("Checking GRPC Service Auth Connection...")

	if EmptyStructCheck(AuthClient) {
		RenderReponseErr(w, errors.New("Could not establish grpc link with Service Auth Client"))
		return
	}

	params := mux.Vars(req)

	client, respErr := GetClientBySID(params["ClientSid"])

	if respErr != nil {
		RenderServiceAuthErr(w, "Get Application Client ", respErr)
		return
	}

	Ip, _, _ := net.SplitHostPort(req.RemoteAddr)

	Ip = net.ParseIP(Ip).String()

	cl := SimpleResponse{
		Client: []Client{

			{DateUpdated: client.DateUpdated,
				PresenceStatus: client.PresenceStatus,
				Nickname:       client.Nickname,
				ClientPassword: client.ClientPassword,
				Uri:            req.URL.EscapedPath(),
				SessionId:      "none",
				AccountSid:     client.AccountSid,
				ApplicationSid: client.ApplicationSid,
				ClientSid:      client.ClientSid,
				DateCreated:    client.DateCreated,
				ApiVersion:     params["APIVersion"],
				RemoteIp:       Ip},
		},
	}

	ext := ReqFormat(params["format"])

	var ClientArg interface{}

	if ext == "csv" {
		ClientArg = cl.Client
	} else {
		ClientArg = cl
	}

	err := HandleResponseEncoding(w, ext, ClientArg)

	if err != nil {
		RenderEncodingErr(w, ext, err)
		return
	}

}

func ListApplicationClients(w http.ResponseWriter, req *http.Request) {

	log.Infoln("ListApplicationClients :")

	log.Infoln("Checking GRPC Service Auth Connection...")

	if EmptyStructCheck(AuthClient) {
		RenderReponseErr(w, errors.New("Could not establish grpc link with Service Auth Client"))
		return
	}

	params := mux.Vars(req)

	page := helpers.ParsePage(req.FormValue("Page"), 0)
	pageSize := helpers.ParsePageSize(req.FormValue("PageSize"), 50)

	Ip, _, _ := net.SplitHostPort(req.RemoteAddr)

	Ip = net.ParseIP(Ip).String()

	c := Client{
		Uri:            req.URL.EscapedPath(),
		SessionId:      "none",
		AccountSid:     params["AccountSid"],
		ApplicationSid: params["ApplicationSid"],
		ApiVersion:     params["APIVersion"],
		RemoteIp:       Ip,
	}

	clientArr, totalCount, respErr := ListAppClients(c, int32(page), int32(pageSize))

	if respErr != nil {
		RenderServiceAuthErr(w, "List Application Client ", respErr)
		return
	}

	p := CreatePagination(req, page, pageSize, totalCount)
	p.Uri = req.URL.EscapedPath()

	resp := &Response{
		Clients: Clients{
			Pagination: p,
			Clients:    clientArr,
		},
	}

	ext := ReqFormat(params["format"])

	var ClientArg interface{}

	if ext == "csv" {
		ClientArg = resp.Clients.Clients
	} else {
		ClientArg = resp
	}

	err := HandleResponseEncoding(w, ext, ClientArg)

	if err != nil {
		RenderEncodingErr(w, ext, err)
		return
	}

}

func DeleteApplicationClient(w http.ResponseWriter, req *http.Request) {
	log.Infoln("DeleteApplicationClient :")

	log.Infoln("Checking GRPC Service Auth Connection...")

	if EmptyStructCheck(AuthClient) {
		RenderReponseErr(w, errors.New("Could not establish grpc link with Service Auth Client"))
		return
	}

	params := mux.Vars(req)

	err := DeleteClients(params["ClientSid"])

	if err != nil {
		RenderServiceAuthErr(w, "Delete Application Client ", err)
		return
	}

	Ip, _, _ := net.SplitHostPort(req.RemoteAddr)

	Ip = net.ParseIP(Ip).String()

	var page, pageSize int64
	page, pageSize = 0, 50

	c := Client{
		Uri:            req.URL.EscapedPath(),
		SessionId:      "none",
		AccountSid:     params["AccountSid"],
		ApplicationSid: params["ApplicationSid"],
		ApiVersion:     params["APIVersion"],
		RemoteIp:       Ip,
	}

	clientArr, totalCount, respErr := ListAppClients(c, int32(page), int32(pageSize))

	if respErr != nil {
		RenderServiceAuthErr(w, "List Application Client ", respErr)
		return
	}

	p := CreatePagination(req, page, pageSize, totalCount)
	p.Uri = req.URL.EscapedPath()

	resp := &Response{
		Clients: Clients{
			Pagination: p,
			Clients:    clientArr,
		},
	}

	ext := ReqFormat(params["format"])

	var ClientArg interface{}

	if ext == "csv" {
		ClientArg = resp.Clients.Clients
	} else {
		ClientArg = resp
	}

	err = HandleResponseEncoding(w, ext, ClientArg)

	if err != nil {
		RenderEncodingErr(w, ext, err)
		return
	}

}

func NoHandleFound(w http.ResponseWriter, req *http.Request) {
	log.Infof("Handle Not Found For Request - %v", req)
	http.Error(w, "Requested Resource not found...", http.StatusNotFound)
}

func HealthCehck(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func ReqContextWithAuth(muxRoute http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		if err := req.ParseForm(); err != nil {
			RenderFormParsingErr(w, err)
			return
		}

		accSid, authToken, _ := req.BasicAuth()

		if !strings.Contains(req.URL.EscapedPath(), "Health") {
			/*excep := rest.NewRequest(w, req).AuthRequired()

			if excep != nil {
				log.Infoln("Authentication failed...")
				return
			}*/
		}

		ctx := context.WithValue(req.Context(), "param", map[string]string{"Account_sid": accSid, "auth_Token": authToken})
		muxRoute.ServeHTTP(w, req.WithContext(ctx))
	})

}
