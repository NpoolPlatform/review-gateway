package kyc

import (
	"context"
	"fmt"

	"github.com/NpoolPlatform/message/npool/third/mgr/v1/usedfor"

	notif1 "github.com/NpoolPlatform/review-gateway/pkg/notif"

	kycmgrcli "github.com/NpoolPlatform/appuser-manager/pkg/client/kyc"
	usercli "github.com/NpoolPlatform/appuser-middleware/pkg/client/user"
	kycmgrpb "github.com/NpoolPlatform/message/npool/appuser/mgr/v2/kyc"
	review1 "github.com/NpoolPlatform/review-gateway/pkg/review"

	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/kyc"
	reviewmgrpb "github.com/NpoolPlatform/message/npool/review/mgr/v2"
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

	eventType := usedfor.UsedFor_KYCApproved
	kycState := kycmgrpb.KycState_Approved
	if state == reviewmgrpb.ReviewState_Rejected {
		kycState = kycmgrpb.KycState_Rejected
		eventType = usedfor.UsedFor_KYCRejected
	}

	_, err = kycmgrcli.UpdateKyc(ctx, &kycmgrpb.KycReq{
		ID:    &objID,
		State: &kycState,
	})
	if err != nil {
		return nil, err
	}

	notif1.CreateNotif(ctx, appID, kycInfo.UserID, &userInfo.Username, nil, nil, eventType)

	if err != nil {
		return nil, err
	}
	return GetKycReview(ctx, id)
}
