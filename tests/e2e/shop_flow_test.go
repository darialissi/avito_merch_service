package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/darialissi/avito_merch_service/internal/schemas/dto"
	"github.com/darialissi/avito_merch_service/tests/testutil"
)

func TestE2E_ShopFlow_RegisterAuthBuySendAndInfo(t *testing.T) {
	env := testutil.SetupEnv(t)
	handler := testutil.BuildHTTPHandler(t, env)
	server := httptest.NewServer(handler)
	defer server.Close()

	alice := dto.AuthForm{Username: "alice_user", Password: "StrongPass01"}
	bob := dto.AuthForm{Username: "bob_user", Password: "StrongPass02"}

	registerUser(t, server.URL, alice)
	registerUser(t, server.URL, bob)

	aliceTokens := authUser(t, server.URL, alice)
	bobTokens := authUser(t, server.URL, bob)

	buyItem(t, server.URL, aliceTokens.AccessToken, "book", dto.BuyItemRequest{Quantity: 2})
	sendCoin(t, server.URL, aliceTokens.AccessToken, dto.TransactionData{ToUser: bob.Username, Amount: 100})

	aliceInfo := getInfo(t, server.URL, aliceTokens.AccessToken)
	require.Equal(t, 800.0, aliceInfo.Coins)
	require.Len(t, aliceInfo.Inventory, 1)
	require.Equal(t, "book", aliceInfo.Inventory[0].ItemName)
	require.Equal(t, 2, aliceInfo.Inventory[0].Quantity)
	require.Len(t, aliceInfo.CoinHistory.Sent, 1)
	require.Equal(t, 100.0, aliceInfo.CoinHistory.Sent[0].Amount)

	bobInfo := getInfo(t, server.URL, bobTokens.AccessToken)
	require.Equal(t, 1100.0, bobInfo.Coins)
	require.Empty(t, bobInfo.Inventory)
	require.Len(t, bobInfo.CoinHistory.Received, 1)
	require.Equal(t, 100.0, bobInfo.CoinHistory.Received[0].Amount)
}

func registerUser(t *testing.T, baseURL string, form dto.AuthForm) {
	t.Helper()

	resp := doJSONRequest(t, http.MethodPost, baseURL+"/api/register", form, "")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func authUser(t *testing.T, baseURL string, form dto.AuthForm) dto.AuthResponse {
	t.Helper()

	resp := doJSONRequest(t, http.MethodPost, baseURL+"/api/auth", form, "")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var out dto.AuthResponse
	decodeJSON(t, resp.Body, &out)
	require.NotEmpty(t, out.AccessToken)
	require.NotEmpty(t, out.RefreshToken)

	return out
}

func buyItem(t *testing.T, baseURL, token, item string, payload dto.BuyItemRequest) {
	t.Helper()

	resp := doJSONRequest(t, http.MethodPost, baseURL+"/api/buy/"+item, payload, token)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func sendCoin(t *testing.T, baseURL, token string, payload dto.TransactionData) {
	t.Helper()

	resp := doJSONRequest(t, http.MethodPost, baseURL+"/api/sendCoin", payload, token)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func getInfo(t *testing.T, baseURL, token string) dto.AggregatedInfo {
	t.Helper()

	resp := doJSONRequest(t, http.MethodGet, baseURL+"/api/info", nil, token)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var out dto.AggregatedInfo
	decodeJSON(t, resp.Body, &out)
	return out
}

func doJSONRequest(t *testing.T, method, url string, payload any, bearerToken string) *http.Response {
	t.Helper()

	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		require.NoError(t, err)
		body = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest(method, url, body)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	return resp
}

func decodeJSON(t *testing.T, body io.Reader, out any) {
	t.Helper()

	require.NoError(t, json.NewDecoder(body).Decode(out))
}
