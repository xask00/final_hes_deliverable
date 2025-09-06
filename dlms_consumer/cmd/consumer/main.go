package main

import (
	"context"
	"dlms_consumer/proto"
	"fmt"
	"io"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func main() {
	fmt.Println("Starting DLMS Consumer")
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}

	fmt.Println("Connected to server")

	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}

	client := proto.NewDLMSProcessorClient(conn)

	fmt.Println("Getting OBIS")
	stream, err := client.GetOBIS(context.Background(), &proto.GetOBISRequest{
		Meter: []*proto.Meter{
			{
				Ip:   "192.168.1.100",
				Port: 4059,
				Obis: "1.0.1.8.0.255",
			},
		},
		Obis: "1.0.1.8.0.255",
	})

	// Method 1: Check error code using status.FromError() and Code()
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.InvalidArgument {
				log.Fatalf("Invalid argument error: %s", st.Message())
			}
			// You can also check the message if needed
			if st.Code() == codes.InvalidArgument && st.Message() == "no meters provided" {
				log.Fatalf("Specific error: no meters provided")
			}
		}
		log.Fatalf("Failed to get OBIS: %v", err)
	}

	/* Alternative methods for checking gRPC errors:

	// Method 2: Using status.Code() directly (shorter)
	if err != nil && status.Code(err) == codes.InvalidArgument {
		log.Fatalf("Invalid argument error: %v", err)
	}

	// Method 3: Check both code and message
	if err != nil {
		st := status.Convert(err) // Convert always succeeds, unlike FromError
		if st.Code() == codes.InvalidArgument && st.Message() == "no meters provided" {
			log.Fatalf("Specific error: no meters provided")
		}
	}

	// Method 4: Switch on error codes for multiple checks
	if err != nil {
		switch status.Code(err) {
		case codes.InvalidArgument:
			log.Fatalf("Invalid argument: %v", err)
		case codes.NotFound:
			log.Fatalf("Not found: %v", err)
		case codes.Unavailable:
			log.Fatalf("Service unavailable: %v", err)
		default:
			log.Fatalf("Other error: %v", err)
		}
	}
	*/

	for {
		fmt.Println("Receiving OBIS")
		resp, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("EOF")
			break
		}
		if err != nil {
			if st, ok := status.FromError(err); ok {
				// if st.Code() == codes.InvalidArgument {
				// 	log.Fatalf("Invalid argument error: %s", st.Message())
				// }
				// You can also check the message if needed
				if st.Code() == codes.InvalidArgument && st.Message() == "no meters provided" {
					log.Fatalf("Specific error: no meters provided")
				}
			}
		}

		fmt.Println("Received OBIS", resp)
		fmt.Println("response: ", resp)
	}

	fmt.Println("Done")
}
