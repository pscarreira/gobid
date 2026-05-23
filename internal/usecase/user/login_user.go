package user

import (
	"context"

	"github.com/pscarreira/gobid/internal/validator"
)

type LoginUserReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (req LoginUserReq) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator
	eval.CheckField(validator.IsEmail(req.Email), "email", "must be a valid email address")
	eval.CheckField(validator.NotBlank(req.Password), "password", "must not be empty")
	return eval
}
