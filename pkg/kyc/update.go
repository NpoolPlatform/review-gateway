package kyc

import (
	"context"
	"fmt"

	npool "github.com/NpoolPlatform/message/npool/review/gw/v2/kyc"
	reviewmgrpb "github.com/NpoolPlatform/message/npool/review/mgr/v2"

	kycmgrcli "github.com/NpoolPlatform/appuser-manager/pkg/client/kyc"
	kycmgrpb "github.com/NpoolPlatform/message/npool/appuser/mgr/v2/kyc"
	review1 "github.com/NpoolPlatform/review-gateway/pkg/review"
)

func UpdateKycReview(
	ctx context.Context,
	id, appID, reviewerAppID, reviewerID string,
	state reviewmgrpb.ReviewState,
	message string,
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

	if err := review1.UpdateReview(ctx, id, appID, reviewerAppID, reviewerID, state, message); err != nil {
		return nil, err
	}

	kycState := kycmgrpb.KycState_Approved
	if state == reviewmgrpb.ReviewState_Rejected {
		kycState = kycmgrpb.KycState_Rejected
	}

	_, err = kycmgrcli.UpdateKyc(ctx, &kycmgrpb.KycReq{
		ID:    &objID,
		State: &kycState,
	})
	if err != nil {
		return nil, err
	}
	return GetKycReview(ctx, id)
}
