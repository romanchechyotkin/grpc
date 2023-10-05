package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"

	pb "client/ecommerce"
)

const ADDRESS = "localhost:5000"

func main() {
	conn, err := grpc.Dial(
		ADDRESS,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(orderUnaryClientInterceptor),
		grpc.WithStreamInterceptor(orderStreamClientInterceptor),
	)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()
	client := pb.NewOrderManagementClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// унарный RPC
	order, err := client.GetOrder(ctx, &wrapperspb.StringValue{Value: "2"})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(order)

	// Потоковый RPC на стороне сервера
	searchingStream, err := client.SearchOrders(ctx, &wrapperspb.StringValue{Value: "Orange"})
	for {
		streamOrder, err := searchingStream.Recv()
		if err == io.EOF {
			break
		}
		log.Println("finding order: ", streamOrder)
	}

	// Потоковый RPC на стороне клиента
	updateStream, err := client.UpdateOrders(ctx)
	if err != nil {
		fmt.Println(err)
	}

	order, _ = client.GetOrder(ctx, &wrapperspb.StringValue{Value: "1"})
	fmt.Println(order)

	err = updateStream.Send(&pb.Order{
		Id:          "1",
		Items:       nil,
		Description: "lololool",
		Price:       10000000,
		Destination: "",
	})
	if err != nil {
		fmt.Println(err)
	}

	stringValue, _ := updateStream.CloseAndRecv()
	fmt.Println(stringValue)

	// Двунаправленный потоковый RPC

}
