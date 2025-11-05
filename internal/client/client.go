// Copyright Starburst Data, Inc. All rights reserved.
//
// The source code is the proprietary and confidential information of Starburst Data, Inc. and
// may be used only for reference purposes in connection with the Terraform Registry. All rights,
// title, interest and ownership of the code and any derivatives, updates, upgrades, enhancements
// and modifications thereof remain with Starburst Data, Inc. You are not permitted to distribute,
// disclose, sell, lease, transfer, assign, modify, create derivative works of, or sublicense the
// code, or use the code to create or develop any products or services.

package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type GalaxyClient struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	HTTPClient   *http.Client

	tokenMu     sync.RWMutex
	accessToken string
	tokenExpiry time.Time
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func NewGalaxyClient(baseURL, clientID, clientSecret string) *GalaxyClient {
	return &GalaxyClient{
		BaseURL:      strings.TrimSuffix(baseURL, "/"),
		ClientID:     clientID,
		ClientSecret: clientSecret,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *GalaxyClient) getAccessToken(ctx context.Context) error {
	credentials := base64.StdEncoding.EncodeToString([]byte(c.ClientID + ":" + c.ClientSecret))

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/oauth/v2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+credentials)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request token: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log the close error but don't override the main error
			tflog.Warn(ctx, "Failed to close response body", map[string]interface{}{
				"error": closeErr.Error(),
			})
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	c.tokenMu.Lock()
	c.accessToken = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second) // Subtract 60s for buffer
	c.tokenMu.Unlock()

	return nil
}

func (c *GalaxyClient) ensureValidToken(ctx context.Context) error {
	c.tokenMu.RLock()
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		c.tokenMu.RUnlock()
		return nil
	}
	c.tokenMu.RUnlock()

	return c.getAccessToken(ctx)
}

func (c *GalaxyClient) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	return c.doRequestWithRetry(ctx, method, path, body, result, 3)
}

func (c *GalaxyClient) doRequestWithRetry(ctx context.Context, method, path string, body interface{}, result interface{}, retries int) error {
	if err := c.ensureValidToken(ctx); err != nil {
		return err
	}

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	fullURL := c.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.tokenMu.RLock()
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	c.tokenMu.RUnlock()

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log the close error but don't override the main error
			tflog.Warn(ctx, "Failed to close response body", map[string]interface{}{
				"error": closeErr.Error(),
			})
		}
	}()

	// Handle token expiry
	if (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden) && retries > 0 {
		c.tokenMu.Lock()
		c.accessToken = ""
		c.tokenMu.Unlock()

		// Exponential backoff for auth failures
		time.Sleep(time.Duration(4-retries) * time.Second)
		return c.doRequestWithRetry(ctx, method, path, body, result, retries-1)
	}

	// Handle rate limiting (429 Too Many Requests)
	if resp.StatusCode == http.StatusTooManyRequests && retries > 0 {
		// Extract wait time from response body or use exponential backoff
		waitTime := time.Duration(4-retries) * 15 * time.Second // More aggressive backoff for rate limits
		tflog.Info(ctx, "Rate limit encountered, retrying after backoff", map[string]interface{}{
			"wait_time":    waitTime.String(),
			"retries_left": retries - 1,
			"endpoint":     path,
			"method":       method,
		})
		time.Sleep(waitTime)
		return c.doRequestWithRetry(ctx, method, path, body, result, retries-1)
	}

	if resp.StatusCode == http.StatusNotFound {
		return &NotFoundError{Message: fmt.Sprintf("resource not found: %s", path)}
	}

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// NotFoundError represents a 404 response
type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

func IsNotFound(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}

