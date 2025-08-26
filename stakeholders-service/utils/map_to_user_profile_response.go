package utils

import "stakeholders-service/models"

func MapToUserProfileResponse(user models.User) models.UserProfileResponse {
	return models.UserProfileResponse{
		Username: user.Username,
		FirstName: user.Profile.FirstName,
		LastName: user.Profile.LastName,
		ProfilePicture: user.Profile.ProfilePicture,
		Biography: user.Profile.Biography,
		Motto: user.Profile.Motto,
	}
}