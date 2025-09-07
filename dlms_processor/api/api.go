package api

import (
	"dlmsprocessor/dlms"
	"dlmsprocessor/proto"
	"log/slog"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DLMSProcessorAPI struct {
	proto.UnimplementedDLMSProcessorServer
}

func NewDLMSProcessorAPI() *DLMSProcessorAPI {
	return &DLMSProcessorAPI{}
}

func (s *DLMSProcessorAPI) GetOBIS(req *proto.GetOBISRequest, stream grpc.ServerStreamingServer[proto.GetOBISResponse]) error {

	if len(req.Meter) == 0 {
		return status.Error(codes.InvalidArgument, "no meters provided")
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(req.Meter))

	for _, reqMeter := range req.Meter {
		wg.Add(1)
		go func(reqMeter *proto.Meter) {
			defer wg.Done()

			slog.Info("NewFakeMeter", "ip", reqMeter.Ip, "port", reqMeter.Port)
			var meter dlms.Meter
			//meter, err := dlms.NewFakeMeter(reqMeter.Ip, int(reqMeter.Port))
			meter, err := dlms.NewRealMeter(dlms.RealMeter{
				MeterIP:           reqMeter.Ip,
				MeterPort:         int(reqMeter.Port),
				AuthPassword:      reqMeter.AuthPassword,
				SystemTitle:       reqMeter.SystemTitle,
				BlockCipherKey:    reqMeter.BlockCipherKey,
				AuthenticationKey: reqMeter.AuthKey,
			})
			if err != nil {
				slog.Error("NewFakeMeter", "error", err)
				errChan <- err
			}

			slog.Info("Connecting to meter")
			err1 := meter.Connect()

			if err1 != nil {
				slog.Error("Connect", "error", err1)
				errChan <- err1
			}
			slog.Info("Connected to meter")

			obis, err := meter.GetOBIS(reqMeter.Obis)
			if err != nil {
				errChan <- err
				return
			}

			err = stream.Send(&proto.GetOBISResponse{Value: obis})
			if err != nil {
				errChan <- err
				return
			}
		}(reqMeter)
	}

	wg.Wait()

	// Check for any errors
	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

func (s *DLMSProcessorAPI) GetBlockLoadProfile(req *proto.GetBlockLoadProfileRequest, stream grpc.ServerStreamingServer[proto.GetBlockLoadProfileResponse]) error {

	if len(req.Meter) == 0 {
		return status.Error(codes.InvalidArgument, "no meters provided")
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(req.Meter))

	for _, reqMeter := range req.Meter {
		wg.Add(1)
		go func(reqMeter *proto.Meter) {
			defer wg.Done()

			slog.Info("NewRealMeter for BlockLoadProfile", "ip", reqMeter.Ip, "port", reqMeter.Port)
			var meter dlms.Meter
			//meter, err := dlms.NewFakeMeter(reqMeter.Ip, int(reqMeter.Port))
			meter, err := dlms.NewRealMeter(dlms.RealMeter{
				MeterIP:           reqMeter.Ip,
				MeterPort:         int(reqMeter.Port),
				AuthPassword:      reqMeter.AuthPassword,
				SystemTitle:       reqMeter.SystemTitle,
				BlockCipherKey:    reqMeter.BlockCipherKey,
				AuthenticationKey: reqMeter.AuthKey,
			})
			if err != nil {
				slog.Error("NewRealMeter", "error", err)
				errChan <- err
				return
			}

			slog.Info("Connecting to meter for BlockLoadProfile")
			err1 := meter.Connect()

			if err1 != nil {
				slog.Error("Connect", "error", err1)
				errChan <- err1
				return
			}
			slog.Info("Connected to meter for BlockLoadProfile")

			profile, err := meter.GetBlockLoadProfile()
			if err != nil {
				errChan <- err
				return
			}

			// Convert from dlms.BlockLoadProfile to proto.BlockLoadProfile
			protoProfile := &proto.BlockLoadProfile{
				DateTime:             profile.DateTime,
				AverageVoltage:       profile.AverageVoltage,
				BlockEnergyWhImport:  profile.BlockEnergyWhImport,
				BlockEnergyVahImport: profile.BlockEnergyVAhImport,
				BlockEnergyWhExport:  profile.BlockEnergyWhExport,
				BlockEnergyVahExport: profile.BlockEnergyVAhExport,
				AverageCurrent:       profile.AverageCurrent,
				MeterHealthIndicator: uint32(profile.MeterHealthIndicator),
			}

			err = stream.Send(&proto.GetBlockLoadProfileResponse{
				Profile: protoProfile,
				MeterIp: reqMeter.Ip,
			})
			if err != nil {
				errChan <- err
				return
			}
		}(reqMeter)
	}

	wg.Wait()

	// Check for any errors
	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}
