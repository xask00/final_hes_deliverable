package dlms

type Meter interface {
	GetOBIS(obis string) (string, error)
	SetClock(clock string) error
	ExecuteFunction(function string, params []string) (string, error)
	FOTA() error
}

type FakeMeter struct {
	ipv6 string
	port int
}

func NewFakeMeter(ipv6 string, port int) (*FakeMeter, error) {
	return &FakeMeter{
		ipv6: ipv6,
		port: port,
	}, nil
}

func (m *FakeMeter) GetOBIS(obis string) (string, error) {
	return obis, nil
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
