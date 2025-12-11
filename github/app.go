package github

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "time"
    "codesage/config"
    jwt "github.com/golang-jwt/jwt/v5"
)

// VerifyWebhookSignature validates X-Hub-Signature-256 using the configured secret
func VerifyWebhookSignature(secret string, body []byte, signatureHeader string) bool {
    if secret == "" || signatureHeader == "" {
        return false
    }
    const prefix = "sha256="
    if !strings.HasPrefix(signatureHeader, prefix) {
        return false
    }
    givenSigHex := signatureHeader[len(prefix):]
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(body)
    expected := mac.Sum(nil)
    given, err := hex.DecodeString(givenSigHex)
    if err != nil {
        return false
    }
    return hmac.Equal(expected, given)
}

// generateAppJWT builds a short-lived JWT used to request installation tokens
func generateAppJWT(cfg *config.Config) (string, error) {
    if cfg.GitHubAppID == "" || cfg.GitHubAppPrivateKey == "" {
        return "", fmt.Errorf("missing GitHub App credentials")
    }
    now := time.Now()
    claims := jwt.MapClaims{
        "iat": now.Add(-time.Minute).Unix(),
        "exp": now.Add(9 * time.Minute).Unix(),
        "iss": cfg.GitHubAppID,
    }
    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(cfg.GitHubAppPrivateKey))
    if err != nil {
        return "", err
    }
    return token.SignedString(privateKey)
}

type installationTokenResponse struct {
    Token     string    `json:"token"`
    ExpiresAt time.Time `json:"expires_at"`
}

// GetInstallationToken exchanges an App JWT for an installation access token
func GetInstallationToken(cfg *config.Config, installationID int64) (string, time.Time, error) {
    appJWT, err := generateAppJWT(cfg)
    if err != nil {
        return "", time.Time{}, err
    }
    url := fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", installationID)
    req, err := http.NewRequest("POST", url, nil)
    if err != nil {
        return "", time.Time{}, err
    }
    req.Header.Set("Authorization", "Bearer "+appJWT)
    req.Header.Set("Accept", "application/vnd.github+json")
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", time.Time{}, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusCreated {
        return "", time.Time{}, fmt.Errorf("failed to get installation token: %s", resp.Status)
    }
    var out installationTokenResponse
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
        return "", time.Time{}, err
    }
    return out.Token, out.ExpiresAt, nil
}



