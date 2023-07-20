package main

import (
	"context"
	"fmt"
	"generator/service"
	"github.com/2q4t-plutus/envopt"
	"github.com/andygrunwald/go-jira"
	"github.com/google/go-github/v39/github"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
	"os"
	"os/exec"
	"runtime"
)

func init() {

	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)

}

func main() {
	var err error
	//var err error
	envopt.Validate("envopt.json")

	openbrowser("http://localhost:8090/")

	s := &service.Service{}

	s.DB = DbConnection()

	// Создание клиента GitHub с токеном доступа.
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: envopt.GetEnv("GITHUB_TOKEN")})
	tc := oauth2.NewClient(context.Background(), ts)
	s.GitHubClient = github.NewClient(tc)

	s.GitLabClient = gitlab.NewClient(nil, envopt.GetEnv("GITLAB_TOKEN"))

	s.WsClients = make(map[*service.WsClient]bool)

	tp := jira.BasicAuthTransport{
		Username: envopt.GetEnv("JIRA_USER"),
		Password: envopt.GetEnv("JIRA_TOKEN"),
	}

	s.JiraClient, err = jira.NewClient(tp.Client(), envopt.GetEnv("JIRA_BASE_URL"))
	if err != nil {
		log.Error(err)
	}

	go initWebSocketHook(s)

	if err := StartWebServer(s); err != nil {
		log.Printf("failure init server %s", err)
	}
}

func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}
