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
    
    // Enhanced prompt for better code review
    prompt := fmt.Sprintf(`You are CodeSage AI, a helpful code review assistant. Analyze this Pull Request and provide constructive feedback.

**PR Title:** %s

**Code Changes:**
%s

Please provide a structured review covering:

### 🔍 Summary
Brief overview of what this PR does.

### ✅ What's Good
Highlight positive aspects of the code.

### 🐛 Potential Issues
- Any bugs or logical errors you spot
- Edge cases that might not be handled

### 🔒 Security Considerations
- Any security vulnerabilities or concerns
- Authentication/authorization issues

### ⚡ Performance & Best Practices
- Performance improvements or concerns
- Code quality and maintainability suggestions

### 💡 Suggestions
Specific, actionable recommendations for improvement.

Keep your feedback constructive, educational, and focus on the most important issues first.`, title, diff)
    
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
    
    url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key=" + apiKey
    
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