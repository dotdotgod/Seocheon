package signing

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// vaultSignRequest is the JSON body for POST /v1/keys/{name}/sign.
type vaultSignRequest struct {
	Data string `json:"data"`
}

// vaultSignResponse is the JSON response from POST /v1/keys/{name}/sign.
type vaultSignResponse struct {
	Signature string `json:"signature"`
}

// vaultAddressResponse is the JSON response from GET /v1/keys/{name}/address.
type vaultAddressResponse struct {
	Address string `json:"address"`
}

// vaultPubKeyResponse is the JSON response from GET /v1/keys/{name}/pubkey.
type vaultPubKeyResponse struct {
	PubKey string `json:"pubkey"`
}

// initVaultService initializes the VaultService by fetching address and pubkey
// from the vault server.
func initVaultService(endpoint, keyName string) (*VaultService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 10 * time.Second}
	baseURL := strings.TrimRight(endpoint, "/")

	// Fetch address
	addrURL := fmt.Sprintf("%s/v1/keys/%s/address", baseURL, keyName)
	addrResp, err := doVaultGet(ctx, client, addrURL)
	if err != nil {
		return nil, fmt.Errorf("fetching vault address: %w", err)
	}

	var addrResult vaultAddressResponse
	if err := json.Unmarshal(addrResp, &addrResult); err != nil {
		return nil, fmt.Errorf("parsing vault address response: %w", err)
	}

	// Fetch public key
	pubKeyURL := fmt.Sprintf("%s/v1/keys/%s/pubkey", baseURL, keyName)
	pubKeyResp, err := doVaultGet(ctx, client, pubKeyURL)
	if err != nil {
		return nil, fmt.Errorf("fetching vault pubkey: %w", err)
	}

	var pubKeyResult vaultPubKeyResponse
	if err := json.Unmarshal(pubKeyResp, &pubKeyResult); err != nil {
		return nil, fmt.Errorf("parsing vault pubkey response: %w", err)
	}

	pubKeyBytes, err := hex.DecodeString(pubKeyResult.PubKey)
	if err != nil {
		return nil, fmt.Errorf("decoding vault pubkey hex: %w", err)
	}

	return &VaultService{
		endpoint: baseURL,
		keyName:  keyName,
		address:  addrResult.Address,
		pubKey:   pubKeyBytes,
		client:   client,
	}, nil
}

// vaultSign sends a sign request to the vault server.
func vaultSign(client *http.Client, endpoint, keyName string, data []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/v1/keys/%s/sign", endpoint, keyName)

	reqBody := vaultSignRequest{
		Data: hex.EncodeToString(data),
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling sign request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("creating sign request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending sign request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading sign response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vault sign failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var signResult vaultSignResponse
	if err := json.Unmarshal(respBody, &signResult); err != nil {
		return nil, fmt.Errorf("parsing sign response: %w", err)
	}

	sigBytes, err := hex.DecodeString(signResult.Signature)
	if err != nil {
		return nil, fmt.Errorf("decoding signature hex: %w", err)
	}

	return sigBytes, nil
}

// doVaultGet performs a GET request to the vault server.
func doVaultGet(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vault request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
