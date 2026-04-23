package copilot

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sozercan/a365cli/internal/commands"
	"github.com/sozercan/a365cli/internal/config"
	"github.com/sozercan/a365cli/internal/output"
)

const copilotAvailableAgentsTool = "M365_Copilot_Get_Available_Agents"

// CopilotAgentsCmd lists available Copilot agents and their chat selectors.
type CopilotAgentsCmd struct{}

type agentInfo struct {
	Name                string
	Selector            string
	TitleID             string
	TitleName           string
	Type                string
	DeveloperName       string
	Description         string
	Version             string
	AcquisitionState    string
	SharedSelector      bool
	SharedSelectorCount int
}

func (c *CopilotAgentsCmd) Run(ctx *commands.Context) error {
	agents, err := fetchAvailableAgents(ctx)
	if err != nil {
		return err
	}
	return ctx.Output.PrintList("agents", output.CopilotAgentColumns, agentRows(agents))
}

func copilotAgentsEndpoint() string {
	return config.Endpoint("dasearch")
}

func fetchAvailableAgents(ctx *commands.Context) ([]agentInfo, error) {
	client := ctx.NewMCPClient(copilotAgentsEndpoint())
	if err := client.Initialize(ctx.Ctx); err != nil {
		return nil, fmt.Errorf("initialize: %w", err)
	}

	resp, err := client.CallTool(ctx.Ctx, copilotAvailableAgentsTool, map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("list Copilot agents: %w", err)
	}

	data, err := output.ExtractContent(resp)
	if err != nil {
		return nil, err
	}

	return normalizeAvailableAgents(data), nil
}

func normalizeAvailableAgents(data map[string]any) []agentInfo {
	rows := output.ToRows(data, "availableAgents")
	if rows == nil {
		rows = output.ToRows(data, "agents")
	}

	agents := make([]agentInfo, 0, len(rows))
	seen := map[string]struct{}{}
	for _, row := range rows {
		agent := normalizeAgentRow(row)
		key := agentDedupeKey(agent)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		agents = append(agents, agent)
	}

	selectorCounts := map[string]int{}
	for _, agent := range agents {
		if agent.Selector != "" {
			selectorCounts[agent.Selector]++
		}
	}
	for i := range agents {
		count := selectorCounts[agents[i].Selector]
		agents[i].SharedSelectorCount = count
		agents[i].SharedSelector = count > 1
	}

	sort.Slice(agents, func(i, j int) bool {
		left := strings.ToLower(agentSortKey(agents[i]))
		right := strings.ToLower(agentSortKey(agents[j]))
		if left != right {
			return left < right
		}
		return agents[i].TitleID < agents[j].TitleID
	})

	return agents
}

func normalizeAgentRow(row map[string]any) agentInfo {
	selector := strings.TrimSpace(stringValue(row, "selector"))
	if selector == "" {
		selector = strings.TrimSpace(stringValue(row, "agentId"))
	}

	name := strings.TrimSpace(stringValue(row, "name"))
	titleName := strings.TrimSpace(stringValue(row, "titleName"))
	if name == "" {
		name = titleName
	}
	if name == "" {
		name = selector
	}
	if name == "" {
		name = strings.TrimSpace(stringValue(row, "titleId"))
	}

	return agentInfo{
		Name:             name,
		Selector:         selector,
		TitleID:          strings.TrimSpace(stringValue(row, "titleId")),
		TitleName:        titleName,
		Type:             strings.TrimSpace(stringValue(row, "type")),
		DeveloperName:    strings.TrimSpace(stringValue(row, "developerName")),
		Description:      strings.TrimSpace(stringValue(row, "description")),
		Version:          strings.TrimSpace(stringValue(row, "version")),
		AcquisitionState: strings.TrimSpace(stringValue(row, "acquisitionState")),
	}
}

func agentDedupeKey(agent agentInfo) string {
	if agent.TitleID != "" {
		return agent.TitleID
	}
	if agent.Name == "" && agent.Selector == "" {
		return ""
	}
	return agent.Name + "\x00" + agent.Selector
}

func agentSortKey(agent agentInfo) string {
	if agent.Name != "" {
		return agent.Name
	}
	if agent.TitleName != "" {
		return agent.TitleName
	}
	if agent.Selector != "" {
		return agent.Selector
	}
	return agent.TitleID
}

func stringValue(row map[string]any, key string) string {
	value, ok := row[key]
	if !ok || value == nil {
		return ""
	}
	s, ok := value.(string)
	if ok {
		return s
	}
	return fmt.Sprintf("%v", value)
}

