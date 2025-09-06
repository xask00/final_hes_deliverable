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
			meter, err := dlms.NewFakeMeter(reqMeter.Ip, int(reqMeter.Port))
			if err != nil {
				slog.Error("NewFakeMeter", "error", err)
				errChan <- err
			}

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
