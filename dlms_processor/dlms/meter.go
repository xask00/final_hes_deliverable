package dlms

import (
	"errors"
	"log/slog"
)

type Meter interface {
	Connect() error
	GetOBIS(obis string) (string, error)
	GetBlockLoadProfile() (*BlockLoadProfile, error)
	GetDailyLoadProfile() (*DailyLoadProfile, error)
	GetBillingDataProfile() (*BillingDataProfile, error)
	GetInstantaneousProfile() (*InstantaneousProfile, error)
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

func (m *FakeMeter) GetDailyLoadProfile() (*DailyLoadProfile, error) {
	// Return fake data for testing
	return &DailyLoadProfile{
		DateTime:                  "2024-01-15 00:00:00",
		CumulativeEnergyWhExport:  1500.25,
		CumulativeEnergyVAhExport: 1600.75,
		CumulativeEnergyWhImport:  12500.50,
		CumulativeEnergyVAhImport: 13000.25,
	}, nil
}

func (m *FakeMeter) GetBillingDataProfile() (*BillingDataProfile, error) {
	// Return fake data for testing
	return &BillingDataProfile{
		BillingDate:               "2024-01-01",
		AveragePFForBillingPeriod: 0.95,
		CumEnergyWhImport:         15000.75,
		CumEnergyWhTZ1:            3000.25,
		CumEnergyWhTZ2:            4000.50,
		CumEnergyWhTZ3:            4500.75,
		CumEnergyWhTZ4:            3500.25,
		CumEnergyVAhImport:        16000.50,
		CumEnergyVAhTZ1:           3200.25,
		CumEnergyVAhTZ2:           4200.75,
		CumEnergyVAhTZ3:           4700.50,
		CumEnergyVAhTZ4:           3900.25,
		MDW:                       5500.75,
		MDWDateTime:               "2024-01-15 14:30:00",
		MDVA:                      5800.25,
		MDVADateTime:              "2024-01-15 14:35:00",
		BillingPowerOnDuration:    720.5,
		CumEnergyWh:               1200.75,
		CumEnergyVAh:              1250.50,
	}, nil
}

func (m *FakeMeter) GetInstantaneousProfile() (*InstantaneousProfile, error) {
	// Return fake data for testing
	return &InstantaneousProfile{
		DateTime:          "2024-01-15 12:30:00",
		Voltage:           230.5,
		PhaseCurrent:      5.25,
		NeutralCurrent:    0.15,
		SignedPowerFactor: 0.98,
		Frequency:         50.02,
		ApparentPower:     1210.75,
		ActivePower:       1186.50,
		CumEnergyWh:       12500.75,
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