func agentRows(agents []agentInfo) []map[string]any {
	rows := make([]map[string]any, 0, len(agents))
	for _, agent := range agents {
		status := "ok"
		targetable := true
		switch {
		case agent.Selector == "":
			status = "missing"
			targetable = false
		case agent.SharedSelector:
			status = "shared"
			targetable = false
		}

		rows = append(rows, map[string]any{
			"name":                agent.Name,
			"selector":            agent.Selector,
			"agentId":             agent.Selector,
			"titleId":             agent.TitleID,
			"titleName":           agent.TitleName,
			"type":                agent.Type,
			"developerName":       agent.DeveloperName,
			"description":         agent.Description,
			"version":             agent.Version,
			"acquisitionState":    agent.AcquisitionState,
			"sharedSelector":      agent.SharedSelector,
			"sharedSelectorCount": agent.SharedSelectorCount,
			"targetable":          targetable,
			"status":              status,
		})
	}
	return rows
}

func resolveAgentForChat(ctx *commands.Context, value string) (string, error) {
	query := strings.TrimSpace(value)
	if query == "" {
		return "", nil
	}

	agents, err := fetchAvailableAgents(ctx)
	if err != nil {
		return "", fmt.Errorf("resolve Copilot agent %q: %w", query, err)
	}

	agent, err := resolveAgent(agents, query)
	if err != nil {
		return "", err
	}
	return agent.Selector, nil
}

func resolveAgent(agents []agentInfo, query string) (agentInfo, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return agentInfo{}, fmt.Errorf("agent name or id is required")
	}

	if matches := exactNameMatches(agents, query); len(matches) > 0 {
		return resolveMatchedAgent(query, matches, agents)
	}
	if matches := exactCaseInsensitiveNameMatches(agents, query); len(matches) > 0 {
		return resolveMatchedAgent(query, matches, agents)
	}
	if matches := exactSelectorMatches(agents, query); len(matches) > 0 {
		agent := matches[0]
		if err := validateResolvedAgent(query, agent, agents); err != nil {
			return agentInfo{}, err
		}
		return agent, nil
	}
	if matches := exactTitleIDMatches(agents, query); len(matches) > 0 {
		return resolveMatchedAgent(query, matches, agents)
	}
	if matches := selectorPrefixMatches(agents, query); len(matches) > 0 {
		if len(matches) > 1 {
			return agentInfo{}, ambiguousAgentError(query, matches)
		}
		agent := matches[0]
		if err := validateResolvedAgent(query, agent, agents); err != nil {
			return agentInfo{}, err
		}
		return agent, nil
	}
	if matches := titleIDPrefixMatches(agents, query); len(matches) > 0 {
		return resolveMatchedAgent(query, matches, agents)
	}

	return agentInfo{}, unknownAgentError(query, agents)
}

func exactNameMatches(agents []agentInfo, query string) []agentInfo {
	var matches []agentInfo
	for _, agent := range agents {
		if agent.Name == query {
			matches = append(matches, agent)
		}
	}
	return matches
}

func exactCaseInsensitiveNameMatches(agents []agentInfo, query string) []agentInfo {
	var matches []agentInfo
	for _, agent := range agents {
		if strings.EqualFold(agent.Name, query) {
			matches = append(matches, agent)
		}
	}
	return matches
}

func exactSelectorMatches(agents []agentInfo, query string) []agentInfo {
	var matches []agentInfo
	for _, agent := range agents {
		if agent.Selector == query {
			matches = append(matches, agent)
		}
	}
	return matches
}

func exactTitleIDMatches(agents []agentInfo, query string) []agentInfo {
	var matches []agentInfo
	for _, agent := range agents {
		if agent.TitleID == query {
			matches = append(matches, agent)
		}
	}
	return matches
}

func selectorPrefixMatches(agents []agentInfo, query string) []agentInfo {
	query = strings.ToLower(query)
	seen := map[string]struct{}{}
	matches := make([]agentInfo, 0)
	for _, agent := range agents {
		selector := strings.ToLower(agent.Selector)
		if selector == "" || !strings.HasPrefix(selector, query) {
			continue
		}
		if _, ok := seen[agent.Selector]; ok {
			continue
		}
		seen[agent.Selector] = struct{}{}
		matches = append(matches, agent)
	}
	return matches
}

func titleIDPrefixMatches(agents []agentInfo, query string) []agentInfo {
	query = strings.ToLower(query)
	var matches []agentInfo
	for _, agent := range agents {
		titleID := strings.ToLower(agent.TitleID)
		if titleID == "" || !strings.HasPrefix(titleID, query) {
			continue
		}
		matches = append(matches, agent)
	}
	return matches
}

func resolveMatchedAgent(query string, matches []agentInfo, catalog []agentInfo) (agentInfo, error) {
	if len(matches) > 1 {
		return agentInfo{}, ambiguousAgentError(query, matches)
	}
	if err := validateResolvedAgent(query, matches[0], catalog); err != nil {
		return agentInfo{}, err
	}
	return matches[0], nil
}

func validateResolvedAgent(query string, agent agentInfo, catalog []agentInfo) error {
	if agent.Selector == "" {
		return fmt.Errorf("copilot agent %q does not expose a usable chat selector. Run 'a365 copilot agents' to inspect available selectors", query)
	}
	if !agent.SharedSelector {
		return nil
	}

	shared := sharedSelectorAgents(catalog, agent.Selector)
	return fmt.Errorf(
		"copilot agent %q resolves to shared selector %q, which is used by %d agents (%s). Individual targeting is not available for this selector; run 'a365 copilot agents' to inspect available selectors",
		query,
		agent.Selector,
		len(shared),
		joinAgentNames(shared, 6),
	)
}

