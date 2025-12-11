// github/api.go - Add this new file to handle GitHub API calls
package github

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "codesage/config"
)

// GitHub API structures
type PullRequestFiles struct {
    Filename string `json:"filename"`
    Patch    string `json:"patch"`
    Status   string `json:"status"`
}

type CommentRequest struct {
    Body string `json:"body"`
}

// GetPRFiles fetches the file changes for a pull request
func GetPRFiles(owner, repo string, prNumber int, cfg *config.Config) ([]PullRequestFiles, error) {
    url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/files", owner, repo, prNumber)
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    // Prefer installation token if available (set by caller via context of cfg.GitHubToken)
    req.Header.Set("Authorization", "Bearer "+cfg.GitHubToken)
    req.Header.Set("Accept", "application/vnd.github.v3+json")
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    
    var files []PullRequestFiles
    if err := json.Unmarshal(body, &files); err != nil {
        return nil, err
    }
    
    return files, nil
}

// PostComment posts a comment on a pull request
func PostComment(owner, repo string, prNumber int, comment string, cfg *config.Config) error {
    url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments", owner, repo, prNumber)
    
    commentReq := CommentRequest{Body: comment}
    jsonData, err := json.Marshal(commentReq)
    if err != nil {
        return err
    }
    
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }
    
    req.Header.Set("Authorization", "Bearer "+cfg.GitHubToken)
    req.Header.Set("Accept", "application/vnd.github.v3+json")
    req.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusCreated {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("failed to post comment: %s", string(body))
    }
    
    return nil
}