// GetAllPaginatedResults fetches all paginated results from an API endpoint
// This should be used by data sources to automatically handle pagination
func (c *GalaxyClient) GetAllPaginatedResults(ctx context.Context, path string) ([]interface{}, error) {
	var allResults []interface{}
	pageToken := ""

	for {
		// Build URL with pagination parameters
		requestPath := path
		if pageToken != "" {
			separator := "?"
			if strings.Contains(path, "?") {
				separator = "&"
			}
			requestPath = fmt.Sprintf("%s%spageToken=%s", path, separator, url.QueryEscape(pageToken))
		}

		var response map[string]interface{}
		err := c.doRequest(ctx, "GET", requestPath, nil, &response)
		if err != nil {
			return nil, err
		}

		// Extract results from the response (API uses "result" field for paginated responses)
		if result, ok := response["result"].([]interface{}); ok {
			allResults = append(allResults, result...)
		}

		// Check if there's a next page
		if nextToken, ok := response["nextPageToken"].(string); ok && nextToken != "" {
			pageToken = nextToken
		} else {
			break // No more pages
		}
	}

	return allResults, nil
}

// Resource-specific methods

func (c *GalaxyClient) CreateCluster(ctx context.Context, cluster interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "POST", "/public/api/v1/cluster", cluster, &result)
	return result, err
}

func (c *GalaxyClient) GetCluster(ctx context.Context, clusterID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/cluster/"+clusterID, nil, &result)
	return result, err
}

func (c *GalaxyClient) UpdateCluster(ctx context.Context, clusterID string, cluster interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "PATCH", "/public/api/v1/cluster/"+clusterID, cluster, &result)
	return result, err
}

func (c *GalaxyClient) DeleteCluster(ctx context.Context, clusterID string) error {
	return c.doRequest(ctx, "DELETE", "/public/api/v1/cluster/"+clusterID, nil, nil)
}

func (c *GalaxyClient) CreateUser(ctx context.Context, user interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "POST", "/public/api/v1/user", user, &result)
	return result, err
}

func (c *GalaxyClient) GetUser(ctx context.Context, userID string) (map[string]interface{}, error) {
	// Check if this is an email-based lookup
	if strings.HasPrefix(userID, "email=") {
		email := strings.TrimPrefix(userID, "email=")

		// First, get all users to find the user by email
		allUsers, err := c.ListUsers(ctx)
		if err != nil {
			return nil, err
		}

		for _, user := range allUsers {
			if userEmail, ok := user["email"].(string); ok && userEmail == email {
				return user, nil
			}
		}

		// User not found
		return nil, &NotFoundError{Message: fmt.Sprintf("user with email %s not found", email)}
	}

	// Regular ID-based lookup
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/user/"+userID, nil, &result)
	return result, err
}

func (c *GalaxyClient) UpdateUser(ctx context.Context, userID string, user interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "PATCH", "/public/api/v1/user/"+userID, user, &result)
	return result, err
}

func (c *GalaxyClient) CreateRole(ctx context.Context, role interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "POST", "/public/api/v1/role", role, &result)
	return result, err
}

func (c *GalaxyClient) GetRole(ctx context.Context, roleID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/role/"+roleID, nil, &result)
	return result, err
}

func (c *GalaxyClient) DeleteRole(ctx context.Context, roleID string) error {
	return c.doRequest(ctx, "DELETE", "/public/api/v1/role/"+roleID, nil, nil)
}

func (c *GalaxyClient) CreateServiceAccount(ctx context.Context, account interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "POST", "/public/api/v1/serviceAccount", account, &result)
	return result, err
}

func (c *GalaxyClient) GetServiceAccount(ctx context.Context, accountID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/serviceAccount/"+accountID, nil, &result)
	return result, err
}

func (c *GalaxyClient) UpdateServiceAccount(ctx context.Context, accountID string, account interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "PATCH", "/public/api/v1/serviceAccount/"+accountID, account, &result)
	return result, err
}

func (c *GalaxyClient) DeleteServiceAccount(ctx context.Context, accountID string) error {
	return c.doRequest(ctx, "DELETE", "/public/api/v1/serviceAccount/"+accountID, nil, nil)
}

