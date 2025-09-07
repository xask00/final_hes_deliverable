package dlms

import (
	"fmt"
	"log/slog"
)

// Meter represents the configuration for connecting to a DLMS energy meter
type RealMeter struct {
	MeterIP           string
	MeterPort         int
	ConnectionTimeout int
	AuthPassword      string
	SystemTitle       string
	BlockCipherKey    string
	AuthenticationKey string
	ClientAddress     int
	ServerAddress     int
	AttributeIndex    int
	MaxEntries        int

	client *MeterClient
}

func NewRealMeter(meter RealMeter) (*RealMeter, error) {
	return &meter, nil
}

func (m *RealMeter) SetClock(clock string) error {
	return nil
}

func (m *RealMeter) ExecuteFunction(function string, params []string) (string, error) {
	return "123", nil
}

func (m *RealMeter) FOTA() error {
	return nil
}

func (m *RealMeter) Connect() error {
	m.client = NewMeterClient()
	if m.client == nil {
		fmt.Println("Failed to create meter client")
		return fmt.Errorf("failed to create meter client")
	}

	if err := m.client.Configure(m); err != nil {
		fmt.Printf("Failed to configure meter: %v\n", err)
		return fmt.Errorf("failed to configure meter: %w", err)
	}

	if m.client != nil {
		fmt.Println("Connect() method - Client is not nil")
	}

	return nil
}

func (m *RealMeter) GetOBIS(obis string) (string, error) {
	if m.client == nil {
		slog.Error("client not initialized")
		return "", fmt.Errorf("client not initialized")
	}

	err := m.client.Connect()
	if err != nil {
		return "", fmt.Errorf("failed to connect to meter: %w", err)
	}

	return obis, nil
}

func (m *RealMeter) GetBlockLoadProfile() (*BlockLoadProfile, error) {
	if m.client == nil {
		slog.Error("client not initialized")
		return nil, fmt.Errorf("client not initialized")
	}

	err := m.client.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to meter: %w", err)
	}

	results, err := ReadProfileDataTyped[BlockLoadProfile](m.client, "1.0.99.1.0.255", 1, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile data: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no data found")
	}

	slog.Info("block load profile results", "results", results[0])

	return &results[0], nil
}
