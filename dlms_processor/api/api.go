package api

import (
	"dlmsprocessor/dlms"
	"dlmsprocessor/proto"
	"sync"

	"google.golang.org/grpc"
)

type DLMSProcessorAPI struct {
	proto.UnimplementedDLMSProcessorServer
}

func NewDLMSProcessorAPI() *DLMSProcessorAPI {
	return &DLMSProcessorAPI{}
}

func (s *DLMSProcessorAPI) GetOBIS(req *proto.GetOBISRequest, stream grpc.ServerStreamingServer[proto.GetOBISResponse]) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(req.Meter))

	for _, reqMeter := range req.Meter {
		wg.Add(1)
		go func(reqMeter *proto.Meter) {
			defer wg.Done()

			meter, err := dlms.NewFakeMeter(reqMeter.Ip, int(reqMeter.Port))
			if err != nil {
				errChan <- err
				return
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
