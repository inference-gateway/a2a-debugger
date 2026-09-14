package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
	zap "go.uber.org/zap"

	types "github.com/inference-gateway/adk/types"
)

// runAuthCmd runs the auth command against the given mock and captures stdout
func runAuthCmd(t *testing.T, mock *mockA2AClient) (string, error) {
	t.Helper()

	originalClient := a2aClient
	originalLogger := logger
	a2aClient = mock
	logger = zap.NewNop()

	viper.Set("output", "yaml")
	viper.Set("auth-header", "Authorization")

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := authCmd.RunE(&cobra.Command{}, []string{"test-token"})

	_ = w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	a2aClient = originalClient
	logger = originalLogger
	viper.Set("token", "")

	return buf.String(), err
}

func agentCardWithExtended(supported bool) *types.AgentCard {
	return &types.AgentCard{
		Name:                      "test-agent",
		SupportsExtendedAgentCard: &supported,
		SecuritySchemes: map[string]types.SecurityScheme{
			"bearer": {},
		},
	}
}

func TestAuthCmd_Success(t *testing.T) {
	called := false
	mock := &mockA2AClient{
		getAgentCardFunc: func(ctx context.Context) (*types.AgentCard, error) {
			return agentCardWithExtended(true), nil
		},
		getAuthenticatedExtendedCardFunc: func(ctx context.Context, params types.GetAuthenticatedExtendedCardParams) (*types.JSONRPCSuccessResponse, error) {
			called = true
			return &types.JSONRPCSuccessResponse{
				Result: map[string]any{"name": "test-agent-extended"},
			}, nil
		},
	}

	output, err := runAuthCmd(t, mock)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if !called {
		t.Error("Expected the extended card RPC to be called")
	}

	for _, want := range []string{"authenticated: true", "header: Authorization", "test-agent-extended"} {
		if !strings.Contains(output, want) {
			t.Errorf("Expected output to contain %q, got:\n%s", want, output)
		}
	}
}

func TestAuthCmd_ExtendedCardUnsupported(t *testing.T) {
	mock := &mockA2AClient{
		getAgentCardFunc: func(ctx context.Context) (*types.AgentCard, error) {
			return agentCardWithExtended(false), nil
		},
		getAuthenticatedExtendedCardFunc: func(ctx context.Context, params types.GetAuthenticatedExtendedCardParams) (*types.JSONRPCSuccessResponse, error) {
			t.Error("Expected the extended card RPC not to be called")
			return nil, nil
		},
	}

	output, err := runAuthCmd(t, mock)
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}
	if !strings.Contains(output, "authenticated: false") {
		t.Errorf("Expected output to contain 'authenticated: false', got:\n%s", output)
	}
}

func TestAuthCmd_RPCErrorsAreFriendly(t *testing.T) {
	tests := []struct {
		name    string
		rpcErr  error
		wantMsg string
	}{
		{"extended card not configured", fmt.Errorf(`{"error":{"code":-32007,"message":"boom"}}`), "Extended agent card not configured"},
		{"not supported", fmt.Errorf(`{"error":{"code":-32004,"message":"boom"}}`), "does not support the authenticated extended card"},
		{"unauthorized", fmt.Errorf(`unexpected status code: 401, body: {"error":"invalid_token"}`), "Authentication rejected"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockA2AClient{
				getAgentCardFunc: func(ctx context.Context) (*types.AgentCard, error) {
					return agentCardWithExtended(true), nil
				},
				getAuthenticatedExtendedCardFunc: func(ctx context.Context, params types.GetAuthenticatedExtendedCardParams) (*types.JSONRPCSuccessResponse, error) {
					return nil, tt.rpcErr
				},
			}

			_, err := runAuthCmd(t, mock)
			if err == nil {
				t.Fatal("Expected an error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("Expected error to contain %q, got: %v", tt.wantMsg, err)
			}
		})
	}
}

func TestAuthHeader(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		header    string
		wantName  string
		wantValue string
	}{
		{"no token", "", "Authorization", "", ""},
		{"bearer prefix added", "jwt", "Authorization", "Authorization", "Bearer jwt"},
		{"bearer prefix kept", "Bearer jwt", "Authorization", "Authorization", "Bearer jwt"},
		{"custom header raw", "secret", "X-Api-Key", "X-Api-Key", "secret"},
		{"empty header defaults", "jwt", "", "Authorization", "Bearer jwt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set("token", tt.token)
			viper.Set("auth-header", tt.header)
			defer func() {
				viper.Set("token", "")
				viper.Set("auth-header", "Authorization")
			}()

			name, value := authHeader()
			if name != tt.wantName || value != tt.wantValue {
				t.Errorf("authHeader() = (%q, %q), want (%q, %q)", name, value, tt.wantName, tt.wantValue)
			}
		})
	}
}
