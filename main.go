package main

import (
	"codesage/config"
	"codesage/server"
	"fmt"
	"log"

)
func main(){
    cfg:=config.Load()
    r:=server.SetupRouter(cfg)
	tokenPreview := ""
if len(cfg.GitHubToken) > 10 {
    tokenPreview = cfg.GitHubToken[:10] + "..."
} else {
    tokenPreview = cfg.GitHubToken // will show full if shorter
}

fmt.Printf("🚀 CodeSage server running on port %s\n", cfg.Port)
fmt.Println("📝 GitHub token loaded:", tokenPreview)

    
if err:=r.Run(":"+cfg.Port);err!=nil{
    log.Fatal(err)
}
}