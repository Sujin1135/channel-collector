package service

import (
	"channel-collector/internal/channel"
	"channel-collector/internal/collector"
	"fmt"
	"github.com/Sujin1135/channel-collector-interface/protobuf/entity"
	"github.com/Sujin1135/channel-collector-interface/protobuf/service"
	"github.com/pkg/errors"
	"sync"
)

type ChannelService struct {
	service.UnimplementedChannelServiceServer
	collector *collector.Collector
}

func NewChannelService() *ChannelService {
	return &ChannelService{collector: collector.NewCollector()}
}

func (s *ChannelService) GetChannels(request *service.GetChannelsRequest, stream service.ChannelService_GetChannelsServer) error {
	var wg sync.WaitGroup
	wg.Add(1)
	var dataWg sync.WaitGroup
	dataWg.Add(2)
	ch := make(chan *channel.Channel, len(request.YoutubeHandles))
	errCh := make(chan error, len(request.YoutubeHandles))

	go s.collector.Collect(request.YoutubeHandles, ch, errCh, &wg)

	err := s.handleGetChannelsStreamResponses(stream, ch, errCh)
	if err != nil {
		fmt.Println("failed to send a stream message cause as follow:", err.Error())
	}

	wg.Wait()
	dataWg.Wait()
	return nil
}

func (s *ChannelService) handleGetChannelsStreamResponses(
	stream service.ChannelService_GetChannelsServer,
	channelCh <-chan *channel.Channel,
	errCh <-chan error,
) error {
	var streamErr error
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for err := range errCh {
			msg := err.Error()
			if sendErr := s.sendGetChannelsErrorResponse(stream, msg); sendErr != nil {
				streamErr = sendErr
			}
		}
	}()

	go func() {
		defer wg.Done()
		for channelData := range channelCh {
			if sendErr := s.sendGetChannelsDataResponse(stream, channelData); sendErr != nil {
				streamErr = sendErr
			}
		}
	}()

	wg.Wait()
	return streamErr
}

func (s *ChannelService) sendGetChannelsErrorResponse(
	stream service.ChannelService_GetChannelsServer,
	errMsg string,
) error {
	err := stream.Send(&service.GetChannelsResponse{
		Value: &service.GetChannelsResponse_Error_{
			Error: &service.GetChannelsResponse_Error{
				Error: &service.GetChannelsResponse_Error_BadRequest{
					BadRequest: &entity.BadRequestError{Message: &errMsg},
				},
			},
		},
	})

	if err != nil {
		return errors.Wrap(err, "failed to send error response to client")
	}
	return nil
}

func (s *ChannelService) sendGetChannelsDataResponse(
	stream service.ChannelService_GetChannelsServer,
	data *channel.Channel,
) error {
	err := stream.Send(&service.GetChannelsResponse{
		Value: &service.GetChannelsResponse_Data_{
			Data: &service.GetChannelsResponse_Data{
				Channel: &entity.Channel{
					ChannelId:       data.ChannelId,
					Title:           data.Title,
					Description:     data.Description,
					IsFamilySafe:    data.IsFamilySafe,
					Keywords:        data.Keywords,
					Thumbnails:      data.Thumbnails,
					Links:           data.Links,
					ViewCount:       int32(data.ViewCount),
					TotalSubscriber: int32(data.TotalSubscriber),
					TotalVideo:      int32(data.TotalVideo),
					Jointed: &entity.Channel_Joined{
						Year:  int32(data.Joined.Year),
						Month: int32(data.Joined.Month),
						Date:  int32(data.Joined.Date),
					},
				},
			},
		},
	})

	if err != nil {
		return errors.Wrap(err, "failed to send channel data to client")
	}
	return nil
}