func (c *GalaxyClient) CreateServiceAccountPassword(ctx context.Context, accountID string, password interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "POST", "/public/api/v1/serviceAccount/"+accountID+"/serviceAccountPassword", password, &result)
	return result, err
}

func (c *GalaxyClient) GetServiceAccountPassword(ctx context.Context, accountID, passwordID string) (map[string]interface{}, error) {
	// The dedicated password GET endpoint is broken on the server side, so we need to
	// get the service account and extract the specific password from the passwords array
	var serviceAccount map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/serviceAccount/"+accountID, nil, &serviceAccount)
	if err != nil {
		return nil, err
	}

	// Extract passwords array
	passwordsRaw, ok := serviceAccount["passwords"]
	if !ok {
		return nil, &NotFoundError{Message: fmt.Sprintf("passwords array not found in service account %s", accountID)}
	}

	passwords, ok := passwordsRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("passwords field is not an array in service account %s", accountID)
	}

	// Find the specific password by ID
	for _, passwordRaw := range passwords {
		password, ok := passwordRaw.(map[string]interface{})
		if !ok {
			continue
		}

		// Check both possible ID field names
		pwdID, ok1 := password["serviceAccountPasswordId"].(string)
		if !ok1 {
			pwdID, ok1 = password["id"].(string)
		}

		if ok1 && pwdID == passwordID {
			return password, nil
		}
	}

	return nil, &NotFoundError{Message: fmt.Sprintf("password %s not found in service account %s", passwordID, accountID)}
}

func (c *GalaxyClient) DeleteServiceAccountPassword(ctx context.Context, accountID, passwordID string) error {
	return c.doRequest(ctx, "DELETE", "/public/api/v1/serviceAccount/"+accountID+"/serviceAccountPassword/"+passwordID, nil, nil)
}

// Catalog methods - for all catalog types
func (c *GalaxyClient) CreateCatalog(ctx context.Context, catalogType string, catalog interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := fmt.Sprintf("/public/api/v1/catalogType/%s/catalog", catalogType)
	err := c.doRequest(ctx, "POST", path, catalog, &result)
	return result, err
}

func (c *GalaxyClient) GetCatalog(ctx context.Context, catalogType, catalogID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := fmt.Sprintf("/public/api/v1/catalogType/%s/catalog/%s", catalogType, catalogID)
	err := c.doRequest(ctx, "GET", path, nil, &result)
	return result, err
}

func (c *GalaxyClient) UpdateCatalog(ctx context.Context, catalogType, catalogID string, catalog interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := fmt.Sprintf("/public/api/v1/catalogType/%s/catalog/%s", catalogType, catalogID)
	err := c.doRequest(ctx, "PATCH", path, catalog, &result)
	return result, err
}

func (c *GalaxyClient) DeleteCatalog(ctx context.Context, catalogType, catalogID string) error {
	path := fmt.Sprintf("/public/api/v1/catalogType/%s/catalog/%s", catalogType, catalogID)
	return c.doRequest(ctx, "DELETE", path, nil, nil)
}

// Data Product methods
func (c *GalaxyClient) CreateDataProduct(ctx context.Context, product interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "POST", "/public/api/v1/dataProduct", product, &result)
	return result, err
}

func (c *GalaxyClient) GetDataProduct(ctx context.Context, productID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/dataProduct/"+productID, nil, &result)
	return result, err
}

func (c *GalaxyClient) UpdateDataProduct(ctx context.Context, productID string, product interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "PATCH", "/public/api/v1/dataProduct/"+productID, product, &result)
	return result, err
}

func (c *GalaxyClient) DeleteDataProduct(ctx context.Context, productID string) error {
	return c.doRequest(ctx, "DELETE", "/public/api/v1/dataProduct/"+productID, nil, nil)
}

// List operations
func (c *GalaxyClient) ListUsers(ctx context.Context) ([]map[string]interface{}, error) {
	var result struct {
		Result []map[string]interface{} `json:"result"`
	}
	err := c.doRequest(ctx, "GET", "/public/api/v1/user", nil, &result)
	return result.Result, err
}

