// github/webhook.go - Fixed to handle both JSON and form-encoded data
package github

import (
    "encoding/json"
    "fmt"
    "strings"
    "io"
    "net/url"
    "github.com/gin-gonic/gin"
    "codesage/config"
    "codesage/ai"
)

func HandleWebhook(c *gin.Context, cfg *config.Config) {
    fmt.Println("📥 GitHub webhook received")
    eventType := c.GetHeader("X-GitHub-Event")
    
    switch eventType {
    case "pull_request":
        handlePullRequest(c, cfg)
        return
    case "ping":
        fmt.Println("🏓 Ping event received - webhook setup successful!")
        c.JSON(200, gin.H{"status": "pong"})
        return
    case "push":
        fmt.Println("📤 Push event received - but CodeSage only analyzes Pull Requests")
        fmt.Println("💡 To test CodeSage, create or update a Pull Request")
        c.JSON(200, gin.H{"status": "received", "message": "Push events ignored"})
        return
    default:
        fmt.Printf("⚠️ Unhandled event: %s\n", eventType)
        c.JSON(200, gin.H{"status": "received", "event": eventType})
        return
    }
}

func handlePullRequest(c *gin.Context, cfg *config.Config) {
    // Read the raw body first
    body, err := io.ReadAll(c.Request.Body)
    if err != nil {
        fmt.Printf("❌ Failed to read request body: %v\n", err)
        c.JSON(400, gin.H{"error": "Failed to read body"})
        return
    }
    
    bodyStr := string(body)
    var payloadStr string
    
    // Check if it's form-encoded data (starts with "payload=")
    if strings.HasPrefix(bodyStr, "payload=") {
        fmt.Println("🔍 Detected form-encoded payload")
        // URL decode the payload
        decoded, err := url.QueryUnescape(bodyStr[8:]) // Remove "payload=" prefix
        if err != nil {
            fmt.Printf("❌ Failed to URL decode: %v\n", err)
            c.JSON(400, gin.H{"error": "Failed to decode payload"})
            return
        }
        payloadStr = decoded
    } else {
        fmt.Println("🔍 Detected raw JSON payload")
        payloadStr = bodyStr
    }
    
    // Debug: Print first 200 characters of the decoded payload
    if len(payloadStr) > 200 {
        fmt.Printf("🔍 Decoded payload preview: %s...\n", payloadStr[:200])
    } else {
        fmt.Printf("🔍 Full decoded payload: %s\n", payloadStr)
    }
    
    // Parse the JSON
    var payload map[string]interface{}
    if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
        fmt.Printf("❌ Failed to parse JSON: %v\n", err)
        fmt.Printf("🔍 Trying to parse: %s\n", payloadStr[:min(500, len(payloadStr))])
        c.JSON(400, gin.H{"error": "Invalid JSON payload"})
        return
    }
    
    // Check if required fields exist
    if payload["action"] == nil {
        fmt.Println("❌ Missing 'action' field in payload")
        c.JSON(400, gin.H{"error": "Missing action field"})
        return
    }
    
    if payload["pull_request"] == nil {
        fmt.Println("❌ Missing 'pull_request' field in payload")
        c.JSON(400, gin.H{"error": "Missing pull_request field"})
        return
    }
    
    action, ok := payload["action"].(string)
    if !ok {
        fmt.Println("❌ Action field is not a string")
        c.JSON(400, gin.H{"error": "Invalid action field type"})
        return
    }
    
    fmt.Printf("🎯 PR Action: %s\n", action)
    
    // Only analyze when PR is opened or synchronized (new commits)
    if action != "opened" && action != "synchronize" {
        fmt.Printf("⏭️ Skipping action: %s\n", action)
        c.JSON(200, gin.H{"status": "received", "message": "Action ignored"})
        return
    }
    
    // Extract PR information with error checking
    prData, ok := payload["pull_request"].(map[string]interface{})
    if !ok {
        fmt.Println("❌ pull_request field is not an object")
        c.JSON(400, gin.H{"error": "Invalid pull_request field"})
        return
    }
    
    repoData, ok := payload["repository"].(map[string]interface{})
    if !ok {
        fmt.Println("❌ repository field is not an object")
        c.JSON(400, gin.H{"error": "Invalid repository field"})
        return
    }
    
    // Safe extraction of nested fields
    title, _ := prData["title"].(string)
    prNumberFloat, _ := prData["number"].(float64)
    prNumber := int(prNumberFloat)
    
    ownerData, _ := repoData["owner"].(map[string]interface{})
    owner, _ := ownerData["login"].(string)
    repo, _ := repoData["name"].(string)
    
    userData, _ := prData["user"].(map[string]interface{})
    user, _ := userData["login"].(string)
    
    if title == "" || owner == "" || repo == "" || prNumber == 0 {
        fmt.Printf("❌ Missing required PR data: title=%s, owner=%s, repo=%s, number=%d\n", 
                   title, owner, repo, prNumber)
        c.JSON(400, gin.H{"error": "Missing required PR data"})
        return
    }
    
    fmt.Printf("📌 Analyzing PR #%d: \"%s\" by %s in %s/%s\n", prNumber, title, user, owner, repo)
    
    // Step 1: Get the file changes from GitHub
    fmt.Println("🔄 Fetching PR file changes from GitHub API...")
    files, err := GetPRFiles(owner, repo, prNumber, cfg)
    if err != nil {
        fmt.Printf("❌ Failed to get PR files: %v\n", err)
        c.JSON(500, gin.H{"error": "Failed to fetch PR files"})
        return
    }
    
    if len(files) == 0 {
        fmt.Println("⚠️ No files changed in this PR")
        c.JSON(200, gin.H{"status": "received", "message": "No files to analyze"})
        return
    }
    
    // Step 2: Combine all file changes into one diff
    var fullDiff strings.Builder
    totalLines := 0
    for _, file := range files {
        if file.Patch != "" {
            fullDiff.WriteString(fmt.Sprintf("\n--- %s ---\n", file.Filename))
            fullDiff.WriteString(file.Patch)
            fullDiff.WriteString("\n")
            // Count approximate lines for debugging
            totalLines += strings.Count(file.Patch, "\n")
        }
    }
    
    if fullDiff.Len() == 0 {
        fmt.Println("⚠️ No code changes to analyze")
        c.JSON(200, gin.H{"status": "received", "message": "No code changes"})
        return
    }
    
    fmt.Printf("📊 Analyzing %d lines of code changes\n", totalLines)
    
    // Step 3: Send to AI for analysis
    fmt.Println("🤖 Sending to AI for analysis...")
    analysis, err := ai.AnalyzeWithGemini(fullDiff.String(), title)
    if err != nil {
        fmt.Printf("❌ AI analysis failed: %v\n", err)
        c.JSON(500, gin.H{"error": "AI analysis failed"})
        return
    }
    
    fmt.Printf("✅ AI analysis completed (%d characters)\n", len(analysis))
    
    // Step 4: Format the comment nicely
    comment := fmt.Sprintf(`## 🤖 CodeSage AI Review

%s

---
*This review was automatically generated by CodeSage. Please review the suggestions and apply them as appropriate.*`, analysis)
    
    // Step 5: Post the comment to GitHub
    fmt.Println("💬 Posting comment to GitHub...")
    if err := PostComment(owner, repo, prNumber, comment, cfg); err != nil {
        fmt.Printf("❌ Failed to post comment: %v\n", err)
        c.JSON(500, gin.H{"error": "Failed to post comment"})
        return
    }
    
    fmt.Println("✅ Successfully posted AI review comment!")
    c.JSON(200, gin.H{"status": "success", "message": "AI review posted"})
}

// Helper function for min (Go doesn't have built-in min for int)
func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}