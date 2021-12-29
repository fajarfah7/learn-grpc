package main

import (
	"context"
	"fmt"
	"io"
	"learn-grpc/blog/domain"
	"log"

	"google.golang.org/grpc"
)

func main() {
	fmt.Println("---> Blog client <---\n")

	opts := grpc.WithInsecure()

	cc, err := grpc.Dial("localhost:50051", opts)
	defer cc.Close()
	if err != nil {
		log.Fatalf("could not connect: %v", err)
		return
	}

	c := domain.NewBlogServiceClient(cc)

	ctx := context.Background()

	// create blog
	req := domain.CreateBlogRequest{
		Blog: &domain.Blog{
			AuthorId: "fajar",
			Title:    "my third blog",
			Content:  "this is my content from my third blog",
		},
	}
	res, err := c.CreateBlog(ctx, &req)
	if err != nil {
		log.Fatalf("error getting response from server\n%v\n", err)
	}
	blogID := res.GetBlog().GetId()
	fmt.Printf("CreateBlogResponse: \n%v\n\n", res)

	// read blog
	_, err2 := c.ReadBlog(ctx, &domain.ReadBlogRequest{BlogId: "61cbdb743f1037d3b5b22eea"})
	if err2 != nil {
		fmt.Printf("error reading blog\n%v\n", err2)
	}

	readBlogReq := &domain.ReadBlogRequest{BlogId: blogID}
	readBlogRes, readBlogErr := c.ReadBlog(ctx, readBlogReq)
	if readBlogErr != nil {
		fmt.Printf("error reading blog\n%v\n", err)
	}
	fmt.Printf("ReadBlogResponse: \n%v\n\n", readBlogRes)

	// update blog
	newBlog := &domain.Blog{
		Id:       blogID,
		AuthorId: "fahrurozi",
		Title:    "my third blog was updated",
		Content:  "this is my content from my third blog that was updated",
	}
	updateBlogRes, updateBlogErr := c.UpdateBlog(ctx, &domain.UpdateBlogRequest{Blog: newBlog})
	if updateBlogErr != nil {
		fmt.Printf("error updating blog\n%v\n", updateBlogErr)
	}
	fmt.Printf("UpdateBlogResponse: \n%v\n\n", updateBlogRes)

	// delete blog
	deleteBlogRes, deleteBlogErr := c.DeleteBlog(ctx, &domain.DeleteBlogRequest{BlogId: blogID})
	if deleteBlogErr != nil {
		fmt.Printf("error delete blog\n%v\n", updateBlogErr)
	}
	fmt.Printf("DeleteBlogResponse: \n%v\n\n", deleteBlogRes)

	// list blog
	fmt.Printf("ListBlog:\n\n")
	stream, err := c.ListBlog(ctx, &domain.ListBlogRequest{})
	if err != nil {
		log.Fatalf("error while setup stream ListBlog\n%v\n", err)
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error while doing streaming ListBlog\n%v\n", err)
		}
		fmt.Println(res.GetBlog())
	}
}
