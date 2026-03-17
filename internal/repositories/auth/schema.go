package auth

const usersTable = "users"

const (
	usersTableColumnID             = "id"
	usersTableColumnUsername       = "username"
	usersTableColumnHashedPassword = "hashed_password"
	usersTableColumnCreatedAt      = "created_at"
)

var usersTableColumns = []string{
	usersTableColumnID,
	usersTableColumnUsername,
	usersTableColumnHashedPassword,
	usersTableColumnCreatedAt,
}
