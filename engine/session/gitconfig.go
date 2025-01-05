package session

import (
	bytes "bytes"
	context "context"
	fmt "fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	grpc "google.golang.org/grpc"
)

var gitConfigMutex sync.Mutex

type GitConfigAttachable struct {
	rootCtx context.Context

	UnimplementedGitConfigServer
}

func NewGitConfigAttachable(rootCtx context.Context) GitConfigAttachable {
	return GitConfigAttachable{
		rootCtx: rootCtx,
	}
}

func (s GitConfigAttachable) Register(srv *grpc.Server) {
	RegisterGitConfigServer(srv, &s)
}

func newGitConfigErrorResponse(errorType GitConfigErrorInfo_GitConfigErrorType, message string) *GitConfigResponse {
	return &GitConfigResponse{
		Result: &GitConfigResponse_Error{
			Error: &GitConfigErrorInfo{
				Type:    errorType,
				Message: message,
			},
		},
	}
}

// GetCredential retrieves Git credentials for the given request using the local Git credential system.
// The function has a timeout of 30 seconds and ensures thread-safe execution.
//
// It follows Git's credential helper protocol and error handling:
// - If Git can't find or execute a helper: CREDENTIAL_RETRIEVAL_FAILED
// - If a helper returns invalid format or no credentials: Git handles it as a failure (CREDENTIAL_RETRIEVAL_FAILED)
// - If the command times out: TIMEOUT
// - If Git is not installed: NO_GIT
// - If the request is invalid: INVALID_REQUEST
func (s GitConfigAttachable) GetConfig(ctx context.Context, req *GitConfigRequest) (*GitConfigResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// // Validate request
	// if req.Host == "" || req.Protocol == "" {
	// 	return newGitConfigErrorResponse(GC_INVALID_REQUEST, "Host and protocol are required"), nil
	// }

	// Check if git is installed
	if _, err := exec.LookPath("git"); err != nil {
		return newGitConfigErrorResponse(GC_NO_GIT, "Git is not installed or not in PATH"), nil
	}

	// Ensure no parallel execution of the git CLI happens
	gitConfigMutex.Lock()
	defer gitConfigMutex.Unlock()

	// Prepare the git credential fill command
	cmd := exec.CommandContext(ctx, "git", "config", "-l")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	// // Prepare input
	// input := fmt.Sprintf("protocol=%s\nhost=%s\n", req.Protocol, req.Host)
	// if req.Path != "" {
	// 	input += fmt.Sprintf("path=%s\n", req.Path)
	// }
	// input += "\n"
	// cmd.Stdin = strings.NewReader(input)

	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"SSH_ASKPASS=echo",
	)

	// Run the command
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return newGitConfigErrorResponse(GC_TIMEOUT, "Git config command timed out"), nil
		}
		return newGitConfigErrorResponse(GC_GIT_CONFIG_RETRIEVAL_FAILED, fmt.Sprintf("Failed to retrieve config: %v", err)), nil
	}

	// Parse the output
	// cred, err := parseGitConfigOutput(stdout.Bytes())
	// if err != nil {
	// 	return newGitConfigErrorResponse(GC_GIT_CONFIG_RETRIEVAL_FAILED, fmt.Sprintf("Failed to retrieve credentials: %v", err)), nil
	// }

	return &GitConfigResponse{
		Result: &GitConfigResponse_X{
			X: &GitConfigX{
				Content:   stdout.String(),
				Goprivate: "this is goprivate",
			},
		},
	}, nil
}

// func parseGitCredentialOutput(output []byte) (*CredentialInfo, error) {
// 	if len(output) == 0 {
// 		return nil, fmt.Errorf("no output from credential helper")
// 	}

// 	cred := make(map[string]string)
// 	scanner := bufio.NewScanner(bytes.NewReader(output))

// 	for scanner.Scan() {
// 		line := scanner.Text()
// 		if line == "" {
// 			continue
// 		}
// 		parts := strings.SplitN(line, "=", 2)
// 		if len(parts) != 2 {
// 			return nil, fmt.Errorf("invalid format: line doesn't match key=value pattern")
// 		}

// 		cred[parts[0]] = parts[1]
// 	}

// 	if err := scanner.Err(); err != nil {
// 		return nil, fmt.Errorf("error reading credential helper output: %w", err)
// 	}

// 	if cred["username"] == "" || cred["password"] == "" {
// 		// should not be possible
// 		return nil, fmt.Errorf("incomplete credentials: missing username or password")
// 	}

// 	return &CredentialInfo{
// 		Protocol: cred["protocol"],
// 		Host:     cred["host"],
// 		Username: cred["username"],
// 		Password: cred["password"],
// 	}, nil
// }
