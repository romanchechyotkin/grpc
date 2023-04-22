package main

import (
	"google.golang.org/grpc"
	"log"
	"net"
	pb "server/ecommerce"
)

const PORT = ":5000"
const NETWORK = "tcp"

var combinedShipmentMap map[string]pb.CombinedShipment

func main() {
	var srv = &server{
		orderMap: map[string]*pb.Order{
			"1": {
				Id:          "1",
				Items:       []string{"Apple", "Orange"},
				Description: "qwerty",
				Price:       1230,
				Destination: "Minsk",
			},
			"2": {
				Id:          "2",
				Items:       []string{"Apple"},
				Description: "qwerty",
				Price:       99901.1,
				Destination: "Minsk",
			},
			"3": {
				Id:          "3",
				Items:       nil,
				Description: "qwerty",
				Price:       0,
				Destination: "Minsk",
			},
			"4": {
				Id:          "4",
				Items:       []string{"Orange"},
				Description: "qwerty",
				Price:       12.321,
				Destination: "Minsk",
			},
			"5": {
				Id:          "5",
				Items:       []string{"Coca-cola"},
				Description: "qwerty",
				Price:       1234,
				Destination: "Minsk",
			},
		},
	}

	listen, err := net.Listen(NETWORK, PORT)
	if err != nil {
		log.Fatal(err)
	}
	s := grpc.NewServer(
		grpc.UnaryInterceptor(orderUnaryServerInterceptor),
		grpc.StreamInterceptor(orderStreamServerInterceptor),
	)
	pb.RegisterOrderManagementServer(s, srv)
	log.Printf("Starting gRPC listener on port " + PORT)
	if err = s.Serve(listen); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
