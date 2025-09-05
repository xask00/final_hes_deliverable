package api

import (
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
	return nil
}
