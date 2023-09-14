package kyc

import (
	"context"
	"fmt"

	appcli "github.com/NpoolPlatform/appuser-middleware/pkg/client/app"
	appusercli "github.com/NpoolPlatform/appuser-middleware/pkg/client/user"
	"github.com/NpoolPlatform/libent-cruder/pkg/cruder"
	"github.com/NpoolPlatform/message/npool/appuser/mw/v1/user"
	reviewtypes "github.com/NpoolPlatform/message/npool/basetypes/review/v1"
	basetyeps "github.com/NpoolPlatform/message/npool/basetypes/v1"
	constant "github.com/NpoolPlatform/review-gateway/pkg/const"
	"github.com/google/uuid"
)

type Handler struct {
	AppID       *string
	TargetAppID *string
	UserID      *string
	ReviewID    *string
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

func WithTargetAppID(targetAppID *string) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		_, err := uuid.Parse(*targetAppID)
		if err != nil {
			return err
		}
		exist, err := appcli.ExistApp(ctx, *targetAppID)
		if err != nil {
			return err
		}
		if !exist {
			return fmt.Errorf("invalid target app")
		}
		h.TargetAppID = targetAppID
		return nil
	}
}

func WithUserID(appID, userID *string) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		_, err := uuid.Parse(*userID)
		if err != nil {
			return err
		}
		exist, err := appusercli.ExistUserConds(ctx, &user.Conds{
			AppID: &basetyeps.StringVal{Op: cruder.EQ, Value: *appID},
			ID:    &basetyeps.StringVal{Op: cruder.EQ, Value: *userID},
		})
		if err != nil {
			return err
		}

		if !exist {
			return fmt.Errorf("invalid user id")
		}
		h.UserID = userID
		return nil
	}
}

func WithReviewID(id *string) func(context.Context, *Handler) error {
	return func(ctx context.Context, h *Handler) error {
		if _, err := uuid.Parse(*id); err != nil {
			return err
		}
		h.ReviewID = id
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
