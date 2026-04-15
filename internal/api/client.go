package zauapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/cristalhq/jwt/v5"
)

var (
	instance               *Client
	ErrCreateRequestFailed = errors.New("create request failed")
	ErrCallFailed          = errors.New("api call failed")
	ErrStatusCode          = errors.New("api returned bad status")
	ErrDecoding            = errors.New("api returned invalid json")
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func Init() {
	baseURL := os.Getenv("ZAU_API_URL")

	u, err := url.Parse(baseURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		log.Panicf("Invalid API URL: %v\n", err)
	}

	instance = &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
	}
}

func GetClient() *Client {
	return instance
}

func generateRequest[T any](
	method string,
	url string,
	body io.Reader,
) (T, error) {
	var result T

	req, err := http.NewRequestWithContext(
		context.Background(),
		method,
		instance.baseURL+url,
		body,
	)
	if err != nil {
		return result, fmt.Errorf("%w: %w", ErrCreateRequestFailed, err)
	}

	req.Header.Set("Authorization", "Bearer "+generateJWT())

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Dbot/1.0")
	req.Header.Set("Content-Type", "application/json")

	resp, err := instance.httpClient.Do(req)
	if err != nil {
		return result, fmt.Errorf("%w: %w", ErrCallFailed, err)
	}

	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("%w: %d", ErrStatusCode, resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return result, fmt.Errorf("%w: %w", ErrDecoding, err)
	}

	return result, nil
}

func generateJWT() string {
	key := []byte(os.Getenv("ZAU_API_KEY"))

	signer, err := jwt.NewSignerHS(jwt.HS256, key)
	if err != nil {
		log.Printf("Error generate JWT: %v\n", err)
		return ""
	}

	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Second)),
		Subject:   "dbot",
	}

	builder := jwt.NewBuilder(signer)

	token, err := builder.Build(claims)
	if err != nil {
		log.Printf("Error signing JWT: %v\n", err)
		return ""
	}

	return token.String()
}

func (c *Client) GetUsers() ([]User, error) {
	return generateRequest[[]User]("GET", "/discord/bot/users", nil)
}

func (c *Client) GetUserByID(id string) (User, error) {
	return generateRequest[User]("GET", "/discord/bot/user/"+id, nil)
}

func (c *Client) GetIronMic() (IronMicResponse, error) {
	return generateRequest[IronMicResponse]("GET", "/discord/bot/ironmic", nil)
}

func (c *Client) GetOnlineATC() (OnlineData, error) {
	return generateRequest[OnlineData]("GET", "/online", nil)
}

func (c *Client) GetStaff() (Staff, error) {
	return generateRequest[Staff]("GET", "/controller/staff", nil)
}

func (c *Client) GetConfig(guildID string) (Config, error) {
	return generateRequest[Config]("GET", "/discord/bot/config/"+guildID, nil)
}

func (c *Client) GetConfigs() ([]Config, error) {
	return generateRequest[[]Config]("GET", "/discord/bot/configs", nil)
}

func (c *Client) UpdateConfig(guildID string, config Config) (Config, error) {
	buf := new(bytes.Buffer)

	err := json.NewEncoder(buf).Encode(config)
	if err != nil {
		return config, err
	}

	return generateRequest[Config]("PATCH", "/discord/bot/config/"+guildID, buf)
}
