package kyc

import (
	"context"
	"fmt"

	kycmgrcli "github.com/NpoolPlatform/appuser-manager/pkg/client/kyc"
	usercli "github.com/NpoolPlatform/appuser-middleware/pkg/client/user"
	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	kycmgrpb "github.com/NpoolPlatform/message/npool/appuser/mgr/v2/kyc"
	review1 "github.com/NpoolPlatform/review-gateway/pkg/review"

	"github.com/NpoolPlatform/libent-cruder/pkg/cruder"

	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/kyc"
	reviewmgrpb "github.com/NpoolPlatform/message/npool/review/mgr/v2"

	channelpb "github.com/NpoolPlatform/message/npool/notif/mgr/v1/channel"
	notifmgrpb "github.com/NpoolPlatform/message/npool/notif/mgr/v1/notif"
	notifcli "github.com/NpoolPlatform/notif-middleware/pkg/client/notif"

	thirdmgrpb "github.com/NpoolPlatform/message/npool/third/mgr/v1/template/notif"
	thirdcli "github.com/NpoolPlatform/third-middleware/pkg/client/template/notif"
	thirdpkg "github.com/NpoolPlatform/third-middleware/pkg/template/notif"

	npoolpb "github.com/NpoolPlatform/message/npool"
)

func UpdateKycReview(
	ctx context.Context,
	id, appID, reviewerAppID, reviewerID, langID string,
	state reviewmgrpb.ReviewState,
	message *string,
) (
	*npool.KycReview, error,
) {
	objID, err := review1.ValidateReview(ctx, id, appID, reviewerAppID, reviewerID, state)
	if err != nil {
		return nil, err
	}

	kycInfo, err := kycmgrcli.GetKyc(ctx, objID)
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

	if err := review1.UpdateReview(ctx, id, appID, reviewerAppID, reviewerID, state, message); err != nil {
		return nil, err
	}

	eventType := notifmgrpb.EventType_KYCApproved
	kycState := kycmgrpb.KycState_Approved
	if state == reviewmgrpb.ReviewState_Rejected {
		kycState = kycmgrpb.KycState_Rejected
		eventType = notifmgrpb.EventType_KYCRejected
	}

	_, err = kycmgrcli.UpdateKyc(ctx, &kycmgrpb.KycReq{
		ID:    &objID,
		State: &kycState,
	})
	if err != nil {
		return nil, err
	}

	createNotif(ctx, appID, kycInfo.UserID, langID, userInfo.Username, eventType)

	if err != nil {
		return nil, err
	}
	return GetKycReview(ctx, id)
}

func createNotif(
	ctx context.Context,
	appID, userID, langID, userName string,
	eventType notifmgrpb.EventType,
) {
	templateInfo, err := thirdcli.GetNotifTemplateOnly(ctx, &thirdmgrpb.Conds{
		AppID: &npoolpb.StringVal{
			Op:    cruder.EQ,
			Value: appID,
		},
		LangID: &npoolpb.StringVal{
			Op:    cruder.EQ,
			Value: langID,
		},
		UsedFor: &npoolpb.Uint32Val{
			Op:    cruder.EQ,
			Value: uint32(eventType.Number()),
		},
	})
	if err != nil {
		logger.Sugar().Errorw("sendNotif", "error", err.Error())
		return
	}
	if templateInfo == nil {
		logger.Sugar().Errorw("sendNotif", "error", "template not exist")
		return
	}

	content := thirdpkg.ReplaceVariable(templateInfo.Content, &userName, nil)
	useTemplate := true

	_, err = notifcli.CreateNotif(ctx, &notifmgrpb.NotifReq{
		AppID:       &appID,
		UserID:      &userID,
		LangID:      &langID,
		EventType:   &eventType,
		UseTemplate: &useTemplate,
		Title:       &templateInfo.Title,
		Content:     &content,
		Channels:    []channelpb.NotifChannel{channelpb.NotifChannel_ChannelEmail},
	})
	if err != nil {
		logger.Sugar().Errorw("sendNotif", "error", err.Error())
		return
	}
}
