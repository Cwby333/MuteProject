package models

type User struct {
	ID       string
	Username string
	Password string
	Email    string
	Role     string
	VersionCredentials int
}
