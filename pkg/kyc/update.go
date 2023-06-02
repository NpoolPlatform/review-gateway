package kyc

import (
	"context"
	"fmt"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	basetypes "github.com/NpoolPlatform/message/npool/basetypes/v1"

	notifmwpb "github.com/NpoolPlatform/message/npool/notif/mw/v1/notif"
	tmplmwpb "github.com/NpoolPlatform/message/npool/notif/mw/v1/template"
	notifmwcli "github.com/NpoolPlatform/notif-middleware/pkg/client/notif"

	kycmwcli "github.com/NpoolPlatform/appuser-middleware/pkg/client/kyc"
	usercli "github.com/NpoolPlatform/appuser-middleware/pkg/client/user"
	kycmwpb "github.com/NpoolPlatform/message/npool/appuser/mw/v1/kyc"
	review1 "github.com/NpoolPlatform/review-gateway/pkg/review"

	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/kyc"
	reviewmgrpb "github.com/NpoolPlatform/message/npool/review/mw/v2"
)

func (h *Handler) UpdateKycReview(ctx context.Context) (*npool.KycReview, error) {
	reviewID := h.ReviewID.String()
	handler, err := review1.NewHandler(
		ctx,
		review1.WithAppID(h.AppID),
		review1.WithUserID(h.UserID),
		review1.WithReviewID(&reviewID),
		review1.WithState(h.State, nil),
		review1.WithMessage(h.Message),
	)
	if err != nil {
		return nil, err
	}

	objID, err := handler.ValidateReview(ctx)
	if err != nil {
		return nil, err
	}

	kycInfo, err := kycmwcli.GetKyc(ctx, objID)
	if err != nil {
		return nil, err
	}
	if kycInfo == nil {
		return nil, fmt.Errorf("invalid kyc")
	}

	userInfo, err := usercli.GetUser(ctx, kycInfo.AppID, kycInfo.UserID)
	if err != nil {
		return nil, err
	}
	if userInfo == nil {
		return nil, fmt.Errorf("invalid user")
	}

	if err := handler.UpdateReview(ctx); err != nil {
		return nil, err
	}

	eventType := basetypes.UsedFor_KYCApproved
	kycState := basetypes.KycState_Approved
	if *h.State == reviewmgrpb.ReviewState_Rejected {
		kycState = basetypes.KycState_Rejected
		eventType = basetypes.UsedFor_KYCRejected
	}

	_, err = kycmwcli.UpdateKyc(ctx, &kycmwpb.KycReq{
		ID:    &objID,
		State: &kycState,
	})
	if err != nil {
		return nil, err
	}

	_, err = notifmwcli.GenerateNotifs(ctx, &notifmwpb.GenerateNotifsRequest{
		AppID:     *h.AppID,
		UserID:    kycInfo.UserID,
		EventType: eventType,
		Vars: &tmplmwpb.TemplateVars{
			Username: &userInfo.Username,
		},
	})
	if err != nil {
		logger.Sugar().Errorw("UpdateKycReview", "Error", err)
	}

	return h.GetKycReview(ctx)
}
