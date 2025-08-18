package forms

type UserLoginForm struct {
	Form
	Email    string `form:"email"`
	Password string `form:"password"`
}

type UserSignUpForm struct {
	Form
	Name            string `form:"name"`
	Email           string `form:"email"`
	Password        string `form:"password"`
	ConfirmPassword string `form:"confirm_password"`
}

type ResetPasswordForm struct {
	Form
	Password        string `form:"password" validate:"required,min=8"`
	ConfirmPassword string `form:"confirm_password" validate:"required"`
	Token           string `form:"token" validate:"required"`
}

type ForgotPasswordForm struct {
	Form
	Email string `form:"email"`
}

type UpdateAccountPasswordForm struct {
	Form
	CurrentPassword string `form:"current_password" validate:"required"`
	NewPassword     string `form:"new_password" validate:"required,min=8"`
	ConfirmPassword string `form:"confirm_password" validate:"required"`
}

type UpdateUserNameAndImageForm struct {
	Form
	Name  string `form:"name"`
	Image string `form:"image"`
}

type DeleteAccountForm struct {
	Form
	Password string `form:"password" validate:"required"`
}