func (c *GalaxyClient) ListClusters(ctx context.Context) ([]map[string]interface{}, error) {
	var result struct {
		Clusters []map[string]interface{} `json:"clusters"`
	}
	err := c.doRequest(ctx, "GET", "/public/api/v1/cluster", nil, &result)
	return result.Clusters, err
}

func (c *GalaxyClient) ListRoles(ctx context.Context) ([]map[string]interface{}, error) {
	var result struct {
		Roles []map[string]interface{} `json:"roles"`
	}
	err := c.doRequest(ctx, "GET", "/public/api/v1/role", nil, &result)
	return result.Roles, err
}

func (c *GalaxyClient) ListServiceAccounts(ctx context.Context) ([]map[string]interface{}, error) {
	var result struct {
		ServiceAccounts []map[string]interface{} `json:"serviceAccounts"`
	}
	err := c.doRequest(ctx, "GET", "/public/api/v1/serviceAccount", nil, &result)
	return result.ServiceAccounts, err
}

func (c *GalaxyClient) ListDataProducts(ctx context.Context) ([]map[string]interface{}, error) {
	var result struct {
		DataProducts []map[string]interface{} `json:"dataProducts"`
	}
	err := c.doRequest(ctx, "GET", "/public/api/v1/dataProduct", nil, &result)
	return result.DataProducts, err
}

// Role Privilege Grant methods
func (c *GalaxyClient) CreateRolePrivilegeGrant(ctx context.Context, grant interface{}) (map[string]interface{}, error) {
	// Extract roleId from the grant request
	grantMap, ok := grant.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("grant must be a map")
	}

	roleId, ok := grantMap["roleId"].(string)
	if !ok || roleId == "" {
		return nil, fmt.Errorf("roleId is required")
	}

	// Remove roleId from request body as it's in the URL
	delete(grantMap, "roleId")

	var result map[string]interface{}
	err := c.doRequest(ctx, "POST", "/public/api/v1/role/"+roleId+"/privilege:grant", grantMap, &result)
	return result, err
}

func (c *GalaxyClient) GetRolePrivilegeGrant(ctx context.Context, grantID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/rolePrivilegeGrant/"+grantID, nil, &result)
	return result, err
}

func (c *GalaxyClient) DeleteRolePrivilegeGrant(ctx context.Context, grantID string) error {
	return c.doRequest(ctx, "DELETE", "/public/api/v1/rolePrivilegeGrant/"+grantID, nil, nil)
}

// Column Mask methods
func (c *GalaxyClient) CreateColumnMask(ctx context.Context, mask interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "POST", "/public/api/v1/columnMask", mask, &result)
	return result, err
}

func (c *GalaxyClient) GetColumnMask(ctx context.Context, maskID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/columnMask/"+maskID, nil, &result)
	return result, err
}

func (c *GalaxyClient) UpdateColumnMask(ctx context.Context, maskID string, mask interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "PATCH", "/public/api/v1/columnMask/"+maskID, mask, &result)
	return result, err
}

func (c *GalaxyClient) DeleteColumnMask(ctx context.Context, maskID string) error {
	return c.doRequest(ctx, "DELETE", "/public/api/v1/columnMask/"+maskID, nil, nil)
}

func (c *GalaxyClient) ListColumnMasks(ctx context.Context) ([]map[string]interface{}, error) {
	var result struct {
		ColumnMasks []map[string]interface{} `json:"columnMasks"`
	}
	err := c.doRequest(ctx, "GET", "/public/api/v1/columnMask", nil, &result)
	return result.ColumnMasks, err
}

// Row Filter methods
func (c *GalaxyClient) CreateRowFilter(ctx context.Context, filter interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "POST", "/public/api/v1/rowFilter", filter, &result)
	return result, err
}

func (c *GalaxyClient) GetRowFilter(ctx context.Context, filterID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/rowFilter/"+filterID, nil, &result)
	return result, err
}

