package service

import (
	"context"

	urlpb "github.com/AlexeySalamakhin/URLShortener/internal/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	urlpb.UnimplementedURLShortenerServer
	logic *URLShortenerService
}

func NewGRPCServer(logic *URLShortenerService) *GRPCServer {
	return &GRPCServer{logic: logic}
}

func (s *GRPCServer) CreateShortURL(ctx context.Context, req *urlpb.CreateShortURLRequest) (*urlpb.CreateShortURLResponse, error) {
	shortKey, conflict := s.logic.Shorten(ctx, req.OriginalUrl, req.UserId)
	return &urlpb.CreateShortURLResponse{
		ShortUrl: shortKey,
		Conflict: conflict,
	}, nil
}

func (s *GRPCServer) GetOriginalURL(ctx context.Context, req *urlpb.GetOriginalURLRequest) (*urlpb.GetOriginalURLResponse, error) {
	record, found := s.logic.GetOriginalURL(ctx, req.ShortUrl)
	return &urlpb.GetOriginalURLResponse{
		OriginalUrl: record.OriginalURL,
		IsDeleted:   record.DeletedFlag,
		Found:       found,
	}, nil
}

func (s *GRPCServer) BatchShorten(ctx context.Context, req *urlpb.BatchShortenRequest) (*urlpb.BatchShortenResponse, error) {
	var resp urlpb.BatchShortenResponse
	for _, item := range req.Items {
		shortKey, _ := s.logic.Shorten(ctx, item.OriginalUrl, req.UserId)
		resp.Items = append(resp.Items, &urlpb.BatchShortenResponseItem{
			CorrelationId: item.CorrelationId,
			ShortUrl:      shortKey,
		})
	}
	return &resp, nil
}

func (s *GRPCServer) GetUserURLs(ctx context.Context, req *urlpb.GetUserURLsRequest) (*urlpb.GetUserURLsResponse, error) {
	urls, err := s.logic.GetUserURLs(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user urls: %v", err)
	}
	var resp urlpb.GetUserURLsResponse
	for _, u := range urls {
		resp.Urls = append(resp.Urls, &urlpb.UserURL{
			ShortUrl:    u.ShortURL,
			OriginalUrl: u.OriginalURL,
			IsDeleted:   u.DeletedFlag,
		})
	}
	return &resp, nil
}

func (s *GRPCServer) DeleteUserURLs(ctx context.Context, req *urlpb.DeleteUserURLsRequest) (*urlpb.DeleteUserURLsResponse, error) {
	err := s.logic.DeleteUserURLs(ctx, req.UserId, req.Ids)
	if err != nil {
		return &urlpb.DeleteUserURLsResponse{Success: false}, status.Errorf(codes.Internal, "failed to delete user urls: %v", err)
	}
	return &urlpb.DeleteUserURLsResponse{Success: true}, nil
}

func (s *GRPCServer) GetStats(ctx context.Context, req *urlpb.GetStatsRequest) (*urlpb.GetStatsResponse, error) {
	urls, users, err := s.logic.GetStats(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get stats: %v", err)
	}
	return &urlpb.GetStatsResponse{
		Urls:  int32(urls),
		Users: int32(users),
	}, nil
}

func (s *GRPCServer) Ping(ctx context.Context, req *urlpb.PingRequest) (*urlpb.PingResponse, error) {
	return &urlpb.PingResponse{Ready: s.logic.StoreReady()}, nil
}
