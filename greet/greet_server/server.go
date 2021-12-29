package main

// https://github.com/simplesteph/grpc-go-course/tree/master/ssl
import (
	"context"
	"fmt"
	"io"
	"learn-grpc/greet/greetpb"
	"log"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

type server struct{}

func (s *server) Greet(ctx context.Context, r *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	fmt.Println("Greet function was invoked with %v", r)
	firstName := r.GetGreeting().GetFirstName()

	result := "Hello " + firstName
	res := &greetpb.GreetResponse{
		Result: result,
	}

	return res, nil
}

func (s *server) GreetManyTimes(req *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer) error {
	fmt.Println("GreetManyTimes function was invoked with %v", req)
	firstName := req.GetGreeting().GetFirstName()
	for i := 0; i < 10; i++ {
		result := "Hello " + firstName + " number " + strconv.Itoa(i)
		res := &greetpb.GreetManyTimesResponse{
			Result: result,
		}
		stream.Send(res)
		time.Sleep(1000 * time.Millisecond)
	}
	return nil
}

func (s *server) LongGreet(stream greetpb.GreetService_LongGreetServer) error {
	fmt.Println("LongGreet function was invoked")
	result := ""
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&greetpb.LongGreetResponse{
				Result: result,
			})
		}
		if stream.Context().Err() == context.Canceled {
			fmt.Println("request was canceled")
			return err
		}
		if err != nil {
			log.Fatalf("error while reading client stream: %v", err)
			return err
		}
		firstName := req.GetGreeting().GetFirstName()
		result += "Hello " + firstName + "!"
	}
}

func (s *server) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error {
	fmt.Println("GreetEveryone function was invoked")

	for {
		req, err := stream.Recv()
		if stream.Context().Err() == context.Canceled {
			fmt.Println("request was canceled")
			return stream.Context().Err()
		}

		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("error while reading client stream: %v", err)
			return err
		}

		firstName := req.GetGreeting().GetFirstName()
		result := "hello " + firstName + "! "
		errSend := stream.Send(&greetpb.GreetEveryoneResponse{Result: result})
		if errSend != nil {
			log.Fatalf("error while send data to client: %v", errSend)
			return err
		}
	}
}

func (s *server) GreetWithDeadline(ctx context.Context, r *greetpb.GreetWithDeadlineRequest) (*greetpb.GreetWithDeadlineResponse, error) {
	fmt.Println("GreetWithDeadline function was invoked with %v", r)

	for i := 0; i < 3; i++ {
		if ctx.Err() == context.Canceled {
			fmt.Println("client canceled the request")
			return nil, status.Error(codes.DeadlineExceeded, "client canceled the request")
		}
		time.Sleep(1 * time.Second)
	}
	firstName := r.GetGreeting().GetFirstName()

	result := "Hello " + firstName
	res := &greetpb.GreetWithDeadlineResponse{Result: result}

	return res, nil
}

func main() {
	fmt.Println("OK")
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("failed to liste: %v", err)
	}
	tls := true
	opts := []grpc.ServerOption{}
	if tls {
		certFile := "ssl/server.crt"
		keyFile := "ssl/server.pem"
		creds, sslErr := credentials.NewServerTLSFromFile(certFile, keyFile)
		if sslErr != nil {
			log.Fatalf("failed loading certificate: %v", sslErr)
			return
		}
		opts = append(opts, grpc.Creds(creds))
	}

	s := grpc.NewServer(opts...)
	greetpb.RegisterGreetServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("falied to serve: %v", err)
	}
}