func (c *GalaxyClient) UpdateRowFilter(ctx context.Context, filterID string, filter interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "PATCH", "/public/api/v1/rowFilter/"+filterID, filter, &result)
	return result, err
}

func (c *GalaxyClient) DeleteRowFilter(ctx context.Context, filterID string) error {
	return c.doRequest(ctx, "DELETE", "/public/api/v1/rowFilter/"+filterID, nil, nil)
}

// Tag methods
func (c *GalaxyClient) CreateTag(ctx context.Context, tag interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "POST", "/public/api/v1/tag", tag, &result)
	return result, err
}

func (c *GalaxyClient) GetTag(ctx context.Context, tagID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/tag/"+tagID, nil, &result)
	return result, err
}

func (c *GalaxyClient) UpdateTag(ctx context.Context, tagID string, tag interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "PATCH", "/public/api/v1/tag/"+tagID, tag, &result)
	return result, err
}

func (c *GalaxyClient) DeleteTag(ctx context.Context, tagID string) error {
	return c.doRequest(ctx, "DELETE", "/public/api/v1/tag/"+tagID, nil, nil)
}

// Catalog validation methods
func (c *GalaxyClient) ValidateCatalog(ctx context.Context, catalogType, catalogID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := fmt.Sprintf("/public/api/v1/catalogType/%s/catalog/%s/validate", catalogType, catalogID)
	err := c.doRequest(ctx, "GET", path, nil, &result)
	return result, err
}

// Catalog metadata - get catalog by ID from the list of all catalogs
func (c *GalaxyClient) GetCatalogMetadata(ctx context.Context, catalogID string) (map[string]interface{}, error) {
	// Check if this is a name-based lookup
	if strings.HasPrefix(catalogID, "name=") {
		catalogName := strings.TrimPrefix(catalogID, "name=")

		// First, find the catalog by name to get its actual ID
		allCatalogs, err := c.GetAllPaginatedResults(ctx, "/public/api/v1/catalog")
		if err != nil {
			return nil, err
		}

		var actualCatalogID string
		for _, catalogInterface := range allCatalogs {
			if catalogMap, ok := catalogInterface.(map[string]interface{}); ok {
				if name, ok := catalogMap["catalogName"].(string); ok && name == catalogName {
					if id, ok := catalogMap["catalogId"].(string); ok {
						actualCatalogID = id
						break
					}
				}
			}
		}

		if actualCatalogID == "" {
			return nil, fmt.Errorf("resource not found: catalog with name %s not found", catalogName)
		}

		// Now get the metadata using the actual catalog ID
		catalogID = actualCatalogID
	}

	// Call the catalog metadata API endpoint directly
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", fmt.Sprintf("/public/api/v1/catalog/%s/catalogMetadata", catalogID), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("resource not found: catalog with ID %s not found", catalogID)
	}

	return result, nil
}

// List methods for catalogs
func (c *GalaxyClient) ListCatalogs(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/catalog", nil, &result)
	return result, err
}

func (c *GalaxyClient) ListBigqueryCatalogs(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/catalogType/bigquery/catalog", nil, &result)
	return result, err
}

func (c *GalaxyClient) ListCassandraCatalogs(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/catalogType/cassandra/catalog", nil, &result)
	return result, err
}

func (c *GalaxyClient) ListGcsCatalogs(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/catalogType/gcs/catalog", nil, &result)
	return result, err
}

func (c *GalaxyClient) ListMongodbCatalogs(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/catalogType/mongodb/catalog", nil, &result)
	return result, err
}

func (c *GalaxyClient) ListMysqlCatalogs(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/catalogType/mysql/catalog", nil, &result)
	return result, err
}

func (c *GalaxyClient) ListOpensearchCatalogs(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/catalogType/opensearch/catalog", nil, &result)
	return result, err
}

func (c *GalaxyClient) ListPostgresqlCatalogs(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/catalogType/postgresql/catalog", nil, &result)
	return result, err
}

