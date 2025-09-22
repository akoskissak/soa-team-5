package rest_clients

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type PurchaseClient struct {
	BaseURL string
	Client  *http.Client
}

type HasPurchasedTourResponse struct {
	Purchased bool `json:"purchased"`
}

func NewPurchaseClient(baseURL string) *PurchaseClient {
	return &PurchaseClient{
		BaseURL: baseURL,
		Client:  &http.Client{},
	}
}

func (pc *PurchaseClient) HasPurchasedTour(userID, tourID string) (bool, error) {
	url := fmt.Sprintf("%s/api/tourist/%s/has-purchased/%s", pc.BaseURL, userID, tourID)
	resp, err := pc.Client.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("purchase-service returned status %d", resp.StatusCode)
	}

	var result HasPurchasedTourResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	return result.Purchased, nil
}
