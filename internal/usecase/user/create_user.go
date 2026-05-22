package user

import (
	"context"

	"github.com/pscarreira/gobid/internal/validator"
)

type CreateUserReq struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Bio      string `json:"bio"`
}

func (c CreateUserReq) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator
	eval.CheckField(validator.NotBlank(c.Username), "username", "Username cannot be blank")
	eval.CheckField(validator.NotBlank(c.Email), "email", "Email cannot be blank")
	eval.CheckField(validator.IsEmail(c.Email), "email", "Email is not valid")
	eval.CheckField(validator.MinChars(c.Password, 8), "password", "Password must be at least 8 characters long")
	eval.CheckField(validator.NotBlank(c.Bio), "bio", "Bio cannot be blank")
	eval.CheckField(
		validator.MinChars(c.Bio, 10) &&
			validator.MaxChars(c.Bio, 255),
		"bio", "Bio must be between 10 and 255 characters long",
	)
	return eval
}
