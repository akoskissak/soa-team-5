package utils

import (
	"stakeholders-service/models"

	stakeproto "stakeholders-service/proto/stakeholders"
)

func MapToUserProtoProfile(profile models.UserProfile) *stakeproto.UserProfile {
	if profile.FirstName == "" && profile.LastName == "" {
		return nil
	}
	return &stakeproto.UserProfile{
		FirstName:      profile.FirstName,
		LastName:       profile.LastName,
		ProfilePicture: profile.ProfilePicture,
		Biography:      profile.Biography,
		Motto:          profile.Motto,
	}
}

// MapToUserProfileResponse konvertuje models.User u stakeholders.UserProfileResponse.
func MapToUserProfileResponse(user models.User) *stakeproto.UserProfileResponse {
	return &stakeproto.UserProfileResponse{
		Username:       user.Username,
		FirstName:      user.Profile.FirstName,
		LastName:       user.Profile.LastName,
		ProfilePicture: user.Profile.ProfilePicture,
		Biography:      user.Profile.Biography,
		Motto:          user.Profile.Motto,
	}
}
