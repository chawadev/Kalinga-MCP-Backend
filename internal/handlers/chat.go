package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/chawadev/kalinga-backend/internal/ai"
	"github.com/chawadev/kalinga-backend/internal/mcp"
	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	aiService *ai.Service
	mcpServer *mcp.Server
}

func NewChatHandler(aiService *ai.Service, mcpServer *mcp.Server) *ChatHandler {
	return &ChatHandler{
		aiService: aiService,
		mcpServer: mcpServer,
	}
}

type ChatRequest struct {
	Message string `json:"message"`
}

type ChatResponse struct {
	Response  string                `json:"response"`
	Logs      []LogEntry            `json:"logs"`
	ToolCalls []ToolCall            `json:"toolCalls,omitempty"`
}

type LogEntry struct {
	Type      string                 `json:"type"`
	Tool      string                 `json:"tool,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Result    interface{}            `json:"result,omitempty"`
	Latency   int64                  `json:"latency,omitempty"`
	Timestamp string                 `json:"timestamp"`
}

type ToolCall struct {
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
}

func (h *ChatHandler) Chat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message is required"})
		return
	}

	logs := []LogEntry{}
	var toolCalls []ToolCall

	// Check if AI service is available
	if h.aiService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI service not available"})
		return
	}

	// Generate initial response
	genReq := &ai.GenerationRequest{
		Message: req.Message,
		System:  "You are Kalinga, a financial intelligence assistant. You help users manage their finances, connect mobile money accounts, build financial profiles, and verify claims. Always identify yourself as Kalinga when asked who you are.",
	}

	genResp, err := h.aiService.GenerateContent(c.Request.Context(), genReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Simple keyword-based tool execution for now
	// TODO: Implement proper function calling when Gemini API supports it
	message := req.Message
	var toolResult map[string]interface{}
	var toolName string

	if containsKeywords(message, []string{"status", "health", "check"}) {
		toolName = "check_status"
		toolResult = handleCheckStatus(map[string]interface{}{})
	} else if containsKeywords(message, []string{"connect", "momo", "mobile money", "account"}) {
		toolName = "connect_momo_account"
		toolResult = handleConnectMomoAccount(map[string]interface{}{
			"phone_number": "+265999123456",
			"provider":     "MTN",
		})
	} else if containsKeywords(message, []string{"financial", "profile", "credit", "loan"}) {
		toolName = "build_financial_profile"
		toolResult = handleBuildFinancialProfile(map[string]interface{}{
			"phone_number": "+265999123456",
		})
	} else if containsKeywords(message, []string{"verify", "claim", "refund"}) {
		toolName = "verify_claim"
		toolResult = handleVerifyClaim(map[string]interface{}{
			"claim_id": "CLM-001",
			"amount":   500,
			"merchant": "Store",
		})
	}

	if toolResult != nil {
		logs = append(logs, LogEntry{
			Type:      "tool_call",
			Tool:      toolName,
			Parameters: map[string]interface{}{},
			Timestamp: getCurrentTimestamp(),
		})

		logs = append(logs, LogEntry{
			Type:      "tool_result",
			Tool:      toolName,
			Result:    toolResult,
			Timestamp: getCurrentTimestamp(),
		})

		toolCalls = append(toolCalls, ToolCall{
			Name:       toolName,
			Parameters: map[string]interface{}{},
		})
	}

	c.JSON(http.StatusOK, ChatResponse{
		Response:  genResp.Text,
		Logs:      logs,
		ToolCalls: toolCalls,
	})
}

func containsKeywords(text string, keywords []string) bool {
	lowerText := strings.ToLower(text)
	for _, keyword := range keywords {
		if strings.Contains(lowerText, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

// Tool handlers
func handleCheckStatus(args map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"status":  "ok",
		"service": "kalinga-mcp-backend",
		"version": "0.1.0",
		"health":  "healthy",
	}
}

func handleConnectMomoAccount(args map[string]interface{}) map[string]interface{} {
	phoneNumber, _ := args["phone_number"].(string)
	provider, _ := args["provider"].(string)

	return map[string]interface{}{
		"status":         "approved",
		"phone_number":   phoneNumber,
		"provider":       provider,
		"account_linked": true,
		"message":        "Mobile money account successfully connected",
	}
}

func handleBuildFinancialProfile(args map[string]interface{}) map[string]interface{} {
	phoneNumber, _ := args["phone_number"].(string)

	return map[string]interface{}{
		"phone_number":          phoneNumber,
		"credit_score":          75,
		"loan_readiness":        "high",
		"income_regularity":     85,
		"volatility":           12,
		"repayment_capacity":    92,
		"expense_to_income":     68,
		"projected_savings":     "+$420/mo",
		"financial_wellness":    78,
		"risk_assessment":       "low",
		"recommended_loan_amt":  5000,
	}
}

func handleVerifyClaim(args map[string]interface{}) map[string]interface{} {
	claimID, _ := args["claim_id"].(string)
	amount, _ := args["amount"].(float64)
	merchant, _ := args["merchant"].(string)

	isSuspicious := amount > 1000 && merchant == "unknown"

	result := map[string]interface{}{
		"claim_id":      claimID,
		"amount":        amount,
		"merchant":      merchant,
		"status":        "verified",
		"risk_level":    "low",
		"flagged":       isSuspicious,
		"pattern_match": "regular_merchant",
		"confidence":    0.95,
	}

	if isSuspicious {
		result["status"] = "flagged"
		result["risk_level"] = "high"
		result["pattern_match"] = "suspicious_pattern"
		result["confidence"] = 0.72
	}

	return result
}

func getCurrentTimestamp() string {
	return time.Now().Format(time.RFC3339)
}
