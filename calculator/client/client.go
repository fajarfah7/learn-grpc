package main

import (
	"context"
	"fmt"
	"io"
	"learn-grpc/calculator/pb"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	fmt.Println("creating request...")
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	defer cc.Close()
	if err != nil {
		log.Fatalf("can not connect to the server: %v", err)
	}

	c := pb.NewCalculatorServiceClient(cc)

	// doUnary(c)
	// doServerStream(c)
	// doClientStream(c)
	// doBiDiStreaming(c)

	doErrorUnary(c)
}

func doUnary(c pb.CalculatorServiceClient) {
	req := &pb.CalculatorRequest{
		Calculator: &pb.Calculator{
			Number_1: int32(3),
			Number_2: int32(10),
		},
	}

	res, err := c.Calculate(context.Background(), req)
	if err != nil {
		log.Fatalf("failed do request to the server: %v", err)
	}

	log.Println("got response from server is: ", res.Result)
}

func doServerStream(c pb.CalculatorServiceClient) {
	fmt.Println("starting to do a PrimeDecompositionRequest RPC...")
	ctxBg := context.Background()
	req := &pb.PrimeNumberDecompositionRequest{
		Number: int64(12390392840),
	}
	stream, err := c.PrimeNumberDecomposition(ctxBg, req)
	if err != nil {
		log.Fatalf("error while calling server stream: %v", err)
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("something happend: %v", err)
		}

		fmt.Println(res.GetPrimeFactor())
	}
}

func doClientStream(c pb.CalculatorServiceClient) {
	data := []int64{1, 2, 3, 4, 5, 6}

	stream, err := c.PrintAverage(context.Background())
	if err != nil {
		log.Fatalf("error while streaming client: %v\n", err)
	}

	for _, d := range data {
		fmt.Printf("send request: %v\n", d)
		stream.Send(&pb.PrintAverageRequest{Number: d})
		time.Sleep(1000 * time.Millisecond)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("failed get response from server: %v\n", err)
	}

	fmt.Println(res)
}

func doBiDiStreaming(c pb.CalculatorServiceClient) {
	fmt.Println("starting find maximum")
	stream, err := c.FindMaximum(context.Background())
	if err != nil {
		log.Fatalf("failed opening stream: %v", err)
	}

	waitc := make(chan struct{})
	go func() {
		numbers := []int32{4, 7, 2, 19, 4, 6, 32}
		for _, number := range numbers {
			fmt.Printf("sending request: %v\n", number)
			stream.Send(&pb.FindMaximumRequest{Number: number})
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()

	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("failed receive response from server: %v\n", err)
				break
			}

			maximum := res.GetMaximum()
			fmt.Printf("maximum %v\n", maximum)
		}
		close(waitc)
	}()
	<-waitc
}

func doErrorUnary(c pb.CalculatorServiceClient) {
	fmt.Println("starting doErrorUnary...")

	// correct
	doErrorCall(c, int32(10))

	// error
	doErrorCall(c, int32(-2))
}

func doErrorCall(c pb.CalculatorServiceClient, numb int32) {

	fmt.Printf("send request: %v\n", numb)
	res, err := c.SquareRoot(context.Background(), &pb.SquareRootRequest{Number: numb})
	if err != nil {
		respErr, ok := status.FromError(err)
		if ok {
			// actual error from gRPC(user error)
			log.Printf("error message: %v", respErr.Message())
			log.Printf("error code: %v\n", respErr.Code())
			if respErr.Code() == codes.InvalidArgument {
				log.Fatalf("we probably sent negative number\n")
				return
			}
		} else {
			log.Fatalf("big error calling SqareRoot: %v\n", err)
			return
		}
	}
	fmt.Printf("result of sqare root of %v is %v\n", numb, res.GetNumberRoot())
}
