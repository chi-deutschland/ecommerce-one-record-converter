// Package neone provides a client for interacting with the NE:ONE Server API.
//
// It handles posting Logistics Objects and notifications, with built-in support
// for rate limiting, retries with exponential backoff, and request timeouts via
// failsafe-go policies. The client also verifies object availability after
// creation to prevent race conditions in rapid successive requests.
package neone

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"strings"
	"time"

	"chi-deutschland.com/ecommerce-one-record-converter/pkg/iata/onerecord"
	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/failsafehttp"
	"github.com/failsafe-go/failsafe-go/timeout"
	"github.com/rs/zerolog/log"
)

// StatusError represents an HTTP error response with a status code and body.
type StatusError struct {
	StatusCode int
	Body       []byte
}

// Error implements the error interface for StatusError.
func (e *StatusError) Error() string {
	return fmt.Sprintf("status %d, body: %s", e.StatusCode, string(e.Body))
}

// Client is a NE-ONE API client for interacting with the NE-ONE server.
type Client struct {
	client                 *http.Client
	accessDelegationURLs   []string
	validateObjectCreation bool
	policies               []failsafe.Policy[*http.Response]
}

// NewServer creates a new NE-ONE API client with the given HTTP client and
// optional failsafe policies.
func NewServer(
	client *http.Client,
	accessDelegationURLs []string,
	validateObjectCreation bool,
	policies ...failsafe.Policy[*http.Response],
) *Client {
	return &Client{
		client:                 client,
		accessDelegationURLs:   accessDelegationURLs,
		validateObjectCreation: validateObjectCreation,
		policies:               policies,
	}
}

// PostLogisticsObject sends a logistics object to the NE-ONE server and returns
// the created object's URL. It performs a POST request with the provided
// logisticsObject and handles error responses.
func (s *Client) PostLogisticsObject(
	ctx context.Context,
	baseURL string,
	auth string,
	logisticsObject onerecord.LogisticsObject,
) (string, error) {
	requestBody, err := json.Marshal(logisticsObject)
	if err != nil {
		return "", fmt.Errorf("failed to marshal logistic object: %w", err)
	}

	const logisticsObjectEndpoint = "logistics-objects/"

	url, err := urlpkg.JoinPath(baseURL, logisticsObjectEndpoint)
	if err != nil {
		return "", fmt.Errorf("failed to join URL path: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header = map[string][]string{
		"Authorization": {auth},
		"Content-Type":  {"application/ld+json"},
	}

	resp, err := failsafehttp.NewRequest(req, s.client, s.policies...).Do()
	if err != nil {
		log.Debug().Str("request_body", string(requestBody)).Msg("request body that caused error response")

		return "", fmt.Errorf("failed to send request: %w", err)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Err(err).Msg("failed to close response body")
		}
	}()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		log.Debug().Str("request_body", string(requestBody)).Msg("request body that caused error response")

		return "", &StatusError{
			StatusCode: resp.StatusCode,
			Body:       responseBody,
		}
	}

	logisticsObjectURL := resp.Header.Get("Location")

	log.Debug().
		Str("location_header", logisticsObjectURL).
		Msg("Received logistics object response from NE-ONE server")

	if s.validateObjectCreation {
		s.waitForObjectAvailability(ctx, baseURL, auth, logisticsObjectURL)
	}

	return logisticsObjectURL, nil
}

// ValidateToken sends a test request to the NE-ONE server to check if the
// provided token is valid. It returns an error if the token is invalid or the
// request fails.
func (s *Client) ValidateToken(ctx context.Context, baseURL, auth string) error {
	const logisticsObjectsEndpoint = "logistics-objects/internal/_all"

	url, err := urlpkg.JoinPath(baseURL, logisticsObjectsEndpoint)
	if err != nil {
		return fmt.Errorf("failed to join URL path: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header = map[string][]string{
		"Authorization": {auth},
		"Accept":        {"application/ld+json"},
	}

	urlQuery := req.URL.Query()
	urlQuery.Add("limit", "1")
	req.URL.RawQuery = urlQuery.Encode()

	const tokenValidationTimeout = 8 * time.Second

	tokenValidationPolicies := []failsafe.Policy[*http.Response]{
		timeout.NewBuilder[*http.Response](tokenValidationTimeout).Build(),
	}

	// using a specific policy as the frontend is synchronously waiting for token
	// validation.
	resp, err := failsafehttp.NewRequest(req, s.client, tokenValidationPolicies...).Do()
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Err(err).Msg("failed to close response body")
		}
	}()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &StatusError{
			StatusCode: resp.StatusCode,
			Body:       responseBody,
		}
	}

	return nil
}

