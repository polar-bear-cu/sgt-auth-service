package clients

import (
	"context"

	userv1 "github.com/polar-bear-cu/sgt-proto/gen/go/user/v1"
	"google.golang.org/grpc"
)

type UserClient struct {
	rpc userv1.UserServiceClient
}

func NewUserClient(conn *grpc.ClientConn) *UserClient {
	return &UserClient{rpc: userv1.NewUserServiceClient(conn)}
}

func (u *UserClient) GetUser(ctx context.Context, id string) (email, name, pictureURL string, err error) {
	resp, err := u.rpc.GetUser(ctx, &userv1.GetUserRequest{Id: id})
	if err != nil {
		return "", "", "", err
	}
	user := resp.GetUser()
	return user.GetEmail(), user.GetName(), user.GetPictureUrl(), nil
}

func (u *UserClient) FindOrCreateUser(ctx context.Context, email, googleSub, name, pictureURL string) (string, error) {
	resp, err := u.rpc.FindOrCreateUser(ctx, &userv1.FindOrCreateUserRequest{
		Email:      email,
		GoogleSub:  googleSub,
		Name:       name,
		PictureUrl: pictureURL,
	})
	if err != nil {
		return "", err
	}
	return resp.GetUser().GetId(), nil
}
