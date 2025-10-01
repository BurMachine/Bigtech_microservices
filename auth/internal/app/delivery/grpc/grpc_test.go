package auth_grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/models"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/usecases/auth/dto"
	"github.com/BurMachine/Bigtech_microservices/auth/internal/app/usecases/auth/mocks" // Импорт моков usecases
	pb "github.com/BurMachine/Bigtech_microservices/auth/pkg/v1/auth"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestService_Login(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockUsecase := mocks.NewMockAuthUsecases(mockCtrl)

	// Создаём Service с мокнутым usecase. Валидатор инициализируется реально (как в New).
	s, err := New(mockUsecase)
	if err != nil {
		t.Fatalf("failed to create Service: %v", err)
	}

	request := &pb.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	t.Run("Success", func(t *testing.T) {
		expectedToken := &models.UserToken{
			ID:           "user1",
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
		}
		mockUsecase.EXPECT().
			Login(ctx, dto.LoginDTO{Email: request.Email, Password: request.Password}).
			Return(expectedToken, nil).
			Times(1)

		resp, err := s.Login(ctx, request)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if resp.UserId != expectedToken.ID ||
			resp.AccessToken != expectedToken.AccessToken ||
			resp.RefreshToken != expectedToken.RefreshToken {
			t.Errorf("expected response %+v, got %+v", expectedToken, resp)
		}
	})

	t.Run("UsecaseError", func(t *testing.T) {
		expectedErr := errors.New("invalid credentials")
		mockUsecase.EXPECT().
			Login(ctx, dto.LoginDTO{Email: request.Email, Password: request.Password}).
			Return(nil, expectedErr).
			Times(1)

		_, err := s.Login(ctx, request)
		if err == nil {
			t.Error("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.Internal {
			t.Errorf("expected codes.Internal, got %v", st.Code())
		}
		if st.Message() != expectedErr.Error() {
			t.Errorf("expected message %q, got %q", expectedErr.Error(), st.Message())
		}
	})
}

func TestService_Register(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockUsecase := mocks.NewMockAuthUsecases(mockCtrl)

	s, err := New(mockUsecase)
	if err != nil {
		t.Fatalf("failed to create Service: %v", err)
	}

	request := &pb.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	t.Run("Success", func(t *testing.T) {
		expectedUser := &models.User{
			ID: "user1",
		}
		mockUsecase.EXPECT().
			Register(ctx, dto.RegisterDTO{Email: request.Email, Password: request.Password}).
			Return(expectedUser, nil).
			Times(1)

		resp, err := s.Register(ctx, request)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if resp.UserId != expectedUser.ID {
			t.Errorf("expected UserId %q, got %q", expectedUser.ID, resp.UserId)
		}
	})

	t.Run("UsecaseError", func(t *testing.T) {
		expectedErr := errors.New("user already exists")
		mockUsecase.EXPECT().
			Register(ctx, dto.RegisterDTO{Email: request.Email, Password: request.Password}).
			Return(nil, expectedErr).
			Times(1)

		_, err := s.Register(ctx, request)
		if err == nil {
			t.Error("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.Internal {
			t.Errorf("expected codes.Internal, got %v", st.Code())
		}
		if st.Message() != expectedErr.Error() {
			t.Errorf("expected message %q, got %q", expectedErr.Error(), st.Message())
		}
	})
}

func TestService_Refresh(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockUsecase := mocks.NewMockAuthUsecases(mockCtrl)

	s, err := New(mockUsecase)
	if err != nil {
		t.Fatalf("failed to create Service: %v", err)
	}

	request := &pb.RefreshRequest{
		RefreshToken: "valid_refresh_token",
	}

	t.Run("Success", func(t *testing.T) {
		expectedToken := &models.UserToken{
			ID:           "user1",
			AccessToken:  "new_access_token",
			RefreshToken: "new_refresh_token",
		}
		mockUsecase.EXPECT().
			Refresh(ctx, request.RefreshToken).
			Return(expectedToken, nil).
			Times(1)

		resp, err := s.Refresh(ctx, request)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if resp.UserId != expectedToken.ID ||
			resp.AccessToken != expectedToken.AccessToken ||
			resp.RefreshToken != expectedToken.RefreshToken {
			t.Errorf("expected response %+v, got %+v", expectedToken, resp)
		}
	})

	t.Run("UsecaseError", func(t *testing.T) {
		expectedErr := errors.New("invalid refresh token")
		mockUsecase.EXPECT().
			Refresh(ctx, request.RefreshToken).
			Return(nil, expectedErr).
			Times(1)

		_, err := s.Refresh(ctx, request)
		if err == nil {
			t.Error("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.Internal {
			t.Errorf("expected codes.Internal, got %v", st.Code())
		}
		if st.Message() != expectedErr.Error() {
			t.Errorf("expected message %q, got %q", expectedErr.Error(), st.Message())
		}
	})
}
