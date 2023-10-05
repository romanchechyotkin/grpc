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
	log.Println("server streaming")
	searchingStream, err := client.SearchOrders(ctx, &wrapperspb.StringValue{Value: "Orange"})
	for {
		streamOrder, err := searchingStream.Recv()
		if err == io.EOF {
			break
		}
		log.Println("finding order: ", streamOrder)
	}

	log.Println("client streaming")
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
	streamProcOrder, _ := client.ProcessOrders(ctx)
	if err := streamProcOrder.Send(
		&wrapperspb.StringValue{Value: "1"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "1", err)
	}
	if err := streamProcOrder.Send(
		&wrapperspb.StringValue{Value: "2"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "2", err)
	}
	if err := streamProcOrder.Send(
		&wrapperspb.StringValue{Value: "3"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "3", err)
	}

	channel := make(chan struct{})
	go asncClientBidirectionalRPC(streamProcOrder, channel)

	time.Sleep(time.Microsecond * 1000)

	if err := streamProcOrder.Send(
		&wrapperspb.StringValue{Value: "4"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "4", err)
	}

	if err := streamProcOrder.CloseSend(); err != nil {
		log.Fatal(err)
	}

	<-channel
}

func asncClientBidirectionalRPC(stream pb.OrderManagement_ProcessOrdersClient, channel chan struct{}) {
	for {
		shipment, err := stream.Recv()
		if err == io.EOF {
			break
		}
		log.Println(shipment.OrdersList)
	}
	channel <- struct{}{}
}

func orderUnaryClientInterceptor(
	ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	log.Println("method", method)

	err := invoker(ctx, method, req, reply, cc, opts...)

	log.Println(reply)
	return err
}

func orderStreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc,
	cc *grpc.ClientConn, method string,
	streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	log.Println("======= [Client Interceptor] ", method)
	s, err := streamer(ctx, desc, cc, method, opts...)
	if err != nil {
		return nil, err
	}
	return newWrappedStream(s), nil
}

type wrappedStream struct {
	grpc.ClientStream
}

func newWrappedStream(s grpc.ClientStream) grpc.ClientStream {
	return &wrappedStream{s}
}

func (w *wrappedStream) RecvMsg(m interface{}) error {
	log.Printf("====== [Client Stream Interceptor] "+
		"Receive a message (Type: %T) at %v",
		m, time.Now().Format(time.RFC3339))
	return w.ClientStream.RecvMsg(m)
}

func (w *wrappedStream) SendMsg(m interface{}) error {
	log.Printf("====== [Client Stream Interceptor] "+
		"Send a message (Type: %T) at %v",
		m, time.Now().Format(time.RFC3339))
	return w.ClientStream.SendMsg(m)
}
