package withdraw

import (
	"context"
	"fmt"

	appcli "github.com/NpoolPlatform/appuser-middleware/pkg/client/app"
	usermwcli "github.com/NpoolPlatform/appuser-middleware/pkg/client/user"
	reviewtypes "github.com/NpoolPlatform/message/npool/basetypes/review/v1"
	constant "github.com/NpoolPlatform/review-gateway/pkg/const"
	"github.com/google/uuid"
)

type Handler struct {
	AppID       *string
	TargetAppID *string
	UserID      *string
	ReviewID    *uuid.UUID
	LangID      *string
	Domain      *string
	State       *reviewtypes.ReviewState
	Message     *string
	Offset      int32
	Limit       int32
}

func NewHandler(ctx context.Context, options ...func(context.Context, *Handler) error) (*Handler, error) {
	handler := &Handler{}
	for _, opt := range options {
		if err := opt(ctx, handler); err != nil {
			return nil, err
		}
	}
	return handler, nil
}

func WithAppID(appID *string) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		_, err := uuid.Parse(*appID)
		if err != nil {
			return err
		}
		exist, err := appcli.ExistApp(ctx, *appID)
		if err != nil {
			return err
		}
		if !exist {
			return fmt.Errorf("invalid app")
		}

		h.AppID = appID
		return nil
	}
}

func WithTargetAppID(appID *string) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		_, err := uuid.Parse(*appID)
		if err != nil {
			return err
		}
		exist, err := appcli.ExistApp(ctx, *appID)
		if err != nil {
			return err
		}
		if !exist {
			return fmt.Errorf("invalid app")
		}
		h.TargetAppID = appID
		return nil
	}
}

func WithUserID(appID, userID *string) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		_, err := uuid.Parse(*userID)
		if err != nil {
			return err
		}
		user, err := usermwcli.GetUser(ctx, *appID, *userID)
		if err != nil {
			return err
		}
		if user == nil {
			return fmt.Errorf("user not found")
		}
		h.UserID = userID
		return nil
	}
}

func WithReviewID(id *string) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		_id, err := uuid.Parse(*id)
		if err != nil {
			return err
		}
		h.ReviewID = &_id
		return nil
	}
}

func WithState(state *reviewtypes.ReviewState, message *string) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		if state == nil {
			return nil
		}
		switch *state {
		case reviewtypes.ReviewState_Rejected:
		case reviewtypes.ReviewState_Approved:
		default:
			return fmt.Errorf("invalid review state")
		}
		if *state == reviewtypes.ReviewState_Rejected && message == nil {
			return fmt.Errorf("message is empty")
		}
		h.State = state
		return nil
	}
}

func WithMessage(message *string) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		if message == nil {
			return nil
		}
		h.Message = message
		return nil
	}
}

func WithOffset(offset int32) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		h.Offset = offset
		return nil
	}
}

func WithLimit(limit int32) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		if limit == 0 {
			limit = constant.DefaultRowLimit
		}
		h.Limit = limit
		return nil
	}
}