func ambiguousAgentError(query string, matches []agentInfo) error {
	return fmt.Errorf(
		"copilot agent %q is ambiguous; matches: %s. Use an exact name or selector, or run 'a365 copilot agents' to list available selectors",
		query,
		joinAgentChoices(matches, 6),
	)
}

func unknownAgentError(query string, agents []agentInfo) error {
	suggestions := suggestAgents(query, agents, 5)
	if len(suggestions) == 0 {
		return fmt.Errorf("unknown Copilot agent %q. Run 'a365 copilot agents' to list available selectors", query)
	}
	return fmt.Errorf(
		"unknown Copilot agent %q. Suggestions: %s. Run 'a365 copilot agents' to list available selectors",
		query,
		strings.Join(suggestions, ", "),
	)
}

func sharedSelectorAgents(catalog []agentInfo, selector string) []agentInfo {
	var shared []agentInfo
	for _, agent := range catalog {
		if agent.Selector == selector {
			shared = append(shared, agent)
		}
	}
	return shared
}

func joinAgentChoices(agents []agentInfo, limit int) string {
	choices := make([]string, 0, len(agents))
	seen := map[string]struct{}{}
	for _, agent := range agents {
		choice := formatAgentChoice(agent)
		if choice == "" {
			continue
		}
		if _, ok := seen[choice]; ok {
			continue
		}
		seen[choice] = struct{}{}
		choices = append(choices, choice)
	}
	sort.Slice(choices, func(i, j int) bool {
		return strings.ToLower(choices[i]) < strings.ToLower(choices[j])
	})
	return joinLimited(choices, limit)
}

func joinAgentNames(agents []agentInfo, limit int) string {
	names := make([]string, 0, len(agents))
	seen := map[string]struct{}{}
	for _, agent := range agents {
		name := agent.Name
		if name == "" {
			name = agent.TitleName
		}
		if name == "" {
			name = agent.TitleID
		}
		if name == "" {
			name = agent.Selector
		}
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		return strings.ToLower(names[i]) < strings.ToLower(names[j])
	})
	return joinLimited(names, limit)
}

func joinLimited(values []string, limit int) string {
	if len(values) == 0 {
		return ""
	}
	if limit <= 0 || len(values) <= limit {
		return strings.Join(values, ", ")
	}
	return strings.Join(values[:limit], ", ") + fmt.Sprintf(", +%d more", len(values)-limit)
}

func formatAgentChoice(agent agentInfo) string {
	name := agent.Name
	if name == "" {
		name = agent.TitleName
	}
	if name == "" {
		name = agent.TitleID
	}
	if name == "" {
		name = agent.Selector
	}
	if name == "" {
		return ""
	}
	if agent.Selector == "" || agent.Selector == name {
		return name
	}
	if agent.SharedSelector {
		return fmt.Sprintf("%s [%s, shared]", name, agent.Selector)
	}
	return fmt.Sprintf("%s [%s]", name, agent.Selector)
}

func suggestAgents(query string, agents []agentInfo, limit int) []string {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return nil
	}

	type suggestion struct {
		label string
		score int
	}

	best := map[string]int{}
	for _, agent := range agents {
		label := formatAgentChoice(agent)
		if label == "" {
			continue
		}
		score := suggestionScore(query, agent)
		if score == 0 {
			continue
		}
		if prev, ok := best[label]; !ok || score > prev {
			best[label] = score
		}
	}

	suggestions := make([]suggestion, 0, len(best))
	for label, score := range best {
		suggestions = append(suggestions, suggestion{label: label, score: score})
	}
	if len(suggestions) == 0 {
		return nil
	}

	sort.Slice(suggestions, func(i, j int) bool {
		if suggestions[i].score != suggestions[j].score {
			return suggestions[i].score > suggestions[j].score
		}
		return strings.ToLower(suggestions[i].label) < strings.ToLower(suggestions[j].label)
	})

	if limit > 0 && len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}

	labels := make([]string, 0, len(suggestions))
	for _, suggestion := range suggestions {
		labels = append(labels, suggestion.label)
	}
	return labels
}

func suggestionScore(query string, agent agentInfo) int {
	best := 0
	for _, candidate := range []string{agent.Name, agent.Selector, agent.TitleID} {
		candidate = strings.ToLower(candidate)
		if candidate == "" {
			continue
		}
		switch {
		case candidate == query:
			if best < 100 {
				best = 100
			}
		case strings.HasPrefix(candidate, query):
			if best < 80 {
				best = 80
			}
		case strings.Contains(candidate, query):
			if best < 60 {
				best = 60
			}
		}
	}
	if best == 0 {
		return 0
	}
	if agent.SharedSelector {
		best -= 1
	} else {
		best += 1
	}
	return best
}
