package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"google.golang.org/protobuf/types/known/wrapperspb"

	pb "server/ecommerce"
)

type server struct {
	orderMap map[string]*pb.Order
}

func (s *server) GetOrder(ctx context.Context, in *wrapperspb.StringValue) (*pb.Order, error) {
	order, exists := s.orderMap[in.Value]
	if exists {
		return order, nil
	}
	return nil, errors.New("no such order")
}

func (s *server) SearchOrders(searchQuery *wrapperspb.StringValue, stream pb.OrderManagement_SearchOrdersServer) error {
	log.Println("searching for", searchQuery.Value)
	for k, order := range s.orderMap {
		log.Println(k, order)

		for _, item := range order.Items {
			log.Println(item)

			if strings.Contains(item, searchQuery.Value) {
				err := stream.Send(order)
				if err != nil {
					return fmt.Errorf("error sending message to stream : %v", err)
				}
				log.Print("Matching Order Found : " + k)
				break
			}
		}
	}
	return nil
}

func (s *server) UpdateOrders(stream pb.OrderManagement_UpdateOrdersServer) error {
	ordersStr := "Updated Order IDs : "

	for {
		order, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&wrapperspb.StringValue{Value: "orders processed. " + ordersStr})
		}
		s.orderMap[order.Id] = order
		log.Println("Order ID", order.Id, "Updated")
		ordersStr += order.Id + ", "
	}

}

func (s *server) ProcessOrders(stream pb.OrderManagement_ProcessOrdersServer) error {

	for {
		orderId, err := stream.Recv()
		fmt.Println(orderId)
		if err == io.EOF {
			for _, comb := range combinedShipmentMap {
				stream.Send(&comb)
			}
			return nil
		}
		if err != nil {
			return err
		}
	}

}
