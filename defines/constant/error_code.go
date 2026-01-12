package constant

const (
	ErrDB                 = 1
	ErrDBClose            = 2
	ErrDBInvalid          = 3
	ErrInternalServer     = 4
	ErrParam              = 6
	ErrJsonEncodeFail     = 7
	ErrFile               = 8
	ErrNotFound           = -100404 //Not Found
	ErrBadRequest         = -100400 //Bad Request
	ErrNotLogin           = -100405 //Not Login
	ErrAuth               = -100401 //Unauthorized 认证失败
	ErrPermission         = -100403 //Forbidden 权限不足
	ErrUserLoginForbidden = -224004
	ErrInventoryNotEnough = -200001 //库存不足
)
