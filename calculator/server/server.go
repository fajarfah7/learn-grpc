package main

import (
	"context"
	"fmt"
	"io"
	"learn-grpc/calculator/pb"
	"log"
	"math"
	"net"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

type ServerCalculator struct{}

// Calculate is unary
func (s *ServerCalculator) Calculate(ctx context.Context, req *pb.CalculatorRequest) (*pb.CalculatorResponse, error) {
	fmt.Println("receive request: %v", req)
	numb1 := req.GetCalculator().GetNumber_1()
	numb2 := req.GetCalculator().GetNumber_2()

	res := &pb.CalculatorResponse{
		Result: int32(numb1 + numb2),
	}

	return res, nil
}

// PrimeNumberDecomposition is server stream
func (s *ServerCalculator) PrimeNumberDecomposition(req *pb.PrimeNumberDecompositionRequest, stream pb.CalculatorService_PrimeNumberDecompositionServer) error {
	number := req.GetNumber()
	divisor := int64(2)

	for number > 1 {
		if number%divisor == 0 {
			stream.Send(&pb.PrimeNumberDecompositionResponse{
				PrimeFactor: divisor,
			})
			number = number / divisor
		} else {
			divisor++
			fmt.Println("divisor has increased to %v \n", divisor)
		}
	}

	return nil
}

// PrintAverage is client stream
func (s *ServerCalculator) PrintAverage(stream pb.CalculatorService_PrintAverageServer) error {
	fmt.Println("PrintAverage was invoked")
	var result int64 = 0
	var iteration int64 = 0
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			response := float64(result) / float64(iteration)
			return stream.SendAndClose(&pb.PrintAverageResponse{Result: response})
		}
		if err != nil {
			log.Fatalf("error while reading client stream: %v", err)
		}

		iteration++
		result += req.GetNumber()
	}
}

func (s *ServerCalculator) FindMaximum(stream pb.CalculatorService_FindMaximumServer) error {
	fmt.Println("FindMaximum was invoked")
	maximum := int32(0)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("error while reading client data: %v", err)
			return err
		}

		number := req.GetNumber()
		if number > maximum {
			maximum = number
			sendErr := stream.Send(&pb.FindMaximumResponse{Maximum: maximum})
			if sendErr != nil {
				log.Fatalf("error while sending data to client: %v", err)
			}
		}
	}
}

func (s *ServerCalculator) SquareRoot(ctx context.Context, req *pb.SquareRootRequest) (*pb.SquareRootResponse, error) {
	fmt.Println("SquareRoot was invoked")
	number := req.GetNumber()
	if number < 0 {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("receive negative number: %v\n", number),
		)
	}

	return &pb.SquareRootResponse{
		NumberRoot: math.Sqrt(float64(number)),
	}, nil
}

func main() {
	fmt.Println("server was started")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterCalculatorServiceServer(server, &ServerCalculator{})

	// register reflection service on gRPC server
	reflection.Register(server)

	err = server.Serve(lis)
	if err != nil {
		log.Fatalf("failed to start the server: %v", err)
	}
}
