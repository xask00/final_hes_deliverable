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
	defer m.client.Close()
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
	defer m.client.Close()
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

func (m *RealMeter) GetDailyLoadProfile() (*DailyLoadProfile, error) {
	if m.client == nil {
		slog.Error("client not initialized")
		return nil, fmt.Errorf("client not initialized")
	}

	err := m.client.Connect()
	defer m.client.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to meter: %w", err)
	}

	results, err := ReadProfileDataTyped[DailyLoadProfile](m.client, "1.0.99.2.0.255", 1, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to read daily load profile data: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no daily load profile data found")
	}

	slog.Info("daily load profile results", "results", results[0])

	return &results[0], nil
}

func (m *RealMeter) GetBillingDataProfile() (*BillingDataProfile, error) {
	if m.client == nil {
		slog.Error("client not initialized")
		return nil, fmt.Errorf("client not initialized")
	}

	err := m.client.Connect()
	defer m.client.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to meter: %w", err)
	}

	results, err := ReadProfileDataTyped[BillingDataProfile](m.client, "0.0.98.1.0.255", 1, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to read billing data profile: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no billing data profile found")
	}

	slog.Info("billing data profile results", "results", results[0])

	return &results[0], nil
}

func (m *RealMeter) GetInstantaneousProfile() (*InstantaneousProfile, error) {
	if m.client == nil {
		slog.Error("client not initialized")
		return nil, fmt.Errorf("client not initialized")
	}

	err := m.client.Connect()
	defer m.client.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to meter: %w", err)
	}

	results, err := ReadProfileDataTyped[InstantaneousProfile](m.client, "1.0.94.7.0.255", 1, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to read instantaneous profile: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no instantaneous profile data found")
	}

	slog.Info("instantaneous profile results", "results", results[0])

	return &results[0], nil
}
