// ai/gemini.go - Improved version
package ai

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "strings"
)

// Request payload
type GeminiRequest struct {
    Contents []struct {
        Parts []struct {
            Text string `json:"text"`
        } `json:"parts"`
    } `json:"contents"`
}

// Response payload
type GeminiResponse struct {
    Candidates []struct {
        Content struct {
            Parts []struct {
                Text string `json:"text"`
            } `json:"parts"`
        } `json:"content"`
    } `json:"candidates"`
}

// AnalyzeWithGemini sends diff + title to Gemini and returns analysis
func AnalyzeWithGemini(diff, title string) (string, error) {
    apiKey := os.Getenv("GEMINI_API_KEY")
    if apiKey == "" {
        return "", fmt.Errorf("Gemini API key missing")
    }
    
    // Limit diff size to avoid hitting API limits
    if len(diff) > 8000 {
        diff = diff[:8000] + "\n... (truncated for analysis)"
    }
    
  prompt := fmt.Sprintf(`Review this code change like a friendly senior developer:

**%s**

%s

Give me 2-3 key points about this change - what's good, what needs attention, any quick suggestions. Keep it conversational and practical.`, title, diff)  
    // Build request
    reqBody := GeminiRequest{
        Contents: []struct {
            Parts []struct {
                Text string `json:"text"`
            } `json:"parts"`
        }{
            {
                Parts: []struct {
                    Text string `json:"text"`
                }{
                    {Text: prompt},
                },
            },
        },
    }
    
    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return "", fmt.Errorf("failed to marshal request: %v", err)
    }
    
url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-pro:generateContent?key=" + apiKey
    
    resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return "", fmt.Errorf("failed to call Gemini API: %v", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("Gemini API returned status %d", resp.StatusCode)
    }
    
    var geminiResp GeminiResponse
    if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
        return "", fmt.Errorf("failed to decode response: %v", err)
    }
    
    if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
        return "", fmt.Errorf("no response from Gemini")
    }
    
    response := geminiResp.Candidates[0].Content.Parts[0].Text
    
    // Clean up the response
    response = strings.TrimSpace(response)
    
    return response, nil
}