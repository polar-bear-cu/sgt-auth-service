package grpc

import (
	"context"

	authv1 "github.com/polar-bear-cu/sgt-proto/gen/go/auth/v1"

	"github.com/polar-bear-cu/sgt-auth-service/usecases"
)

type AuthServer struct {
	authv1.UnimplementedAuthServiceServer
	uc *usecases.AuthUsecase
}

func NewAuthServer(uc *usecases.AuthUsecase) *AuthServer {
	return &AuthServer{uc: uc}
}

func (s *AuthServer) DeleteExpiredRefreshTokens(
	ctx context.Context,
	_ *authv1.DeleteExpiredRefreshTokensRequest,
) (*authv1.DeleteExpiredRefreshTokensResponse, error) {
	n, err := s.uc.CleanupRefreshTokens(ctx)
	if err != nil {
		return nil, err
	}
	return &authv1.DeleteExpiredRefreshTokensResponse{DeletedCount: n}, nil
}
