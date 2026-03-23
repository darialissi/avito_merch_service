package auth

const usersTable = "users"

const (
	usersTableColumnID             = "id"
	usersTableColumnUsername       = "username"
	usersTableColumnHashedPassword = "hashed_password"
	usersTableColumnCoins          = "coins"
	usersTableColumnCreatedAt      = "created_at"
)

var usersTableColumns = []string{
	usersTableColumnID,
	usersTableColumnUsername,
	usersTableColumnHashedPassword,
	usersTableColumnCoins,
	usersTableColumnCreatedAt,
}
