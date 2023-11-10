package kyc

import (
	"context"
	"fmt"

	kycmwcli "github.com/NpoolPlatform/appuser-middleware/pkg/client/kyc"
	usercli "github.com/NpoolPlatform/appuser-middleware/pkg/client/user"
	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	kycmwpb "github.com/NpoolPlatform/message/npool/appuser/mw/v1/kyc"
	appusermwpb "github.com/NpoolPlatform/message/npool/appuser/mw/v1/user"
	reviewtypes "github.com/NpoolPlatform/message/npool/basetypes/review/v1"
	basetypes "github.com/NpoolPlatform/message/npool/basetypes/v1"
	notifmwpb "github.com/NpoolPlatform/message/npool/notif/mw/v1/notif"
	tmplmwpb "github.com/NpoolPlatform/message/npool/notif/mw/v1/template"
	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/kyc"
	reviewmwpb "github.com/NpoolPlatform/message/npool/review/mw/v2/review"
	notifmwcli "github.com/NpoolPlatform/notif-middleware/pkg/client/notif"
	reviewcli "github.com/NpoolPlatform/review-middleware/pkg/client/review"
)

type updateHandler struct {
	*Handler
	review *reviewmwpb.Review
	kyc    *kycmwpb.Kyc
	user   *appusermwpb.User
}

func (h *updateHandler) checkUser(ctx context.Context) error {
	info, err := usercli.GetUser(ctx, *h.AppID, *h.UserID)
	if err != nil {
		return err
	}
	if info == nil {
		return fmt.Errorf("invalid user")
	}
	return nil
}

func (h *updateHandler) checkReview(ctx context.Context) error {
	info, err := reviewcli.GetReview(ctx, *h.ReviewID)
	if err != nil {
		return err
	}
	if info == nil {
		return fmt.Errorf("invalid review")
	}
	if *h.TargetAppID != info.AppID {
		return fmt.Errorf("appid mismatch")
	}
	if info.State != reviewtypes.ReviewState_Wait {
		return fmt.Errorf("current state can not be updated")
	}
	if *h.State == reviewtypes.ReviewState_Rejected && h.Message == nil {
		return fmt.Errorf("message is must")
	}

	h.review = info
	return nil
}

func (h *updateHandler) getKyc(ctx context.Context) error {
	info, err := kycmwcli.GetKyc(ctx, h.review.ObjectID)
	if err != nil {
		return err
	}
	if info == nil {
		return fmt.Errorf("invalid kyc")
	}
	if info.ReviewID != *h.ReviewID {
		return fmt.Errorf("reviewid mismatch")
	}
	h.KycID = &info.ID
	h.kyc = info

	info1, err := usercli.GetUser(ctx, info.AppID, info.UserID)
	if err != nil {
		return err
	}
	if info1 == nil {
		return fmt.Errorf("invalid user")
	}
	h.user = info1

	return nil
}

func (h *updateHandler) updateReview(ctx context.Context) error {
	if _, err := reviewcli.UpdateReview(ctx, &reviewmwpb.ReviewReq{
		ID:         &h.review.ID,
		ReviewerID: h.UserID,
		State:      h.State,
		Message:    h.Message,
	}); err != nil {
		return err
	}
	return nil
}

func (h *updateHandler) updateKyc(ctx context.Context) error {
	kycState := basetypes.KycState_Approved
	if *h.State == reviewtypes.ReviewState_Rejected {
		kycState = basetypes.KycState_Rejected
	}

	if _, err := kycmwcli.UpdateKyc(ctx, &kycmwpb.KycReq{
		ID:    &h.kyc.ID,
		State: &kycState,
	}); err != nil {
		return err
	}
	return nil
}

func (h *updateHandler) generateNotifs(ctx context.Context) {
	eventType := basetypes.UsedFor_KYCApproved
	if *h.State == reviewtypes.ReviewState_Rejected {
		eventType = basetypes.UsedFor_KYCRejected
	}
	if _, err := notifmwcli.GenerateNotifs(ctx, &notifmwpb.GenerateNotifsRequest{
		AppID:     h.kyc.AppID,
		UserID:    &h.kyc.UserID,
		EventType: eventType,
		NotifType: basetypes.NotifType_NotifUnicast,
		Vars: &tmplmwpb.TemplateVars{
			Username: &h.user.Username,
		},
	}); err != nil {
		logger.Sugar().Errorw("UpdateKycReview", "Generate Notif Failed", "Error", err)
	}
}

func (h *Handler) UpdateKycReview(ctx context.Context) (*npool.KycReview, error) {
	handler := &updateHandler{
		Handler: h,
	}

	if err := handler.checkUser(ctx); err != nil {
		return nil, err
	}
	if err := handler.checkReview(ctx); err != nil {
		return nil, err
	}
	if err := handler.getKyc(ctx); err != nil {
		return nil, err
	}
	if err := handler.updateReview(ctx); err != nil {
		return nil, err
	}
	if err := handler.updateKyc(ctx); err != nil {
		return nil, err
	}
	handler.generateNotifs(ctx)

	return h.GetKycReview(ctx)
}
