package main

// stephan mareek repository
// https://github.com/simplesteph/grpc-go-course/tree/master/blog

import (
	"context"
	"fmt"
	"learn-grpc/blog/database/mongodb"
	"learn-grpc/blog/domain"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
)

var collection *mongo.Collection

// Server conain server interface for Blog service
type Server struct{}

type BlogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Content  string             `bson:"content"`
	Title    string             `bson:"title" validate:"required"`
}

func (s *Server) CreateBlog(ctx context.Context, req *domain.CreateBlogRequest) (*domain.CreateBlogResponse, error) {
	// this will be in delivery(catch data)
	blog := req.GetBlog()
	fmt.Println("CreateBlog\n", blog)
	// on here should there are some validate for incoming data

	// this will be in delivery and usecase
	data := BlogItem{
		AuthorID: blog.GetAuthorId(),
		Title:    blog.GetTitle(),
		Content:  blog.GetContent(),
	}

	// this will be in repository
	res, err := collection.InsertOne(ctx, data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("internal error: %v", err))
	}
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("can not convert to oid"))
	}

	// this will be in delivery
	return &domain.CreateBlogResponse{
		Blog: &domain.Blog{
			Id:       oid.Hex(),
			AuthorId: blog.GetAuthorId(),
			Title:    blog.GetTitle(),
			Content:  blog.GetContent(),
		},
	}, nil
}

func (s *Server) ReadBlog(ctx context.Context, req *domain.ReadBlogRequest) (*domain.ReadBlogResponse, error) {
	blogID := req.GetBlogId()
	fmt.Println("ReadBlog blogID:\n", blogID)

	oid, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("can not parse id\n%v\n", err),
		)
	}

	data := new(BlogItem)
	filter := bson.M{"_id": oid}

	res := collection.FindOne(ctx, filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("data not found\n%v\n", err),
		)
	}

	return &domain.ReadBlogResponse{
		Blog: dataToBlogPb(data),
	}, nil
}

func (s *Server) UpdateBlog(ctx context.Context, req *domain.UpdateBlogRequest) (*domain.UpdateBlogResponse, error) {
	blog := req.GetBlog()
	fmt.Println("UpdateBlog\n", blog)

	oid, err := primitive.ObjectIDFromHex(blog.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("can not parse id\n%v\n", err))
	}

	data := new(BlogItem)
	filter := bson.M{"_id": oid}

	res := collection.FindOne(ctx, filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("data not found", err),
		)
	}

	data.AuthorID = blog.GetAuthorId()
	data.Content = blog.GetContent()
	data.Title = blog.GetTitle()

	_, updateErr := collection.ReplaceOne(ctx, filter, data)
	if updateErr != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("error while update data", updateErr),
		)
	}

	return &domain.UpdateBlogResponse{
		Blog: dataToBlogPb(data),
	}, nil
}

func (s *Server) DeleteBlog(ctx context.Context, req *domain.DeleteBlogRequest) (*domain.DeleteBlogResponse, error) {
	blogID := req.GetBlogId()
	fmt.Println("DeleteBlog\n", blogID)

	oid, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("error while parse blog id\n%v\n", err),
		)
	}

	filter := bson.M{"_id": oid}
	res, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("error while delete blog\n%v\n", err),
		)
	}

	if res.DeletedCount == 0 {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("can not find data database\n%v\n", err),
		)
	}

	return &domain.DeleteBlogResponse{BlogId: blogID}, nil
}

func (s *Server) ListBlog(req *domain.ListBlogRequest, stream domain.BlogService_ListBlogServer) error {
	fmt.Println("ListBlog")
	ctxBg := context.Background()
	cur, err := collection.Find(ctxBg, primitive.D{{}})
	if err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("unknown error occured when find data\n%v\n", err),
		)
	}

	defer cur.Close(ctxBg)
	for cur.Next(ctxBg) {
		data := new(BlogItem)
		err := cur.Decode(data)
		if err != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("error while decoding data\n%v\n", err),
			)
		}
		stream.Send(&domain.ListBlogResponse{Blog: dataToBlogPb(data)})
		time.Sleep(500 * time.Millisecond)
	}
	if err := cur.Err(); err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("unknown error occured\n%v\n", err),
		)
	}
	return nil
}

func dataToBlogPb(data *BlogItem) *domain.Blog {
	return &domain.Blog{
		Id:       data.ID.Hex(),
		AuthorId: data.AuthorID,
		Content:  data.Content,
		Title:    data.Title,
	}
}

func main() {
	// if we crash the go code, we get the file name and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Blog service started")

	// setup mongodb
	// fmt.Println("connecting to mongodb")
	// ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	// defer cancel()
	// client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	// if err != nil {
	// 	log.Fatalf("error while setup mongo client\n%v", err)
	// 	return
	// }
	// collection = client.Database("mydb").Collection("blog")

	fmt.Println("connecting to mongodb")
	m := mongodb.MongoDB{URI: "mongodb://localhost:27017", DB: "mydb", Table: "blog"}
	client, coll, err := m.Connect()
	if err != nil {
		log.Fatalf("failed to connect to mongodb\n%v\n", err)
		return
	}
	collection = coll

	listener, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("failed to setup listener \n%v\n", err)
		return
	}

	opts := []grpc.ServerOption{}
	server := grpc.NewServer(opts...)

	domain.RegisterBlogServiceServer(server, &Server{})

	go func() {
		fmt.Println("starting server")
		if err := server.Serve(listener); err != nil {
			log.Fatalf("failed to serve \n%v\n", err)
			return
		}
	}()

	// wait for ctrl + ct to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// block until a signal is received
	<-ch
	fmt.Println("stopping the server")
	server.Stop()

	fmt.Println("closing the listener")
	listener.Close()

	fmt.Println("closeing mongodb")
	client.Disconnect(context.Background())

	fmt.Println("end of program")
}
