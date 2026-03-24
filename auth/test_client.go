package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"auth/api/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial("localhost:8001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := proto.NewAuthClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("=== Testing Register API ===")
	registerReq := &proto.RegisterRequest{
		Username: "testuser123",
		Password: "password123",
		Email:    "testuser123@example.com",
	}

	registerResp, err := client.Register(ctx, registerReq)
	if err != nil {
		log.Printf("Register failed: %v", err)
	} else {
		fmt.Printf("Register Success!\n")
		fmt.Printf("  Token: %s\n", registerResp.Token)
		fmt.Printf("  User ID: %d\n", registerResp.UserInfo.Id)
		fmt.Printf("  Username: %s\n", registerResp.UserInfo.Username)
		fmt.Printf("  Email: %s\n", registerResp.UserInfo.Email)
	}

	fmt.Println("\n=== Testing Login API ===")
	loginReq := &proto.LoginRequest{
		Username: "testuser123",
		Password: "password123",
	}

	loginResp, err := client.Login(ctx, loginReq)
	if err != nil {
		log.Printf("Login failed: %v", err)
	} else {
		fmt.Printf("Login Success!\n")
		fmt.Printf("  Token: %s\n", loginResp.Token)
		fmt.Printf("  User ID: %d\n", loginResp.UserInfo.Id)
		fmt.Printf("  Username: %s\n", loginResp.UserInfo.Username)
		fmt.Printf("  Email: %s\n", loginResp.UserInfo.Email)
	}

	fmt.Println("\n=== Testing Login with Invalid Credentials ===")
	invalidLoginReq := &proto.LoginRequest{
		Username: "wronguser",
		Password: "wrongpass",
	}

	_, err = client.Login(ctx, invalidLoginReq)
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	} else {
		fmt.Println("ERROR: Login should have failed with invalid credentials!")
	}
}
