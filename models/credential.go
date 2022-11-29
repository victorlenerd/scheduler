package models

import (
	"encoding/json"
	"time"
)

// CredentialModel credential model
type CredentialModel struct {
	ID          int64     `json:"id"`
	Archived    bool      `json:"archived"`
	ApiKey      string    `json:"api_key"`
	ApiSecret   string    `json:"api_secret"`
	DateCreated time.Time `json:"date_created"`
}

// PaginatedCredential paginated container of credential transformer
type PaginatedCredential struct {
	Total  int64             `json:"total"`
	Offset int64             `json:"offset"`
	Limit  int64             `json:"limit"`
	Data   []CredentialModel `json:"credentials"`
}

// ToJSON returns content of transformer as JSON
func (credentialModel *CredentialModel) ToJSON() ([]byte, error) {
	if data, err := json.Marshal(credentialModel); err != nil {
		return data, err
	} else {
		return data, nil
	}
}

// FromJSON extracts content of JSON into transformer
func (credentialModel *CredentialModel) FromJSON(body []byte) error {
	if err := json.Unmarshal(body, credentialModel); err != nil {
		return err
	}
	return nil
}
