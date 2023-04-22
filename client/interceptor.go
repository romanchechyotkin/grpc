package main

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"time"
)

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
