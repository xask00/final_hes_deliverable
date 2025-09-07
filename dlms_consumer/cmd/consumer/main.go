package main

import (
	"context"
	"dlms_consumer/proto"
	"fmt"
	"io"
	"log"
	"log/slog"
	"time"

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
				Ip:             "2401:4900:833f:2688:0000:0000:0000:0002",
				Port:           40591,
				SystemTitle:    "6162636465666768",
				AuthPassword:   "0000000000000000",
				BlockCipherKey: "49423031494230324942303349423034",
				AuthKey:        "49423031494230324942303349423034",
				ClientAddress:  "48",
				ServerAddress:  "1",
				Obis:           "1.0.1.8.0.255",
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
		if resp == nil {
			slog.Error("Response is nil")
			break
		}

		fmt.Println("Received OBIS", resp)
		fmt.Println("response: ", resp)
	}
	time.Sleep(5 * time.Second)

	fmt.Println("Done with OBIS")

	// Now test GetBlockLoadProfile
	fmt.Println("\n=== Getting Block Load Profile ===")

	profileStream, err := client.GetBlockLoadProfile(context.Background(), &proto.GetBlockLoadProfileRequest{
		Meter: []*proto.Meter{
			{
				Ip:             "2401:4900:833f:2688:0000:0000:0000:0002",
				Port:           4059,
				SystemTitle:    "6162636465666768",
				AuthPassword:   "0000000000000000",
				BlockCipherKey: "49423031494230324942303349423034",
				AuthKey:        "49423031494230324942303349423034",
				ClientAddress:  "48",
				ServerAddress:  "1",
				Obis:           "1.0.1.8.0.255",
			},
		},
		Retries:           3,
		RetryDelay:        1000,
		ConnectionTimeout: 5000,
	})

	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.InvalidArgument {
				log.Fatalf("Invalid argument error: %s", st.Message())
			}
		}
		log.Fatalf("Failed to get Block Load Profile: %v", err)
	}

	for {
		fmt.Println("Receiving Block Load Profile")
		profileResp, err := profileStream.Recv()
		if err == io.EOF {
			fmt.Println("Block Load Profile EOF")
			break
		}
		if err != nil {
			if st, ok := status.FromError(err); ok {
				if st.Code() == codes.InvalidArgument && st.Message() == "no meters provided" {
					log.Fatalf("Specific error: no meters provided")
				}
			}
			log.Fatalf("Failed to receive Block Load Profile: %v", err)
		}
		if profileResp == nil {
			log.Fatalf("Block Load Profile response is nil")
		}

		fmt.Printf("Received Block Load Profile from meter %s:\n", profileResp.MeterIp)
		profile := profileResp.Profile
		fmt.Printf("  DateTime: %s\n", profile.DateTime)
		fmt.Printf("  Average Voltage: %.2f V\n", profile.AverageVoltage)
		fmt.Printf("  Block Energy Wh Import: %.2f Wh\n", profile.BlockEnergyWhImport)
		fmt.Printf("  Block Energy VAh Import: %.2f VAh\n", profile.BlockEnergyVahImport)
		fmt.Printf("  Block Energy Wh Export: %.2f Wh\n", profile.BlockEnergyWhExport)
		fmt.Printf("  Block Energy VAh Export: %.2f VAh\n", profile.BlockEnergyVahExport)
		fmt.Printf("  Average Current: %.2f A\n", profile.AverageCurrent)
		fmt.Printf("  Meter Health Indicator: %d\n", profile.MeterHealthIndicator)
		fmt.Println("---")
	}

	fmt.Println("Done with Block Load Profile")
}