// DelegateAccess sends a request to the NE-ONE server to delegate access to the
// created logistics object for the configured NE-ONE servers. This allows the
// specified NE-ONE servers to access the logistics object as needed.
func (s *Client) DelegateAccess(
	ctx context.Context,
	baseURL string,
	auth string,
	description string,
	objectIDs []string,
) (string, error) {
	if len(s.accessDelegationURLs) == 0 || len(objectIDs) == 0 {
		return "", nil
	}

	const delegateAccessEndpoint = "access-delegations"

	url, err := urlpkg.JoinPath(baseURL, delegateAccessEndpoint)
	if err != nil {
		return "", fmt.Errorf("failed to join URL path: %w", err)
	}

	idOnlyLOs := make([]onerecord.IDOnlyLogisticsObject, len(objectIDs))
	for i, objectID := range objectIDs {
		idOnlyLOs[i] = onerecord.IDOnlyLogisticsObject{
			ID: objectID,
		}
	}

	delegationRequest := onerecord.CargoAPIAccessDelegationRequest(
		description,
		[]onerecord.APIPermission{
			onerecord.APIPermissionGetLogisticsObject(),
			onerecord.APIPermissionPatchLogisticsObject(),
		},
		s.accessDelegationURLs,
		idOnlyLOs,
	)

	requestBody, err := json.Marshal(delegationRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal delegation request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header = map[string][]string{
		"Authorization": {auth},
		"Content-Type":  {"application/ld+json"},
	}

	resp, err := failsafehttp.NewRequest(req, s.client, s.policies...).Do()
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Err(err).Msg("failed to close response body")
		}
	}()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		log.Debug().Str("request_body", string(requestBody)).Msg("request body that caused error response")

		return "", &StatusError{
			StatusCode: resp.StatusCode,
			Body:       responseBody,
		}
	}

	actionRequestURL := resp.Header.Get("Location")

	log.Debug().
		Str("location_header", actionRequestURL).
		Msg("Received action request response from NE-ONE server")

	if s.validateObjectCreation {
		s.waitForObjectAvailability(ctx, baseURL, auth, actionRequestURL)
	}

	return actionRequestURL, nil
}

// waitForObjectAvailability performs a GET request to verify that a newly created
// logistics object is available and consistent in the NE-ONE server before
// subsequent operations. This ensures proper synchronization and prevents
// potential race conditions with rapid successive requests.
func (s *Client) waitForObjectAvailability(cxt context.Context, baseURL, auth, logisticsObjectURL string) {
	if !strings.HasPrefix(logisticsObjectURL, baseURL) {
		log.Error().Msg("Location header in response from NE-ONE Server contains unexpected base URL. " +
			"Skipping availability verification.")

		return
	}

	req, err := http.NewRequestWithContext(cxt, http.MethodGet, logisticsObjectURL, nil)
	if err != nil {
		log.Err(err).Msg("failed to create request for availability verification")

		return
	}

	req.Header = map[string][]string{
		"Accept":        {"application/ld+json"},
		"Authorization": {auth},
	}

	log.Debug().Str("URL", logisticsObjectURL).
		Msg("verifying object availability in the NE-ONE Server")

	resp, err := failsafehttp.NewRequest(req, s.client, s.policies...).Do()
	if err != nil {
		log.Err(err).Msg("failed to verify object availability")

		return
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Err(err).Msg("failed to close response body of availability verification request")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Err(err).Msg("failed to read response body of availability verification request")

			return
		}

		log.Error().Int("StatusCode", resp.StatusCode).Str("Body",
			string(body)).Msg("object availability verification returned error status code")

		return
	}

	log.Debug().Msg("object availability verified successfully.")
}

// GetLogisticsObjectEmbeddedIDs returns a list including the ID of the given Logistic Object
// and the IDs of all embedded objects.
func (s *Client) GetLogisticsObjectEmbeddedIDs(
	cxt context.Context,
	baseURL string,
	auth string,
	logisticsObjectURL string,
) ([]string, error) {
	if !strings.HasPrefix(logisticsObjectURL, baseURL) {
		return nil, fmt.Errorf("unexpected base URL in logistics object URL: %s", logisticsObjectURL)
	}

	targetURL, err := urlpkg.Parse(logisticsObjectURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse logistics object URL: %w", err)
	}

	params := urlpkg.Values{}
	params.Set("embedded", "true")
	targetURL.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(cxt, http.MethodGet, targetURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for embedded IDs: %w", err)
	}

	req.Header = map[string][]string{
		"Accept":        {"application/ld+json"},
		"Authorization": {auth},
	}

	resp, err := failsafehttp.NewRequest(req, s.client, s.policies...).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to send request for embedded IDs: %w", err)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Err(err).Msg("failed to close response body")
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body for embedded IDs: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &StatusError{
			StatusCode: resp.StatusCode,
			Body:       body,
		}
	}

	ids, err := parseEmbeddedIDs(baseURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedded IDs: %w", err)
	}

	return ids, nil
}

type graphResponse struct {
	Graph []idObject `json:"@graph"`
}

type idObject struct {
	ID string `json:"@id"`
}

func parseEmbeddedIDs(baseURL string, body []byte) ([]string, error) {
	var graphResp graphResponse

	err := json.Unmarshal(body, &graphResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	ids := make([]string, 0, len(graphResp.Graph))
	for _, obj := range graphResp.Graph {
		if strings.HasPrefix(obj.ID, baseURL) {
			ids = append(ids, obj.ID)

			continue
		}

		if !strings.HasPrefix(obj.ID, "neone:") {
			log.Warn().
				Str("id", obj.ID).
				Str("baseURL", baseURL).
				Msg("skipping ID with unexpected base URL")
		}
	}

	return ids, nil
}
