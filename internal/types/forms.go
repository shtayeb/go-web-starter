package types

import "go-htmx-sqlite/internal/validator"

type UserLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type UserSignUpForm struct {
	Name                 string `form:"name"`
	Email                string `form:"email"`
	Password             string `form:"password"`
	PasswordConfirmation string `form:"password-confirmation"`
	validator.Validator  `form:"-"`
}

type ResetPasswordForm struct {
	Password             string `form:"password"`
	PasswordConfirmation string `form:"password-confirmation"`
	Token                string `form:"token"`
	validator.Validator  `form:"-"`
}

type ForgotPasswordForm struct {
	Email               string `form:"email"`
	validator.Validator `form:"-"`
}

type UpdateAccountPasswordForm struct {
	CurrentPassword     string `form:"current_password"`
	NewPassword         string `form:"new_password"`
	ConfirmPassword     string `form:"password_confirmation"`
	validator.Validator `form:"-"`
}

type UpdateUserNameAndImageForm struct {
	Name                string `form:"name"`
	Image               string `form:"image"`
	validator.Validator `form:"-"`
}