func (c *GalaxyClient) ListRedshiftCatalogs(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/catalogType/redshift/catalog", nil, &result)
	return result, err
}

func (c *GalaxyClient) ListS3Catalogs(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/catalogType/s3/catalog", nil, &result)
	return result, err
}

func (c *GalaxyClient) ListSnowflakeCatalogs(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/catalogType/snowflake/catalog", nil, &result)
	return result, err
}

func (c *GalaxyClient) ListSqlserverCatalogs(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/catalogType/sqlserver/catalog", nil, &result)
	return result, err
}

// Cross Account IAM Role methods
func (c *GalaxyClient) CreateCrossAccountIamRole(ctx context.Context, role interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "POST", "/public/api/v1/crossAccountIamRole", role, &result)
	return result, err
}

func (c *GalaxyClient) GetCrossAccountIamRole(ctx context.Context, aliasName string) (map[string]interface{}, error) {
	// Cross-account IAM roles API doesn't support getting individual roles
	// We need to list all roles and find the one with matching alias name
	allRoles, err := c.ListCrossAccountIamRoles(ctx)
	if err != nil {
		return nil, err
	}

	// Look for the role with matching alias name in the results
	if result, ok := allRoles["result"].([]interface{}); ok {
		for _, roleInterface := range result {
			if roleMap, ok := roleInterface.(map[string]interface{}); ok {
				if roleAliasName, exists := roleMap["aliasName"].(string); exists && roleAliasName == aliasName {
					return roleMap, nil
				}
			}
		}
	}

	// Role not found
	return nil, &NotFoundError{Message: "cross-account IAM role '" + aliasName + "' not found"}
}

func (c *GalaxyClient) UpdateCrossAccountIamRole(ctx context.Context, roleID string, role interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "PUT", "/public/api/v1/crossAccountIamRole/"+roleID, role, &result)
	return result, err
}

func (c *GalaxyClient) DeleteCrossAccountIamRole(ctx context.Context, awsIamArn string) error {
	// URL encode the ARN since it contains special characters like : and /
	encodedArn := url.QueryEscape(awsIamArn)
	return c.doRequest(ctx, "DELETE", "/public/api/v1/crossAccountIamRole/"+encodedArn, nil, nil)
}

// List methods for other resources
func (c *GalaxyClient) ListCrossAccountIamRoles(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/crossAccountIamRole", nil, &result)
	return result, err
}

func (c *GalaxyClient) ListRoleGrants(ctx context.Context, roleID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/role/"+roleID+"/rolegrant", nil, &result)
	return result, err
}

func (c *GalaxyClient) ListRolePrivileges(ctx context.Context, roleID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/role/"+roleID+"/privileges", nil, &result)
	return result, err
}

func (c *GalaxyClient) ListTags(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/tag", nil, &result)
	return result, err
}

func (c *GalaxyClient) ListPolicies(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/policy", nil, &result)
	return result, err
}

func (c *GalaxyClient) ListRowFilters(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/policy/rowFilter", nil, &result)
	return result, err
}

// Policy methods
func (c *GalaxyClient) GetPolicy(ctx context.Context, policyID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/policy/"+policyID, nil, &result)
	return result, err
}

func (c *GalaxyClient) CreatePolicy(ctx context.Context, policy interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "POST", "/public/api/v1/policy", policy, &result)
	return result, err
}

func (c *GalaxyClient) UpdatePolicy(ctx context.Context, policyID string, policy interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "PATCH", "/public/api/v1/policy/"+policyID, policy, &result)
	return result, err
}

func (c *GalaxyClient) DeletePolicy(ctx context.Context, policyID string) error {
	return c.doRequest(ctx, "DELETE", "/public/api/v1/policy/"+policyID, nil, nil)
}

// UpdateRolePrivilegeGrant updates an existing role privilege grant
func (c *GalaxyClient) UpdateRolePrivilegeGrant(ctx context.Context, grantID string, grant interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "PUT", "/public/api/v1/role/privilege/grant/"+grantID, grant, &result)
	return result, err
}

