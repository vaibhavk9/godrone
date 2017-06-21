package main

import (
	"context"
	log "github.com/Sirupsen/logrus"
	pb "github.com/zang-cloud/micro-registration-auth/protos"
	"google.golang.org/grpc"
	"os"
	"os/signal"
)

var (
	grpc_addr = "localhost:8888"
	httpAddr  = ":8889"

	ParentContext context.Context
	ContextCancel context.CancelFunc
	AuthClient    *ServiceAuth
)

type ServiceAuth struct {
	*grpc.ClientConn
	ServiceClient pb.ServiceAuthClient
}

func init() {

	log.SetOutput(os.Stdout)

	if endpoint := os.Getenv("GRPC_SERVICE_AUTH_ENDPOINT"); len(endpoint) > 0 {
		grpc_addr = endpoint
	}

	if port := os.Getenv("ADDR"); len(port) > 0 {
		httpAddr = port
	}

	//TO Use Account MOCK setup
	if AccMock := os.Getenv("ACCOUNTS_MOCK"); len(AccMock) > 0 {
		os.Setenv("ACCOUNTS_MOCK", AccMock)
	}

}

func main() {

	log.Infoln("Starting Rest Authentication for App Client Service...")

	ParentContext, ContextCancel = context.WithCancel(context.Background())
	defer ContextCancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	httpErrChan := make(chan error)

	go func() {
		httpErrChan <- HttpServe()
	}()

	go grpcServiceAuthClient()

	select {

	case err := <-httpErrChan:
		log.Errorf("Error Starting HTTP service : %v", err.Error())

	case <-c:
		log.Infoln("OS Interrupt Signal! Exiting...  ")
		os.Exit(0)

	case <-ParentContext.Done():
		log.Infoln("Main Context Closed... ")
	}

}
