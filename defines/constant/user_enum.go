package constant

type UserRoleType = int64

const (
	UserRoleTypeSuperAdmin UserRoleType = 1
	UserRoleTypeAdmin      UserRoleType = 2
	UserRoleTypeTeacher    UserRoleType = 3
	UserRoleTypeStudent    UserRoleType = 4
)