// UpdateRole updates an existing role
func (c *GalaxyClient) UpdateRole(ctx context.Context, roleID string, role interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "PATCH", "/public/api/v1/role/"+roleID, role, &result)
	return result, err
}

// UpdateServiceAccountPassword updates a service account password
func (c *GalaxyClient) UpdateServiceAccountPassword(ctx context.Context, passwordID string, password interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "PUT", "/public/api/v1/serviceAccount/password/"+passwordID, password, &result)
	return result, err
}

// SQL Job methods
func (c *GalaxyClient) CreateSqlJob(ctx context.Context, sqlJob interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "POST", "/public/api/v1/sqlJob", sqlJob, &result)
	return result, err
}

func (c *GalaxyClient) GetSqlJob(ctx context.Context, sqlJobID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/sqlJob/"+sqlJobID, nil, &result)
	return result, err
}

func (c *GalaxyClient) UpdateSqlJob(ctx context.Context, sqlJobID string, sqlJob interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "PATCH", "/public/api/v1/sqlJob/"+sqlJobID, sqlJob, &result)
	return result, err
}

func (c *GalaxyClient) DeleteSqlJob(ctx context.Context, sqlJobID string) error {
	return c.doRequest(ctx, "DELETE", "/public/api/v1/sqlJob/"+sqlJobID, nil, nil)
}

// SQL Job Status data source
func (c *GalaxyClient) GetSqlJobStatus(ctx context.Context, sqlJobID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/sqlJob/"+sqlJobID+"/sqlJobStatus", nil, &result)
	return result, err
}

// SQL Job History data source
func (c *GalaxyClient) GetSqlJobHistory(ctx context.Context, sqlJobID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/sqlJob/"+sqlJobID+"/sqlJobHistory", nil, &result)
	return result, err
}

// List SQL Jobs data source
func (c *GalaxyClient) ListSqlJobs(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/sqlJob", nil, &result)
	return result, err
}

// Privatelink data sources
func (c *GalaxyClient) GetPrivatelink(ctx context.Context, privatelinkID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/privatelink/"+privatelinkID, nil, &result)
	return result, err
}

func (c *GalaxyClient) ListPrivatelinks(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/privatelink", nil, &result)
	return result, err
}

// Data Quality data sources
func (c *GalaxyClient) GetDataQualitySummary(ctx context.Context, catalogID, schemaID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := fmt.Sprintf("/public/api/v1/catalog/%s/schema/%s/dataQualitySummary", catalogID, schemaID)
	err := c.doRequest(ctx, "GET", path, nil, &result)
	return result, err
}

func (c *GalaxyClient) ListDataQualitySummaries(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/dataQualitySummary", nil, &result)
	return result, err
}

// Table data source
func (c *GalaxyClient) ListTables(ctx context.Context, catalogID, schemaID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := fmt.Sprintf("/public/api/v1/catalog/%s/schema/%s/table", catalogID, schemaID)
	err := c.doRequest(ctx, "GET", path, nil, &result)
	return result, err
}

// Schema data source
func (c *GalaxyClient) ListSchemas(ctx context.Context, catalogID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := fmt.Sprintf("/public/api/v1/catalog/%s/schema", catalogID)
	err := c.doRequest(ctx, "GET", path, nil, &result)
	return result, err
}

// Column data source
func (c *GalaxyClient) ListColumns(ctx context.Context, catalogID, schemaID, tableID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := fmt.Sprintf("/public/api/v1/catalog/%s/schema/%s/table/%s/column", catalogID, schemaID, tableID)
	err := c.doRequest(ctx, "GET", path, nil, &result)
	return result, err
}

// Cross Account IAM Role Metadatas data source
func (c *GalaxyClient) ListCrossAccountIamRoleMetadatas(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, "GET", "/public/api/v1/crossAccountIamRoleMetadata", nil, &result)
	return result, err
}
