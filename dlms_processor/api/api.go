package api

import (
	"dlmsprocessor/dlms"
	"dlmsprocessor/proto"

	"google.golang.org/grpc"
)

type DLMSProcessorAPI struct {
	proto.UnimplementedDLMSProcessorServer
}

func NewDLMSProcessorAPI() *DLMSProcessorAPI {
	return &DLMSProcessorAPI{}
}

func (s *DLMSProcessorAPI) GetOBIS(req *proto.GetOBISRequest, stream grpc.ServerStreamingServer[proto.GetOBISResponse]) error {
	for _, reqMeter := range req.Meter {
		go func(reqMeter *proto.Meter) error {
			meter, err := dlms.NewFakeMeter(reqMeter.Ip, int(reqMeter.Port))
			if err != nil {
				return err
			}
			_, err = meter.GetOBIS(req.Obis)
			if err != nil {
				return err
			}
			stream.Send(&proto.GetOBISResponse{Value: reqMeter.Obis})
			return nil
		}(reqMeter)
	}
	return nil
}
