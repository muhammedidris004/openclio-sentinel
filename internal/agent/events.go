package agent

// AgentEventType identifies the kind of delegation event emitted during a run.
type AgentEventType string

const (
	EventDelegationStart AgentEventType = "delegation_start"
	EventSubagentSpawn   AgentEventType = "subagent_spawn"
	EventSubagentTool    AgentEventType = "subagent_tool"
	EventSubagentDone    AgentEventType = "subagent_done"
	EventDelegationDone  AgentEventType = "delegation_done"
)

// AgentEvent carries real-time delegation telemetry to the caller via the
// onEvent callback registered on RunStream.
type AgentEvent struct {
	Type AgentEventType `json:"type"`

	// delegation_start
	TaskCount int      `json:"task_count,omitempty"`
	Tasks     []string `json:"tasks,omitempty"`

	// subagent_spawn
	SubagentID   string   `json:"subagent_id,omitempty"`
	Task         string   `json:"task,omitempty"`
	Tools        []string `json:"tools,omitempty"`
	MemoryAccess string   `json:"memory_access,omitempty"`

	// subagent_tool
	ToolName string `json:"tool_name,omitempty"`

	// subagent_done
	ResultSummary string     `json:"result_summary,omitempty"`
	Usage         TotalUsage `json:"usage,omitempty"`
	Error         string     `json:"error,omitempty"`

	// delegation_done
	SubagentCount int `json:"subagent_count,omitempty"`
}
