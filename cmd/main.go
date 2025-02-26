package main

import (
	grpc2 "channel-collector/api/grpc/service"
	"flag"
	"fmt"
	pb "github.com/Sujin1135/channel-collector-interface/protobuf/service"
	"google.golang.org/grpc"
	"log"
	"net"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	channelService := grpc2.NewChannelService()
	s := grpc.NewServer()
	pb.RegisterChannelServiceServer(s, channelService)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
