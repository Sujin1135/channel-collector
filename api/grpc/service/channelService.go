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

	go func() {
		defer dataWg.Done()

		for err := range errCh {
			msg := err.Error()
			err := stream.Send(&service.GetChannelsResponse{
				Value: &service.GetChannelsResponse_Error_{Error: &service.GetChannelsResponse_Error{
					Error: &service.GetChannelsResponse_Error_BadRequest{
						BadRequest: &entity.BadRequestError{Message: &msg},
					},
				}},
			})
			if err != nil {
				fmt.Println(errors.Wrap(err, "occurred an error when send a error message to client"))
			}
		}
	}()

	go func() {
		defer dataWg.Done()

		for data := range ch {
			err := stream.Send(&service.GetChannelsResponse{
				Value: &service.GetChannelsResponse_Data_{Data: &service.GetChannelsResponse_Data{
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
				}},
			})
			if err != nil {
				fmt.Println(errors.Wrap(err, "occurred an error when send a channel message to client"))
			}
		}
	}()

	wg.Wait()
	dataWg.Wait()
	return nil
}
