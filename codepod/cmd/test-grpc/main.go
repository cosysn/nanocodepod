package main

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	proto "github.com/codepod-io/codepod/proto"
)

func main() {
	conn, err := grpc.Dial("localhost:22028", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}
	defer conn.Close()

	client := proto.NewAgentClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test GetStatus
	status, err := client.GetStatus(ctx, &proto.StatusRequest{})
	if err != nil {
		fmt.Printf("GetStatus failed: %v\n", err)
	} else {
		fmt.Printf("Status: %s, Uptime: %ds\n", status.Status, status.UptimeSeconds)
	}

	// Test ExecuteCommand
	execResp, err := client.ExecuteCommand(ctx, &proto.CommandRequest{Command: "echo hello from grpc"})
	if err != nil {
		fmt.Printf("ExecuteCommand failed: %v\n", err)
	} else {
		fmt.Printf("Exit code: %d, Output: %s\n", execResp.ExitCode, execResp.Output)
	}
}
