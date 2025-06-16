package userrouter

import "github.com/Cwby333/user-microservice/internal/models"

//easyjson:json
type UserDTO struct {
	ID       string `json:"id" omitempty:"true"`
	Username string `json:"username" omitempty:"true"`
	Password string `json:"password" omitempty:"true"`
	Email    string `json:"email" omitempty:"true"`
	Role     string `json:"role" omitempty:"true"`
	VersionCredentials int `json:"version_credentials" omitempty:"true"`
}

func UserToDTO(u models.User) UserDTO {
	DTO := UserDTO{
		ID:       u.ID,
		Username: u.Username,
		Password: u.Password,
		Email:    u.Email,
		Role:     u.Role,
		VersionCredentials: u.VersionCredentials,
	}

	return DTO
}

func DTOToUser(u UserDTO) models.User {
	user := models.User{
		ID:       u.ID,
		Username: u.Username,
		Password: u.Password,
		Email:    u.Email,
		Role:     u.Role,
		VersionCredentials: u.VersionCredentials,
	}

	return user
}
