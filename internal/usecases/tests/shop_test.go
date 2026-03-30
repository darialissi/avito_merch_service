package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/darialissi/avito_merch_service/internal/mocks"
	"github.com/darialissi/avito_merch_service/internal/models"
	shopRepo "github.com/darialissi/avito_merch_service/internal/repositories/shop"
	"github.com/darialissi/avito_merch_service/internal/schemas/dto"
	"github.com/darialissi/avito_merch_service/internal/usecases"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

type runRepeatableReadResult struct {
	runTx bool
	err   error
}

type getUsersCoinsByUsernamesResult struct {
	users []models.User
	err   error
}

type getItemByNameResult struct {
	item *models.Item
	err  error
}

type getUserItemsByUserIDResult struct {
	items []models.UserItemExtended
	err   error
}

type getTransactionsByUserIDResult struct {
	transactions []models.Transaction
	err          error
}

func TestShopUsecase_SendCoin(t *testing.T) {
	senderUsername := "sender"
	receiverUsername := "receiver"
	data := &dto.TransactionData{
		ToUser: receiverUsername,
		Amount: 30,
	}

	repoErr := errors.New("repo error")
	tmErr := errors.New("tm error")

	tests := []struct {
		name                    string
		runRepeatableReadResp   runRepeatableReadResult
		getUsersCoinsByUsername getUsersCoinsByUsernamesResult
		updateSenderCoinsErr    error
		updateReceiverCoinsErr  error
		saveTransactionErr      error
		wantErr                 error
	}{
		{
			name:                  "Success with reordered users",
			runRepeatableReadResp: runRepeatableReadResult{runTx: true},
			getUsersCoinsByUsername: getUsersCoinsByUsernamesResult{
				users: []models.User{
					{Username: receiverUsername, Coins: 50},
					{Username: senderUsername, Coins: 100},
				},
			},
		},
		{
			name:                  "RunRepeatableRead error",
			runRepeatableReadResp: runRepeatableReadResult{runTx: false, err: tmErr},
			wantErr:               tmErr,
		},
		{
			name:                  "GetUsersCoinsByUsernames error",
			runRepeatableReadResp: runRepeatableReadResult{runTx: true},
			getUsersCoinsByUsername: getUsersCoinsByUsernamesResult{
				err: repoErr,
			},
			wantErr: repoErr,
		},
		{
			name:                  "Less than two users",
			runRepeatableReadResp: runRepeatableReadResult{runTx: true},
			getUsersCoinsByUsername: getUsersCoinsByUsernamesResult{
				users: []models.User{{Username: senderUsername, Coins: 100}},
			},
			wantErr: usecases.ErrUserNotFound,
		},
		{
			name:                  "Not enough coins",
			runRepeatableReadResp: runRepeatableReadResult{runTx: true},
			getUsersCoinsByUsername: getUsersCoinsByUsernamesResult{
				users: []models.User{
					{Username: senderUsername, Coins: 20},
					{Username: receiverUsername, Coins: 50},
				},
			},
			wantErr: usecases.ErrNotEnoughCoins,
		},
		{
			name:                  "Update users coins error",
			runRepeatableReadResp: runRepeatableReadResult{runTx: true},
			getUsersCoinsByUsername: getUsersCoinsByUsernamesResult{
				users: []models.User{
					{Username: senderUsername, Coins: 100},
					{Username: receiverUsername, Coins: 50},
				},
			},
			updateSenderCoinsErr: repoErr,
			wantErr:              repoErr,
		},
		{
			name:                  "Update receiver coins error",
			runRepeatableReadResp: runRepeatableReadResult{runTx: true},
			getUsersCoinsByUsername: getUsersCoinsByUsernamesResult{
				users: []models.User{
					{Username: senderUsername, Coins: 100},
					{Username: receiverUsername, Coins: 50},
				},
			},
			updateReceiverCoinsErr: repoErr,
			wantErr:                repoErr,
		},
		{
			name:                  "SaveTransaction error",
			runRepeatableReadResp: runRepeatableReadResult{runTx: true},
			getUsersCoinsByUsername: getUsersCoinsByUsernamesResult{
				users: []models.User{
					{Username: senderUsername, Coins: 100},
					{Username: receiverUsername, Coins: 50},
				},
			},
			saveTransactionErr: repoErr,
			wantErr:            repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockShopRepository(ctrl)
			tm := mocks.NewMockTransactionManager(ctrl)

			tm.EXPECT().
				RunRepeatableRead(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, f func(txCtx context.Context) error) error {
					if !tt.runRepeatableReadResp.runTx {
						return tt.runRepeatableReadResp.err
					}
					return f(ctx)
				})

			if tt.runRepeatableReadResp.runTx {
				repo.EXPECT().
					GetUsersCoinsByUsernames(gomock.Any(), []string{senderUsername, receiverUsername}, true).
					Return(tt.getUsersCoinsByUsername.users, tt.getUsersCoinsByUsername.err)

				if tt.getUsersCoinsByUsername.err == nil && len(tt.getUsersCoinsByUsername.users) >= 2 {
					sender, receiver := tt.getUsersCoinsByUsername.users[0], tt.getUsersCoinsByUsername.users[1]
					if sender.Username != senderUsername {
						sender, receiver = receiver, sender
					}

					if sender.Coins >= data.Amount {
						senderCoins := sender.Coins - data.Amount
						receiverCoins := receiver.Coins + data.Amount

						repo.EXPECT().
							UpdateUserCoinsByUsername(gomock.Any(), gomock.Any()).
							DoAndReturn(func(_ context.Context, userCoins *dto.UserCoins) ([]models.User, error) {
								if userCoins.Username != sender.Username {
									t.Fatalf("expected sender Username %q, got %q", sender.Username, userCoins.Username)
								}
								if userCoins.Coins != senderCoins {
									t.Fatalf("expected sender Coins %f, got %f", senderCoins, userCoins.Coins)
								}
								return nil, tt.updateSenderCoinsErr
							})

						if tt.updateSenderCoinsErr == nil {
							repo.EXPECT().
								UpdateUserCoinsByUsername(gomock.Any(), gomock.Any()).
								DoAndReturn(func(_ context.Context, userCoins *dto.UserCoins) ([]models.User, error) {
									if userCoins.Username != receiver.Username {
										t.Fatalf("expected receiver Username %q, got %q", receiver.Username, userCoins.Username)
									}
									if userCoins.Coins != receiverCoins {
										t.Fatalf("expected receiver Coins %f, got %f", receiverCoins, userCoins.Coins)
									}
									return nil, tt.updateReceiverCoinsErr
								})

							if tt.updateReceiverCoinsErr == nil {
								repo.EXPECT().
									SaveTransaction(gomock.Any(), gomock.Any()).
									DoAndReturn(func(_ context.Context, txData *dto.TransactionFullData) (*models.Transaction, error) {
										if txData.FromUserID != sender.ID {
											t.Fatalf("expected FromUserID %q, got %q", sender.ID, txData.FromUserID)
										}
										if txData.ToUserID != receiver.ID {
											t.Fatalf("expected ToUserID %q, got %q", receiver.ID, txData.ToUserID)
										}
										if txData.Amount != data.Amount {
											t.Fatalf("expected Amount %f, got %f", data.Amount, txData.Amount)
										}
										return nil, tt.saveTransactionErr
									})
							}
						}
					}
				}
			}

			uc := usecases.NewShopUsecase(repo, tm)
			err := uc.SendCoin(context.Background(), senderUsername, data)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestShopUsecase_BuyItem(t *testing.T) {
	username := "buyer"
	itemID := uuid.New()
	userID := uuid.New()
	data := &dto.BuyItemData{ItemName: "t-shirt", Quantity: 2}

	repoErr := errors.New("repo error")
	tmErr := errors.New("tm error")

	tests := []struct {
		name                  string
		runRepeatableReadResp runRepeatableReadResult
		getItemResp           getItemByNameResult
		getUsersResp          getUsersCoinsByUsernamesResult
		updateUserCoinsErr    error
		upsertUserItemErr     error
		wantErr               error
	}{
		{
			name:                  "Success",
			runRepeatableReadResp: runRepeatableReadResult{runTx: true},
			getItemResp: getItemByNameResult{
				item: &models.Item{ID: itemID, Name: data.ItemName, Price: 40},
			},
			getUsersResp: getUsersCoinsByUsernamesResult{
				users: []models.User{{ID: userID, Username: username, Coins: 100}},
			},
		},
		{
			name:                  "RunRepeatableRead error",
			runRepeatableReadResp: runRepeatableReadResult{runTx: false, err: tmErr},
			wantErr:               tmErr,
		},
		{
			name:                  "GetItemByName not found",
			runRepeatableReadResp: runRepeatableReadResult{runTx: true},
			getItemResp:           getItemByNameResult{err: shopRepo.ErrNotFound},
			wantErr:               usecases.ErrItemNotFound,
		},
		{
			name:                  "GetItemByName error",
			runRepeatableReadResp: runRepeatableReadResult{runTx: true},
			getItemResp:           getItemByNameResult{err: repoErr},
			wantErr:               repoErr,
		},
		{
			name:                  "GetUsersCoinsByUsernames error",
			runRepeatableReadResp: runRepeatableReadResult{runTx: true},
			getItemResp:           getItemByNameResult{item: &models.Item{ID: itemID, Name: data.ItemName, Price: 40}},
			getUsersResp:          getUsersCoinsByUsernamesResult{err: repoErr},
			wantErr:               repoErr,
		},
		{
			name:                  "User not found",
			runRepeatableReadResp: runRepeatableReadResult{runTx: true},
			getItemResp:           getItemByNameResult{item: &models.Item{ID: itemID, Name: data.ItemName, Price: 40}},
			getUsersResp:          getUsersCoinsByUsernamesResult{users: []models.User{}},
			wantErr:               usecases.ErrUserNotFound,
		},
		{
			name:                  "Not enough coins",
			runRepeatableReadResp: runRepeatableReadResult{runTx: true},
			getItemResp:           getItemByNameResult{item: &models.Item{ID: itemID, Name: data.ItemName, Price: 40}},
			getUsersResp:          getUsersCoinsByUsernamesResult{users: []models.User{{ID: userID, Username: username, Coins: 30}}},
			wantErr:               usecases.ErrNotEnoughCoins,
		},
		{
			name:                  "UpdateUserCoinsByUsername error",
			runRepeatableReadResp: runRepeatableReadResult{runTx: true},
			getItemResp:           getItemByNameResult{item: &models.Item{ID: itemID, Name: data.ItemName, Price: 40}},
			getUsersResp:          getUsersCoinsByUsernamesResult{users: []models.User{{ID: userID, Username: username, Coins: 100}}},
			updateUserCoinsErr:    repoErr,
			wantErr:               repoErr,
		},
		{
			name:                  "UpsertUserItemQuantity error",
			runRepeatableReadResp: runRepeatableReadResult{runTx: true},
			getItemResp:           getItemByNameResult{item: &models.Item{ID: itemID, Name: data.ItemName, Price: 40}},
			getUsersResp:          getUsersCoinsByUsernamesResult{users: []models.User{{ID: userID, Username: username, Coins: 100}}},
			upsertUserItemErr:     repoErr,
			wantErr:               repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockShopRepository(ctrl)
			tm := mocks.NewMockTransactionManager(ctrl)

			tm.EXPECT().
				RunRepeatableRead(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, f func(txCtx context.Context) error) error {
					if !tt.runRepeatableReadResp.runTx {
						return tt.runRepeatableReadResp.err
					}
					return f(ctx)
				})

			if tt.runRepeatableReadResp.runTx {
				repo.EXPECT().
					GetItemByName(gomock.Any(), data.ItemName).
					Return(tt.getItemResp.item, tt.getItemResp.err)

				if tt.getItemResp.err == nil {
					repo.EXPECT().
						GetUsersCoinsByUsernames(gomock.Any(), []string{username}, true).
						Return(tt.getUsersResp.users, tt.getUsersResp.err)

					if tt.getUsersResp.err == nil && len(tt.getUsersResp.users) > 0 {
						user := tt.getUsersResp.users[0]
						totalPrice := tt.getItemResp.item.Price * float64(data.Quantity)

						if user.Coins >= totalPrice {
							repo.EXPECT().
								UpdateUserCoinsByUsername(gomock.Any(), &dto.UserCoins{
									Username: user.Username,
									Coins:    user.Coins - totalPrice,
								}).
								Return(nil, tt.updateUserCoinsErr)

							if tt.updateUserCoinsErr == nil {
								repo.EXPECT().
									UpsertUserItemQuantity(gomock.Any(), gomock.Any()).
									DoAndReturn(func(_ context.Context, itemData *dto.UserItemData) (*models.UserItem, error) {
										if itemData.UserID != user.ID {
											t.Fatalf("expected UserID %v, got %v", user.ID, itemData.UserID)
										}
										if itemData.ItemID != tt.getItemResp.item.ID {
											t.Fatalf("expected ItemID %v, got %v", tt.getItemResp.item.ID, itemData.ItemID)
										}
										if itemData.Quantity != data.Quantity {
											t.Fatalf("expected Quantity %d, got %d", data.Quantity, itemData.Quantity)
										}
										return nil, tt.upsertUserItemErr
									})
							}
						}
					}
				}
			}

			uc := usecases.NewShopUsecase(repo, tm)
			err := uc.BuyItem(context.Background(), username, data)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestShopUsecase_Info(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	username := "test_user"
	repoErr := errors.New("repo error")

	tests := []struct {
		name                string
		getUsersResp        getUsersCoinsByUsernamesResult
		getItemsResp        getUserItemsByUserIDResult
		getTransactionsResp getTransactionsByUserIDResult
		wantResp            *dto.AggregatedInfo
		wantErr             error
	}{
		{
			name: "Success",
			getUsersResp: getUsersCoinsByUsernamesResult{
				users: []models.User{{ID: userID, Username: username, Coins: 500}},
			},
			getItemsResp: getUserItemsByUserIDResult{
				items: []models.UserItemExtended{
					{ItemName: "cup", Quantity: 1},
					{ItemName: "pen", Quantity: 2},
				},
			},
			getTransactionsResp: getTransactionsByUserIDResult{
				transactions: []models.Transaction{
					{FromUser: userID, ToUser: otherUserID, Coins: 25},
					{FromUser: otherUserID, ToUser: userID, Coins: 40},
				},
			},
			wantResp: &dto.AggregatedInfo{
				Coins: 500,
				Inventory: []dto.InventoryUnit{
					{ItemName: "cup", Quantity: 1},
					{ItemName: "pen", Quantity: 2},
				},
				CoinHistory: dto.CoinHistory{
					Sent:     []dto.SentTransaction{{ToUser: otherUserID.String(), Amount: 25}},
					Received: []dto.ReceivedTransaction{{FromUser: otherUserID.String(), Amount: 40}},
				},
			},
		},
		{
			name:         "GetUsersCoinsByUsernames error",
			getUsersResp: getUsersCoinsByUsernamesResult{err: repoErr},
			wantErr:      repoErr,
		},
		{
			name:         "User not found",
			getUsersResp: getUsersCoinsByUsernamesResult{users: []models.User{}},
			wantErr:      usecases.ErrUserNotFound,
		},
		{
			name: "GetUserItemsByUserID error",
			getUsersResp: getUsersCoinsByUsernamesResult{
				users: []models.User{{ID: userID, Username: username, Coins: 500}},
			},
			getItemsResp: getUserItemsByUserIDResult{err: repoErr},
			wantErr:      repoErr,
		},
		{
			name: "GetTransactionsByUserID error",
			getUsersResp: getUsersCoinsByUsernamesResult{
				users: []models.User{{ID: userID, Username: username, Coins: 500}},
			},
			getItemsResp: getUserItemsByUserIDResult{items: []models.UserItemExtended{}},
			getTransactionsResp: getTransactionsByUserIDResult{
				err: repoErr,
			},
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockShopRepository(ctrl)
			tm := mocks.NewMockTransactionManager(ctrl)

			repo.EXPECT().
				GetUsersCoinsByUsernames(gomock.Any(), []string{username}, false).
				Return(tt.getUsersResp.users, tt.getUsersResp.err)

			if tt.getUsersResp.err == nil && len(tt.getUsersResp.users) > 0 {
				user := tt.getUsersResp.users[0]

				repo.EXPECT().
					GetUserItemsByUserID(gomock.Any(), user.ID).
					Return(tt.getItemsResp.items, tt.getItemsResp.err)

				if tt.getItemsResp.err == nil {
					repo.EXPECT().
						GetTransactionsByUserID(gomock.Any(), user.ID).
						Return(tt.getTransactionsResp.transactions, tt.getTransactionsResp.err)
				}
			}

			uc := usecases.NewShopUsecase(repo, tm)
			resp, err := uc.Info(context.Background(), username)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp.Coins != tt.wantResp.Coins {
				t.Fatalf("expected Coins %f, got %f", tt.wantResp.Coins, resp.Coins)
			}
			if len(resp.Inventory) != len(tt.wantResp.Inventory) {
				t.Fatalf("expected inventory len %d, got %d", len(tt.wantResp.Inventory), len(resp.Inventory))
			}
			for i := range resp.Inventory {
				if resp.Inventory[i].ItemName != tt.wantResp.Inventory[i].ItemName {
					t.Fatalf("expected Inventory[%d].ItemName %q, got %q", i, tt.wantResp.Inventory[i].ItemName, resp.Inventory[i].ItemName)
				}
				if resp.Inventory[i].Quantity != tt.wantResp.Inventory[i].Quantity {
					t.Fatalf("expected Inventory[%d].Quantity %d, got %d", i, tt.wantResp.Inventory[i].Quantity, resp.Inventory[i].Quantity)
				}
			}

			if len(resp.CoinHistory.Sent) != len(tt.wantResp.CoinHistory.Sent) {
				t.Fatalf("expected sent len %d, got %d", len(tt.wantResp.CoinHistory.Sent), len(resp.CoinHistory.Sent))
			}
			for i := range resp.CoinHistory.Sent {
				if resp.CoinHistory.Sent[i].ToUser != tt.wantResp.CoinHistory.Sent[i].ToUser {
					t.Fatalf("expected Sent[%d].ToUser %q, got %q", i, tt.wantResp.CoinHistory.Sent[i].ToUser, resp.CoinHistory.Sent[i].ToUser)
				}
				if resp.CoinHistory.Sent[i].Amount != tt.wantResp.CoinHistory.Sent[i].Amount {
					t.Fatalf("expected Sent[%d].Amount %f, got %f", i, tt.wantResp.CoinHistory.Sent[i].Amount, resp.CoinHistory.Sent[i].Amount)
				}
			}

			if len(resp.CoinHistory.Received) != len(tt.wantResp.CoinHistory.Received) {
				t.Fatalf("expected received len %d, got %d", len(tt.wantResp.CoinHistory.Received), len(resp.CoinHistory.Received))
			}
			for i := range resp.CoinHistory.Received {
				if resp.CoinHistory.Received[i].FromUser != tt.wantResp.CoinHistory.Received[i].FromUser {
					t.Fatalf("expected Received[%d].FromUser %q, got %q", i, tt.wantResp.CoinHistory.Received[i].FromUser, resp.CoinHistory.Received[i].FromUser)
				}
				if resp.CoinHistory.Received[i].Amount != tt.wantResp.CoinHistory.Received[i].Amount {
					t.Fatalf("expected Received[%d].Amount %f, got %f", i, tt.wantResp.CoinHistory.Received[i].Amount, resp.CoinHistory.Received[i].Amount)
				}
			}
		})
	}
}
