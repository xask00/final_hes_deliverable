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

	fmt.Println("Waiting for 5 seconds")
	time.Sleep(5 * time.Second)
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

	fmt.Println("Waiting for 5 seconds")
	time.Sleep(5 * time.Second)

	// Test GetDailyLoadProfile
	fmt.Println("\n=== Getting Daily Load Profile ===")

	dailyProfileStream, err := client.GetDailyLoadProfile(context.Background(), &proto.GetDailyLoadProfileRequest{
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
		log.Fatalf("Failed to get Daily Load Profile: %v", err)
	}

	for {
		fmt.Println("Receiving Daily Load Profile")
		dailyResp, err := dailyProfileStream.Recv()
		if err == io.EOF {
			fmt.Println("Daily Load Profile EOF")
			break
		}
		if err != nil {
			if st, ok := status.FromError(err); ok {
				if st.Code() == codes.InvalidArgument && st.Message() == "no meters provided" {
					log.Fatalf("Specific error: no meters provided")
				}
			}
			log.Fatalf("Failed to receive Daily Load Profile: %v", err)
		}
		if dailyResp == nil {
			log.Fatalf("Daily Load Profile response is nil")
		}

		fmt.Printf("Received Daily Load Profile from meter %s:\n", dailyResp.MeterIp)
		daily := dailyResp.Profile
		fmt.Printf("  DateTime: %s\n", daily.DateTime)
		fmt.Printf("  Cumulative Energy Wh Export: %.2f Wh\n", daily.CumulativeEnergyWhExport)
		fmt.Printf("  Cumulative Energy VAh Export: %.2f VAh\n", daily.CumulativeEnergyVahExport)
		fmt.Printf("  Cumulative Energy Wh Import: %.2f Wh\n", daily.CumulativeEnergyWhImport)
		fmt.Printf("  Cumulative Energy VAh Import: %.2f VAh\n", daily.CumulativeEnergyVahImport)
		fmt.Println("---")
	}

	fmt.Println("Done with Daily Load Profile")

	fmt.Println("Waiting for 5 seconds")
	time.Sleep(5 * time.Second)

	// Test GetBillingDataProfile
	fmt.Println("\n=== Getting Billing Data Profile ===")

	billingProfileStream, err := client.GetBillingDataProfile(context.Background(), &proto.GetBillingDataProfileRequest{
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
		log.Fatalf("Failed to get Billing Data Profile: %v", err)
	}

	for {
		fmt.Println("Receiving Billing Data Profile")
		billingResp, err := billingProfileStream.Recv()
		if err == io.EOF {
			fmt.Println("Billing Data Profile EOF")
			break
		}
		if err != nil {
			if st, ok := status.FromError(err); ok {
				if st.Code() == codes.InvalidArgument && st.Message() == "no meters provided" {
					log.Fatalf("Specific error: no meters provided")
				}
			}
			log.Fatalf("Failed to receive Billing Data Profile: %v", err)
		}
		if billingResp == nil {
			log.Fatalf("Billing Data Profile response is nil")
		}

		fmt.Printf("Received Billing Data Profile from meter %s:\n", billingResp.MeterIp)
		billing := billingResp.Profile
		fmt.Printf("  Billing Date: %s\n", billing.BillingDate)
		fmt.Printf("  Average PF for Billing Period: %.3f\n", billing.AveragePfForBillingPeriod)
		fmt.Printf("  Cumulative Energy Wh Import: %.2f Wh\n", billing.CumEnergyWhImport)
		fmt.Printf("  Cumulative Energy Wh TZ1: %.2f Wh\n", billing.CumEnergyWhTz1)
		fmt.Printf("  Cumulative Energy Wh TZ2: %.2f Wh\n", billing.CumEnergyWhTz2)
		fmt.Printf("  Cumulative Energy Wh TZ3: %.2f Wh\n", billing.CumEnergyWhTz3)
		fmt.Printf("  Cumulative Energy Wh TZ4: %.2f Wh\n", billing.CumEnergyWhTz4)
		fmt.Printf("  Cumulative Energy VAh Import: %.2f VAh\n", billing.CumEnergyVahImport)
		fmt.Printf("  Cumulative Energy VAh TZ1: %.2f VAh\n", billing.CumEnergyVahTz1)
		fmt.Printf("  Cumulative Energy VAh TZ2: %.2f VAh\n", billing.CumEnergyVahTz2)
		fmt.Printf("  Cumulative Energy VAh TZ3: %.2f VAh\n", billing.CumEnergyVahTz3)
		fmt.Printf("  Cumulative Energy VAh TZ4: %.2f VAh\n", billing.CumEnergyVahTz4)
		fmt.Printf("  MD W: %.2f W\n", billing.Mdw)
		fmt.Printf("  MD W DateTime: %s\n", billing.MdwDateTime)
		fmt.Printf("  MD VA: %.2f VA\n", billing.Mdva)
		fmt.Printf("  MD VA DateTime: %s\n", billing.MdvaDateTime)
		fmt.Printf("  Billing Power On Duration: %.2f hours\n", billing.BillingPowerOnDuration)
		fmt.Printf("  Cumulative Energy Wh: %.2f Wh\n", billing.CumEnergyWh)
		fmt.Printf("  Cumulative Energy VAh: %.2f VAh\n", billing.CumEnergyVah)
		fmt.Println("---")
	}

	fmt.Println("Done with Billing Data Profile")

	fmt.Println("Waiting for 5 seconds")
	time.Sleep(5 * time.Second)

	// Test GetInstantaneousProfile
	fmt.Println("\n=== Getting Instantaneous Profile ===")

	instantProfileStream, err := client.GetInstantaneousProfile(context.Background(), &proto.GetInstantaneousProfileRequest{
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
		log.Fatalf("Failed to get Instantaneous Profile: %v", err)
	}

	for {
		fmt.Println("Receiving Instantaneous Profile")
		instantResp, err := instantProfileStream.Recv()
		if err == io.EOF {
			fmt.Println("Instantaneous Profile EOF")
			break
		}
		if err != nil {
			if st, ok := status.FromError(err); ok {
				if st.Code() == codes.InvalidArgument && st.Message() == "no meters provided" {
					log.Fatalf("Specific error: no meters provided")
				}
			}
			log.Fatalf("Failed to receive Instantaneous Profile: %v", err)
		}
		if instantResp == nil {
			log.Fatalf("Instantaneous Profile response is nil")
		}

		fmt.Printf("Received Instantaneous Profile from meter %s:\n", instantResp.MeterIp)
		instant := instantResp.Profile
		fmt.Printf("  DateTime: %s\n", instant.DateTime)
		fmt.Printf("  Voltage: %.2f V\n", instant.Voltage)
		fmt.Printf("  Phase Current: %.2f A\n", instant.PhaseCurrent)
		fmt.Printf("  Neutral Current: %.2f A\n", instant.NeutralCurrent)
		fmt.Printf("  Signed Power Factor: %.3f\n", instant.SignedPowerFactor)
		fmt.Printf("  Frequency: %.2f Hz\n", instant.Frequency)
		fmt.Printf("  Apparent Power: %.2f VA\n", instant.ApparentPower)
		fmt.Printf("  Active Power: %.2f W\n", instant.ActivePower)
		fmt.Printf("  Cumulative Energy Wh: %.2f Wh\n", instant.CumEnergyWh)
		fmt.Println("---")
	}

	fmt.Println("Done with Instantaneous Profile")
}
