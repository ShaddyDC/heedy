package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/connectordb/connectordb/api"
	"github.com/connectordb/connectordb/api/pb"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/afero"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	"github.com/connectordb/connectordb/assets"

	log "github.com/sirupsen/logrus"
)

var serverAddress = "localhost:3000"

func GetCert() (*tls.Certificate, *x509.CertPool) {
	serverCrt, err := ioutil.ReadFile("out/server.crt")
	if err != nil {
		log.Fatal(err)
	}
	serverKey, err := ioutil.ReadFile("out/server.key")
	if err != nil {
		log.Fatal(err)
	}

	pair, err := tls.X509KeyPair(serverCrt, serverKey)
	if err != nil {
		log.Fatal(err)
	}
	demoKeyPair := &pair
	demoCertPool := x509.NewCertPool()
	ok := demoCertPool.AppendCertsFromPEM(serverCrt)
	if !ok {
		log.Fatal("bad certs")
	}

	return demoKeyPair, demoCertPool
}

//https://github.com/dhrp/grpc-rest-go-example/blob/master/server/main.go
// grpcHandlerFunc returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise. Copied from cockroachdb.
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

// getRestMux initializes a new multiplexer, and registers each endpoint
// - in this case only the EchoService

func getRestMux(certPool *x509.CertPool, opts ...runtime.ServeMuxOption) (*runtime.ServeMux, error) {

	// Because we run our REST endpoint on the same port as the GRPC the address is the same.
	upstreamGRPCServerAddress := serverAddress

	// get context, this allows control of the connection
	ctx := context.Background()

	// These credentials are for the upstream connection to the GRPC server
	dcreds := credentials.NewTLS(&tls.Config{
		ServerName: upstreamGRPCServerAddress,
		//RootCAs:    certPool,
		InsecureSkipVerify: true,
	})
	dopts := []grpc.DialOption{grpc.WithTransportCredentials(dcreds)}

	// Which multiplexer to register on.
	// gwmux := runtime.NewServeMux()
	gwmux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard,
		&runtime.JSONPb{OrigName: true, EmitDefaults: true}))

	err := pb.RegisterPingHandlerFromEndpoint(ctx, gwmux, upstreamGRPCServerAddress, dopts)
	if err != nil {
		fmt.Printf("serve: %v\n", err)
		return nil, err
	}

	return gwmux, nil
}

func RunServer() {
	fmt.Println("Running server")
	a, err := assets.NewAssets("./testdb", nil)
	if err != nil {
		log.Error(err.Error())
		return
	}
	b, err := json.MarshalIndent(a.Config, "", " ")
	if err != nil {
		log.Error(err.Error())
		return
	}
	fmt.Println(string(b))

	/*
		f, err := a.AssetFS.Open("/setup/app.css")
		if err != nil {
			log.Error(err.Error())
			return
		}

		finfo, err := f.Stat()
		if err != nil {
			log.Error(err.Error())
			return
		}
		log.Println(finfo.Name())

		buf := new(bytes.Buffer)

		buf.ReadFrom(f)
		fmt.Println(buf.String())
	*/

	crt, pool := GetCert()

	errHandler := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		log.Printf("Got message")
		resp, err := handler(ctx, req)
		if err != nil {
			log.Printf("method %q failed: %s", info.FullMethod, err)
		}
		return resp, err
	}
	//creds := credentials.NewClientTLSFromCert(pool, serverAddress)
	// Start the gRPC server
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(errHandler)) //, grpc.Creds(creds))
	pb.RegisterPingServer(grpcServer, &api.API{})
	reflection.Register(grpcServer)

	restMux, err := getRestMux(pool)
	if err != nil {
		log.Panic(err)
	}

	// Set up the mux
	mux := http.NewServeMux()
	mux.Handle("/api/v1/cdb/", restMux)
	mux.Handle("/", http.FileServer(afero.NewHttpFs(a.AssetFS)))

	// the grpcHandlerFunc takes an grpc server and a http muxer and will
	// route the request to the right place at runtime.
	mergeHandler := grpcHandlerFunc(grpcServer, mux)

	// configure TLS for our server. TLS is REQUIRED to make this setup work.
	// check https://golang.org/src/net/http/server.go?#L2746
	if err != nil {
		log.Panic(err)
	}
	srv := &http.Server{
		Addr:    serverAddress,
		Handler: mergeHandler,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*crt},
			NextProtos:   []string{"h2"},
			//InsecureSkipVerify: true,
		},
	}
	// start listening on the socket
	// Note that if you listen on localhost:<port> you'll not be able to accept
	// connections over the network. Change it to ":port"  if you want it.
	conn, err := net.Listen("tcp", serverAddress)
	if err != nil {
		panic(err)
	}

	// start the server
	fmt.Printf("starting GRPC and REST on: %v\n", serverAddress)
	err = srv.Serve(tls.NewListener(conn, srv.TLSConfig))
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}