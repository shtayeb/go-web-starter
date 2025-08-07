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
	Password        string `form:"password"`
	ConfirmPassword string `form:"confirm_password"`
	Token           string `form:"token"`
}

type ForgotPasswordForm struct {
	Form
	Email string `form:"email"`
}

type UpdateAccountPasswordForm struct {
	Form
	CurrentPassword string `form:"current_password"`
	NewPassword     string `form:"new_password"`
	ConfirmPassword string `form:"confirm_password"`
}

type UpdateUserNameAndImageForm struct {
	Form
	Name  string `form:"name"`
	Image string `form:"image"`
}
