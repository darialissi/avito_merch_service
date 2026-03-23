package shop

const usersTable = "users"

const (
	usersTableColumnID       = "id"
	usersTableColumnUsername = "username"
	usersTableColumnCoins    = "coins"
)

var usersTableColumns = []string{
	usersTableColumnID,
	usersTableColumnUsername,
	usersTableColumnCoins,
}

const itemsTable = "items"
const itemsTablePrefix = "i"

const (
	itemsTableColumnID    = "id"
	itemsTableColumnName  = "name"
	itemsTableColumnPrice = "price"
)

var itemsTableColumns = []string{
	itemsTableColumnID,
	itemsTableColumnName,
	itemsTableColumnPrice,
}

const userItemsTable = "user_items"
const userItemsTablePrefix = "ui"

const (
	userItemsTableColumnUserID   = "user_id"
	userItemsTableColumnItemID   = "item_id"
	userItemsTableColumnQuantity = "quantity"
)

var userItemsTableColumns = []string{
	userItemsTableColumnUserID,
	userItemsTableColumnItemID,
	userItemsTableColumnQuantity,
}

const transactionsTable = "transactions"

const (
	transactionsTableColumnID         = "id"
	transactionsTableColumnFromUserID = "from_user_id"
	transactionsTableColumnToUserID   = "to_user_id"
	transactionsTableColumnCoins      = "coins"
	transactionsTableColumnCreatedAt  = "created_at"
)

var transactionsTableColumns = []string{
	transactionsTableColumnID,
	transactionsTableColumnFromUserID,
	transactionsTableColumnToUserID,
	transactionsTableColumnCoins,
	transactionsTableColumnCreatedAt,
}
