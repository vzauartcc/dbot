package zauapi

import (
	"encoding/json"
	"time"

	"net/http"
	httpTest "net/http/httptest"
	"strings"
	"testing"

	"github.com/cristalhq/jwt/v5"
)

func initClient(t *testing.T, srvURL string) {
	t.Setenv("ZAU_API_URL", srvURL)
	t.Setenv("ZAU_API_KEY", "testkey")
	Init()
}

// Test successful GetUsers call with correct headers and JSON decoding.
func TestGetUsersSuccess(t *testing.T) {
	var capturedReq *http.Request

	srv := httpTest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		capturedReq = req
		users := []User{{CID: 1, FirstName: "John", LastName: "Doe", DiscordID: "123"}}
		b, _ := json.Marshal(users)

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write(b)
	}))
	defer srv.Close()

	initClient(t, srv.URL)

	client := GetClient()

	got, err := client.GetUsers()
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}

	if len(got) == 0 || got[0].FirstName != "John" {
		t.Errorf("expected user John Doe, got %+v\n", got)
	}

	if capturedReq.Method != http.MethodGet {
		t.Fatalf("method %s not GET\n", capturedReq.Method)
	}

	expectedPath := "/discord/bot/users"
	if capturedReq.URL.Path != expectedPath {
		t.Errorf("expected path %s, got %s\n", expectedPath, capturedReq.URL.Path)
	}

	auth := strings.ReplaceAll(capturedReq.Header.Get("Authorization"), "Bearer ", "")

	validateJWT(t, "testkey", auth)
}

// Test that non-200 status returns ErrStatusCode error with code.
func TestGetUsersBadStatus(t *testing.T) {
	srv := httpTest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	initClient(t, srv.URL)

	_, err := GetClient().GetUsers()
	if err == nil || !strings.Contains(err.Error(), "api returned bad status") {
		t.Fatalf("expected ErrStatusCode error, got %v\n", err)
	}
}

// Test that malformed JSON returns ErrDecoding.
func TestGetUsersBadJSON(t *testing.T) {
	srv := httpTest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("{invalid}"))
	}))
	defer srv.Close()

	initClient(t, srv.URL)

	_, err := GetClient().GetUsers()
	if err == nil || !strings.Contains(err.Error(), "api returned invalid json") {
		t.Fatalf("expected ErrDecoding error, got %v\n", err)
	}
}

// Test GetUserByID success.
func TestGetUserByIDSuccess(t *testing.T) {
	var capturedReq *http.Request

	srv := httpTest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		capturedReq = req
		user := User{CID: 2, FirstName: "Alice", LastName: "Smith", DiscordID: "456"}
		b, _ := json.Marshal(user)

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write(b)
	}))
	defer srv.Close()

	initClient(t, srv.URL)

	got, err := GetClient().GetUserByID("2")
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}

	if got.FirstName != "Alice" || got.DiscordID != "456" {
		t.Errorf("expected Alice Smith got %+v\n", got)
	}

	expectedPath := "/discord/bot/user/2"
	if capturedReq.URL.Path != expectedPath {
		t.Fatalf("path %s not expected %s\n", capturedReq.URL.Path, expectedPath)
	}
}

// Test GetUserByID non-200.
func TestGetUserByIDBadStatus(t *testing.T) {
	srv := httpTest.NewServer(
		http.HandlerFunc(
			func(writer http.ResponseWriter, _ *http.Request) { writer.WriteHeader(http.StatusNotFound) },
		),
	)
	defer srv.Close()

	initClient(t, srv.URL)

	_, err := GetClient().GetUserByID("2")
	if err == nil || !strings.Contains(err.Error(), "api returned bad status") {
		t.Fatalf("expected ErrStatusCode, got %v\n", err)
	}
}

// Test GetIronMic success.
func TestGetIronMicSuccess(t *testing.T) {
	var capturedReq *http.Request

	srv := httpTest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		capturedReq = req
		resp := IronMicResponse{Results: IronMicResult{}, Period: ActivityPeriod{}}
		b, _ := json.Marshal(resp)

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write(b)
	}))
	defer srv.Close()

	initClient(t, srv.URL)

	_, err := GetClient().GetIronMic()
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}

	expectedPath := "/discord/bot/ironmic"
	if capturedReq == nil || capturedReq.URL.Path != expectedPath {
		t.Errorf("path %s expected %s", capturedReq.URL.Path, expectedPath)
	}
}

// Test GetOnlineATC success.
func TestGetOnlineATCSuccess(t *testing.T) {
	var capturedReq *http.Request

	srv := httpTest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		capturedReq = req
		resp := OnlineData{Pilots: nil, Controllers: []OnlineController{{CID: 1}}}
		b, _ := json.Marshal(resp)

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write(b)
	}))
	defer srv.Close()

	initClient(t, srv.URL)

	_, err := GetClient().GetOnlineATC()
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}

	expectedPath := "/online"
	if capturedReq == nil || capturedReq.URL.Path != expectedPath {
		t.Errorf("path %s expected %s\n", capturedReq.URL.Path, expectedPath)
	}
}

// Test GetStaff success.
func TestGetStaffSuccess(t *testing.T) {
	var capturedReq *http.Request

	srv := httpTest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		capturedReq = req
		resp := Staff{ATM: StaffPosition{Title: "admin", Code: "A"}}
		b, _ := json.Marshal(resp)

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write(b)
	}))
	defer srv.Close()

	initClient(t, srv.URL)

	_, err := GetClient().GetStaff()
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}

	expectedPath := "/controller/staff"
	if capturedReq == nil || capturedReq.URL.Path != expectedPath {
		t.Errorf("path %s expected %s\n", capturedReq.URL.Path, expectedPath)
	}
}

// Test JWT generation.
func TestGenerateJWT(t *testing.T) {
	testKey := "super-secret-test-key"
	t.Setenv("ZAU_API_KEY", testKey)

	tokenStr := generateJWT()

	if tokenStr == "" {
		t.Fatal("Expected token string, got empty string\n")
	}

	validateJWT(t, testKey, tokenStr)
}

func validateJWT(t *testing.T, key, tokenStr string) {
	verifier, err := jwt.NewSignerHS(jwt.HS256, []byte(key))
	if err != nil {
		t.Fatalf("Failed to create verifier: %v\n", err)
	}

	token, err := jwt.Parse([]byte(tokenStr), verifier)
	if err != nil {
		t.Fatalf("Failed to parse/verify token: %v\n", err)
	}

	var claims jwt.RegisteredClaims

	err = token.DecodeClaims(&claims)
	if err != nil {
		t.Fatalf("Failed to decode claims: %v\n", err)
	}

	if claims.Subject != "dbot" {
		t.Errorf("Expected subject 'dbot', got '%s'\n", claims.Subject)
	}

	if !claims.IsValidAt(time.Now()) {
		t.Error("Token is already expired or invalid\n")
	}
}
