package api

import (
	"context"
	"dlmsprocessor/proto"
	"io"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	proto.RegisterDLMSProcessorServer(s, NewDLMSProcessorAPI())
	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func getTestClient(ctx context.Context) (proto.DLMSProcessorClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	client := proto.NewDLMSProcessorClient(conn)
	return client, conn, nil
}

func TestGetOBIS_SingleMeter_Success(t *testing.T) {
	ctx := context.Background()
	client, conn, err := getTestClient(ctx)
	if err != nil {
		t.Fatalf("Failed to get test client: %v", err)
	}
	defer conn.Close()

	// Test with a single meter
	req := &proto.GetOBISRequest{
		Meter: []*proto.Meter{
			{
				Ip:   "192.168.1.100",
				Port: 4059,
				Obis: "1.0.1.8.0.255",
			},
		},
		Obis: "1.0.1.8.0.255",
	}

	stream, err := client.GetOBIS(ctx, req)
	if err != nil {
		t.Fatalf("GetOBIS failed: %v", err)
	}

	// Collect responses
	var responses []*proto.GetOBISResponse
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Failed to receive response: %v", err)
		}
		responses = append(responses, resp)
	}

	// Verify we got at least one response
	if len(responses) == 0 {
		t.Fatal("Expected at least one response, got none")
	}

	// Verify the response contains the expected OBIS value
	if responses[0].Value != "1.0.1.8.0.255" {
		t.Errorf("Expected OBIS value '1.0.1.8.0.255', got '%s'", responses[0].Value)
	}
}

func TestGetOBIS_MultipleMeter_Success(t *testing.T) {
	ctx := context.Background()
	client, conn, err := getTestClient(ctx)
	if err != nil {
		t.Fatalf("Failed to get test client: %v", err)
	}
	defer conn.Close()

	// Test with multiple meters
	req := &proto.GetOBISRequest{
		Meter: []*proto.Meter{
			{
				Ip:   "192.168.1.100",
				Port: 4059,
				Obis: "1.0.1.8.0.255",
			},
			{
				Ip:   "192.168.1.101",
				Port: 4059,
				Obis: "1.0.2.8.0.255",
			},
			{
				Ip:   "192.168.1.102",
				Port: 4060,
				Obis: "1.0.3.8.0.255",
			},
		},
		Obis: "1.0.1.8.0.255",
	}

	stream, err := client.GetOBIS(ctx, req)
	if err != nil {
		t.Fatalf("GetOBIS failed: %v", err)
	}

	// Collect responses
	var responses []*proto.GetOBISResponse
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Failed to receive response: %v", err)
		}
		responses = append(responses, resp)
	}

	// Verify we got responses for all meters
	if len(responses) != 3 {
		t.Errorf("Expected 3 responses, got %d", len(responses))
	}

	// Verify each response contains the correct OBIS values
	expectedObis := []string{"1.0.1.8.0.255", "1.0.2.8.0.255", "1.0.3.8.0.255"}
	receivedObis := make(map[string]bool)

	for _, resp := range responses {
		receivedObis[resp.Value] = true
	}

	for _, expected := range expectedObis {
		if !receivedObis[expected] {
			t.Errorf("Expected OBIS value '%s' not found in responses", expected)
		}
	}
}

func TestGetOBIS_EmptyMeterList(t *testing.T) {
	ctx := context.Background()
	client, conn, err := getTestClient(ctx)
	if err != nil {
		t.Fatalf("Failed to get test client: %v", err)
	}
	defer conn.Close()

	// Test with empty meter list
	req := &proto.GetOBISRequest{
		Meter: []*proto.Meter{},
		Obis:  "1.0.1.8.0.255",
	}

	stream, err := client.GetOBIS(ctx, req)
	if err != nil {
		t.Fatalf("GetOBIS failed: %v", err)
	}

	// Collect responses
	var responses []*proto.GetOBISResponse
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Failed to receive response: %v", err)
		}
		responses = append(responses, resp)
	}

	// Should receive no responses for empty meter list
	if len(responses) != 0 {
		t.Errorf("Expected 0 responses for empty meter list, got %d", len(responses))
	}
}

func TestGetOBIS_DifferentOBISCodes(t *testing.T) {
	ctx := context.Background()
	client, conn, err := getTestClient(ctx)
	if err != nil {
		t.Fatalf("Failed to get test client: %v", err)
	}
	defer conn.Close()

	testCases := []struct {
		name         string
		obisCode     string
		meterObis    string
		expectedObis string
	}{
		{
			name:         "Energy Register",
			obisCode:     "1.0.1.8.0.255",
			meterObis:    "1.0.1.8.0.255",
			expectedObis: "1.0.1.8.0.255",
		},
		{
			name:         "Voltage",
			obisCode:     "1.0.32.7.0.255",
			meterObis:    "1.0.32.7.0.255",
			expectedObis: "1.0.32.7.0.255",
		},
		{
			name:         "Current",
			obisCode:     "1.0.31.7.0.255",
			meterObis:    "1.0.31.7.0.255",
			expectedObis: "1.0.31.7.0.255",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &proto.GetOBISRequest{
				Meter: []*proto.Meter{
					{
						Ip:   "192.168.1.100",
						Port: 4059,
						Obis: tc.meterObis,
					},
				},
				Obis: tc.obisCode,
			}

			stream, err := client.GetOBIS(ctx, req)
			if err != nil {
				t.Fatalf("GetOBIS failed for %s: %v", tc.name, err)
			}

			// Get first response
			resp, err := stream.Recv()
			if err != nil {
				t.Fatalf("Failed to receive response for %s: %v", tc.name, err)
			}

			if resp.Value != tc.expectedObis {
				t.Errorf("For %s: expected OBIS '%s', got '%s'", tc.name, tc.expectedObis, resp.Value)
			}

			// Close the stream
			for {
				_, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					break
				}
			}
		})
	}
}

func TestGetOBIS_ConcurrentRequests(t *testing.T) {
	ctx := context.Background()
	client, conn, err := getTestClient(ctx)
	if err != nil {
		t.Fatalf("Failed to get test client: %v", err)
	}
	defer conn.Close()

	// Test concurrent requests
	numConcurrent := 5
	done := make(chan bool, numConcurrent)
	errors := make(chan error, numConcurrent)

	for i := 0; i < numConcurrent; i++ {
		go func(id int) {
			req := &proto.GetOBISRequest{
				Meter: []*proto.Meter{
					{
						Ip:   "192.168.1.100",
						Port: int32(4059 + id),
						Obis: "1.0.1.8.0.255",
					},
				},
				Obis: "1.0.1.8.0.255",
			}

			stream, err := client.GetOBIS(ctx, req)
			if err != nil {
				errors <- err
				return
			}

			// Consume all responses
			responseCount := 0
			for {
				_, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					errors <- err
					return
				}
				responseCount++
			}

			if responseCount != 1 {
				errors <- err
				return
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	completedCount := 0
	errorCount := 0
	for i := 0; i < numConcurrent; i++ {
		select {
		case <-done:
			completedCount++
		case err := <-errors:
			t.Errorf("Concurrent request failed: %v", err)
			errorCount++
		}
	}

	if completedCount != numConcurrent {
		t.Errorf("Expected %d successful concurrent requests, got %d (errors: %d)",
			numConcurrent, completedCount, errorCount)
	}
}
