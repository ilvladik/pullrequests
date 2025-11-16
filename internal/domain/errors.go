package domain

type ErrCode string

const (
	ErrTeamExistsCode  ErrCode = "TEAM_EXISTS"
	ErrPRExistsCode    ErrCode = "PR_EXISTS"
	ErrPRMergedCode    ErrCode = "PR_MERGED"
	ErrNotAssignedCode ErrCode = "NOT_ASSIGNED"
	ErrNoCandidateCode ErrCode = "NO_CANDIDATE"
	ErrNotFoundCode    ErrCode = "NOT_FOUND"
	ErrInternalCode    ErrCode = "INTERNAL_ERROR"
)

var descriptions = map[ErrCode]string{
	ErrTeamExistsCode:  "team already exists",
	ErrPRExistsCode:    "PR already exists",
	ErrPRMergedCode:    "PR is already merged",
	ErrNotAssignedCode: "item is not assigned",
	ErrNoCandidateCode: "no candidate found",
	ErrNotFoundCode:    "resource not found",
	ErrInternalCode:    "internal server error",
}

type DomainError struct {
	code    ErrCode
	message string
}

func (err DomainError) Error() string {
	return string(err.code) + ": " + err.message
}

func (err DomainError) Code() string {
	return string(err.code)
}

func (err DomainError) Message() string {
	return err.message
}

func NewDomainError(code ErrCode) DomainError {
	if desc, ok := descriptions[code]; ok {
		return DomainError{code: code, message: desc}
	}
	return DomainError{code: ErrInternalCode, message: descriptions[ErrInternalCode]}
}
