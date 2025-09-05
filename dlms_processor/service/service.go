package service

type Meter interface {
	GetOBIS(obis string) (string, error)
	SetClock(clock string) error
	ExecuteFunction(function string, params []string) (string, error)
	FOTA() error
}
