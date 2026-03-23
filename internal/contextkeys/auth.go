package contextkeys

type contextKey string

const (
	UserKey   contextKey = "auth_username"
	UserIDKey contextKey = "auth_user_id"
)
