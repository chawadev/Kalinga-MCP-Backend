package mcp

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type JSONRPCRequest struct {
	Jsonrpc string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
}

type JSONRPCResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *RPCError) Error() string {
	return e.Message
}

// MCP Initialize Request/Response structures
type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      map[string]interface{} `json:"clientInfo"`
}

type InitializeResult struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ServerInfo      map[string]interface{} `json:"serverInfo"`
}

// MCP Tools structures
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

// MCP Tool Call structures
type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ToolCallResult struct {
	Content []ContentItem `json:"content"`
}

type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Server struct {
	tools map[string]ToolHandler
}

type ToolHandler func(args map[string]interface{}) map[string]interface{}

func NewServer() *Server {
	s := &Server{
		tools: make(map[string]ToolHandler),
	}
	s.registerDefaultTools()
	return s
}

func (s *Server) registerDefaultTools() {
	// Register default financial tools
	s.RegisterTool("check_status", handleCheckStatus)
	s.RegisterTool("connect_momo_account", handleConnectMomoAccount)
	s.RegisterTool("build_financial_profile", handleBuildFinancialProfile)
	s.RegisterTool("verify_claim", handleVerifyClaim)
}

func (s *Server) RegisterTool(name string, handler ToolHandler) {
	s.tools[name] = handler
}

func (s *Server) HandleRequest(c *gin.Context) {
	var rpcReq JSONRPCRequest
	if err := c.ShouldBindJSON(&rpcReq); err != nil {
		c.JSON(http.StatusOK, JSONRPCResponse{
			Jsonrpc: "2.0",
			ID:      nil,
			Error:   &RPCError{Code: -32700, Message: "Parse error"},
		})
		return
	}

	var result interface{}
	var err error

	switch rpcReq.Method {
	case "initialize":
		result = s.handleInitialize(rpcReq.Params)
	case "tools/list":
		result = s.handleToolsList()
	case "tools/call":
		result, err = s.handleToolCall(rpcReq.Params)
	default:
		err = &RPCError{Code: -32601, Message: "Method not found"}
	}

	if err != nil {
		c.JSON(http.StatusOK, JSONRPCResponse{
			Jsonrpc: "2.0",
			ID:      rpcReq.ID,
			Error:   err.(*RPCError),
		})
		return
	}

	c.JSON(http.StatusOK, JSONRPCResponse{
		Jsonrpc: "2.0",
		ID:      rpcReq.ID,
		Result:  result,
	})
}

func (s *Server) handleInitialize(params map[string]interface{}) InitializeResult {
	return InitializeResult{
		ProtocolVersion: "2025-11-25",
		Capabilities: map[string]interface{}{
			"tools": map[string]interface{}{
				"listChanged": false,
			},
		},
		ServerInfo: map[string]interface{}{
			"name":    "kalinga-mcp-backend",
			"version": "0.1.0",
		},
	}
}

func (s *Server) handleToolsList() ToolsListResult {
	tools := []Tool{
		{
			Name:        "check_status",
			Description: "Verifies backend connection health and server status.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []interface{}{},
			},
		},
		{
			Name:        "connect_momo_account",
			Description: "Connects a Mobile Money account using phone number and provider.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"phone_number": map[string]interface{}{
						"type":        "string",
						"description": "User's mobile number",
					},
					"provider": map[string]interface{}{
						"type":        "string",
						"description": "e.g., MTN, Airtel",
					},
				},
				"required": []interface{}{"phone_number", "provider"},
			},
		},
		{
			Name:        "build_financial_profile",
			Description: "Calculates credit readiness and financial metrics for a user.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"phone_number": map[string]interface{}{
						"type":        "string",
						"description": "User's phone number",
					},
				},
				"required": []interface{}{"phone_number"},
			},
		},
		{
			Name:        "verify_claim",
			Description: "Evaluates incoming refund or wrong-number transfer claims against fraud patterns.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"claim_id": map[string]interface{}{
						"type":        "string",
						"description": "Unique identifier for the claim",
					},
					"amount": map[string]interface{}{
						"type":        "number",
						"description": "Transaction amount",
					},
				},
				"required": []interface{}{"claim_id", "amount"},
			},
		},
	}

	return ToolsListResult{Tools: tools}
}

func (s *Server) handleToolCall(params map[string]interface{}) (interface{}, error) {
	name, _ := params["name"].(string)
	arguments, _ := params["arguments"].(map[string]interface{})

	handler, exists := s.tools[name]
	if !exists {
		return nil, &RPCError{Code: -32601, Message: "Tool not found"}
	}

	toolResult := handler(arguments)

	// Convert tool result to JSON string for MCP content format
	resultJSON, err := json.Marshal(toolResult)
	if err != nil {
		log.Printf("Error marshaling tool result: %v", err)
		return nil, &RPCError{Code: -32603, Message: "Internal error"}
	}

	return ToolCallResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: string(resultJSON),
			},
		},
	}, nil
}

// Default tool handlers
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
