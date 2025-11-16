package domain

type Team struct {
	Name string
}

type TeamMember struct {
	UserID   string
	Username string
	IsActive bool
}
