package dlms

import (
	"errors"
	"log/slog"
)

type Meter interface {
	Connect() error
	GetOBIS(obis string) (string, error)
	GetBlockLoadProfile() (*BlockLoadProfile, error)
	SetClock(clock string) error
	ExecuteFunction(function string, params []string) (string, error)
	FOTA() error
}

type FakeMeter struct {
	ipv6 string
	port int
}

func NewFakeMeter(ipv6 string, port int) (*FakeMeter, error) {
	slog.Info("NewFakeMeter", "ipv6", ipv6, "port", port)
	if ipv6 == "" {
		slog.Error("ipv6 is required")
		return nil, errors.New("ipv6 is required")
	}

	if port == 0 {
		slog.Error("port is required")
		return nil, errors.New("port is required")
	}

	return &FakeMeter{
		ipv6: ipv6,
		port: port,
	}, nil
}

func (m *FakeMeter) Connect() error {
	return nil
}

func (m *FakeMeter) GetOBIS(obis string) (string, error) {
	return obis, nil
}

func (m *FakeMeter) GetBlockLoadProfile() (*BlockLoadProfile, error) {
	// Return fake data for testing
	return &BlockLoadProfile{
		DateTime:             "2024-01-15 12:00:00",
		AverageVoltage:       230.5,
		BlockEnergyWhImport:  1250.75,
		BlockEnergyVAhImport: 1300.25,
		BlockEnergyWhExport:  50.25,
		BlockEnergyVAhExport: 55.75,
		AverageCurrent:       5.45,
		MeterHealthIndicator: 1,
	}, nil
}

func (m *FakeMeter) SetClock(clock string) error {
	return nil
}

func (m *FakeMeter) ExecuteFunction(function string, params []string) (string, error) {
	return "123", nil
}

func (m *FakeMeter) FOTA() error {
	return nil
}
