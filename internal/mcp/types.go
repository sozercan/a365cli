package mcp

// JSONRPCRequest is a JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// JSONRPCResponse is a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string     `json:"jsonrpc"`
	ID      int        `json:"id"`
	Result  *Result    `json:"result,omitempty"`
	Error   *RPCError  `json:"error,omitempty"`
}

// RPCError is a JSON-RPC error object.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (e *RPCError) Error() string {
	return e.Message
}

// Result is the MCP result envelope.
type Result struct {
	Content []Content `json:"content,omitempty"`
	// For tools/list responses
	Tools []ToolInfo `json:"tools,omitempty"`
	// For initialize responses
	ServerInfo   *ServerInfo `json:"serverInfo,omitempty"`
	ProtocolVersion string   `json:"protocolVersion,omitempty"`
	Capabilities any        `json:"capabilities,omitempty"`
}

// Content is a single content block in an MCP result.
type Content struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// ToolInfo describes an available MCP tool.
type ToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	InputSchema any    `json:"inputSchema,omitempty"`
}

// ServerInfo from the initialize response.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ToolCallParams is the params for a tools/call request.
type ToolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

// InitializeParams is the params for an initialize request.
type InitializeParams struct {
	ProtocolVersion string     `json:"protocolVersion"`
	ClientInfo      ClientInfo `json:"clientInfo"`
	Capabilities    struct{}   `json:"capabilities"`
}

// ClientInfo identifies the client in the initialize handshake.
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
