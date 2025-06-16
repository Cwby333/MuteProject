package postgres

import (
	"github.com/Cwby333/user-microservice/internal/models"
)

type UserDTO struct {
	ID       string `db:"id"`
	Username string `db:"username"`
	Password string `db:"password"`
	Role     string `db:"role"`
	Email    string `db:"email"`
	VersionCredentials int `db:"version_credentials"`
}

func ToUserDTO(u models.User) UserDTO {
	userDTO := UserDTO{
		ID:       u.ID,
		Username: u.Username,
		Role:     u.Role,
		Password: u.Password,
		Email:    u.Email,
		VersionCredentials: u.VersionCredentials,
	}

	return userDTO
}

func DTOToUser(u UserDTO) models.User {
	user := models.User{
		ID:       u.ID,
		Username: u.Username,
		Role:     u.Role,
		Password: u.Password,
		Email:    u.Email,
		VersionCredentials: u.VersionCredentials,
	}

	return user
}
