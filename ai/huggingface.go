package ai

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "strings"
)

type HFRequest struct {
    Inputs string `json:"inputs"`
}

func AnalyzeWithHFCodeReview(diff, title string) (string, error) {
    apiKey := os.Getenv("HF_API_KEY")
    if apiKey == "" {
        return "", fmt.Errorf("Hugging Face API key missing")
    }

    // Truncate diff if too long
    if len(diff) > 8000 {
        diff = diff[:8000] + "\n... (truncated)"
    }

    // Construct prompt similar to your Gemini version
    prompt := fmt.Sprintf(`You are CodeSage AI, a helpful code review assistant.

PR Title: %s

Code Changes:
%s

Please provide:
- Summary
- What's Good
- Potential Issues
- Security Considerations
- Performance & Best Practices
- Suggestions`, title, diff)

    modelID := "meta-llama/CodeLlama-7b-Instruct-hf"
    url := "https://api-inference.huggingface.co/models/" + modelID

    reqBody := HFRequest{Inputs: prompt}
    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return "", fmt.Errorf("failed to marshal request: %v", err)
    }

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return "", fmt.Errorf("failed to build request: %v", err)
    }
    req.Header.Set("Authorization", "Bearer "+apiKey)
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", fmt.Errorf("Hugging Face API call failed: %v", err)
    }
    defer resp.Body.Close()

    bodyBytes, _ := ioutil.ReadAll(resp.Body)
    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("HF API error: %d - %s", resp.StatusCode, string(bodyBytes))
    }

    result := string(bodyBytes)
    return strings.TrimSpace(result), nil
}
