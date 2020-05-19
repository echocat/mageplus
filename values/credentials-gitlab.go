package values

import (
	"context"
	"errors"
	mhttp "github.com/echocat/mageplus/http"
	"net/http"
	"os"
)

var (
	NoGitLabCredentialsError = errors.New("there is neither a CI_JOB_TOKEN nor GITLAB_TOKEN set." +
		" Please go to your GitLab profile page and create yourself a token with at least the permission" +
		" `read_api` and store it as environment variables GITLAB_TOKEN=<created_token>" +
		" ; For more details refer: https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html")
)

type GitLabCredentialsType string

const (
	GitLabJobToken     GitLabCredentialsType = "Job-Token"
	GitLabPrivateToken GitLabCredentialsType = "Private-Token"
)

type GitLabCredentials struct {
	Type  GitLabCredentialsType
	Token string
}

func GetGitLabCredentials() (GitLabCredentials, error) {
	if v, ok := os.LookupEnv("CI_JOB_TOKEN"); ok {
		return GitLabCredentials{GitLabJobToken, v}, nil
	}
	if v, ok := os.LookupEnv("GITLAB_TOKEN"); ok {
		return GitLabCredentials{GitLabPrivateToken, v}, nil
	}
	return GitLabCredentials{}, NoGitLabCredentialsError
}

func MustGetGitLabCredentials() GitLabCredentials {
	v, err := GetGitLabCredentials()
	if err != nil {
		errorLog.Fatalln("Error:", err)
	}
	return v
}

func (instance GitLabCredentials) BeforeRequest(ctx context.Context, req *http.Request) (context.Context, *http.Request, error) {
	req.Header.Set(string(instance.Type), instance.Token)
	return ctx, req, nil
}

func (instance GitLabCredentials) Self() mhttp.Plugin {
	return instance
}
