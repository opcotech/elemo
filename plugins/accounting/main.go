package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	plugin "github.com/opcotech/elemo/sdk/plugin"
)

const reportPageSize = 1000
const defaultBudgetThreshold = 80

func init() {
	plugin.Register(handle)
}

func handle(req plugin.Request) ([]byte, error) {
	switch req.Function {
	case "start", "stop":
		return plugin.Reply(nil)
	case "onEvent":
		return onEvent(req)
	case "account.create":
		return accountCreate(req)
	case "account.list":
		return accountList(req)
	case "account.update":
		return accountUpdate(req)
	case "account.delete":
		return accountDelete(req)
	case "budget.create":
		return budgetCreate(req)
	case "budget.list":
		return budgetList(req)
	case "budget.update":
		return budgetUpdate(req)
	case "budget.delete":
		return budgetDelete(req)
	case "charge.set":
		return chargeSet(req)
	case "charge.clear":
		return chargeClear(req)
	case "charge.get":
		return chargeGet(req)
	case "budget.summary":
		return budgetSummary(req)
	case "report.get":
		return reportGet(req)
	default:
		return nil, fmt.Errorf("unknown function %s", req.Function)
	}
}

type graphNode struct {
	ID         string         `json:"id"`
	PluginID   string         `json:"plugin_id"`
	Kind       string         `json:"kind"`
	Properties map[string]any `json:"properties"`
	ParentID   string         `json:"parent_id"`
	ParentType string         `json:"parent_type"`
	CreatedAt  string         `json:"created_at"`
	UpdatedAt  string         `json:"updated_at"`
}

type graphRelation struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	From     string `json:"from"`
	FromType string `json:"from_type"`
	To       string `json:"to"`
	ToType   string `json:"to_type"`
}

type graphBinding struct {
	PluginID string `json:"plugin_id"`
	Kind     string `json:"kind"`
}

type budgetReport struct {
	Budget    graphNode   `json:"budget"`
	Entries   []graphNode `json:"entries"`
	Seconds   int         `json:"seconds"`
	Used      int         `json:"used"`
	Remaining int         `json:"remaining"`
}

type accountReport struct {
	Account graphNode      `json:"account"`
	Budgets []budgetReport `json:"budgets"`
}

type chargePayload struct {
	ParentID       string `json:"parentId"`
	ParentType     string `json:"parentType"`
	BudgetID       string `json:"budgetId"`
	IssueID        string `json:"issueId"`
	ProjectID      string `json:"projectId"`
	AccountID      string `json:"accountId"`
	OrganizationID string `json:"organizationId"`
}

type domainEvent struct {
	Type    string `json:"type"`
	Payload struct {
		PluginID   string `json:"plugin_id"`
		Kind       string `json:"kind"`
		ID         string `json:"id"`
		ParentID   string `json:"parent_id"`
		ParentType string `json:"parent_type"`
		ScopeID    string `json:"scope_id"`
	} `json:"payload"`
}

func hostNodes(req plugin.Request, payload map[string]any) ([]graphNode, error) {
	raw, err := plugin.Host("graph.nodes.list", req.ScopeID, payload)
	if err != nil {
		return nil, err
	}
	var nodes []graphNode
	if err := json.Unmarshal(raw, &nodes); err != nil {
		return nil, err
	}
	return nodes, nil
}

func hostRelations(req plugin.Request, payload map[string]any) ([]graphRelation, error) {
	raw, err := plugin.Host("graph.relations.list", req.ScopeID, payload)
	if err != nil {
		return nil, err
	}
	var rels []graphRelation
	if err := json.Unmarshal(raw, &rels); err != nil {
		return nil, err
	}
	return rels, nil
}

func createNode(req plugin.Request, kind, parentID, parentType string, properties map[string]any) (graphNode, error) {
	raw, err := plugin.Host("graph.nodes.create", req.ScopeID, map[string]any{
		"kind":       kind,
		"parentId":   parentID,
		"parentType": parentType,
		"properties": properties,
	})
	if err != nil {
		return graphNode{}, err
	}
	var node graphNode
	if err := json.Unmarshal(raw, &node); err != nil {
		return graphNode{}, err
	}
	return node, nil
}

func createRelation(req plugin.Request, kind, fromID, fromType, toID, toType string) error {
	_, err := plugin.Host("graph.relations.create", req.ScopeID, map[string]any{
		"kind":     kind,
		"fromId":   fromID,
		"fromType": fromType,
		"toId":     toID,
		"toType":   toType,
	})
	return err
}

func deleteOutgoing(req plugin.Request, kind, fromID, fromType string) error {
	rels, err := hostRelations(req, map[string]any{
		"kind":      kind,
		"nodeId":    fromID,
		"nodeType":  fromType,
		"direction": "outgoing",
	})
	if err != nil {
		return err
	}
	for _, rel := range rels {
		if _, err := plugin.Host("graph.relations.delete", req.ScopeID, map[string]any{"id": rel.ID}); err != nil {
			return err
		}
	}
	return nil
}

func timeSource(req plugin.Request) (graphBinding, bool) {
	raw, err := plugin.Host("plugin.config.get", req.ScopeID, nil)
	if err != nil || len(raw) == 0 {
		return graphBinding{}, false
	}
	var cfg map[string]json.RawMessage
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return graphBinding{}, false
	}
	body, ok := cfg["time_source"]
	if !ok || len(body) == 0 || string(body) == "null" {
		return graphBinding{}, false
	}
	var binding graphBinding
	if err := json.Unmarshal(body, &binding); err != nil || binding.PluginID == "" || binding.Kind == "" {
		return graphBinding{}, false
	}
	return binding, true
}

func intProp(props map[string]any, key string) int {
	if props == nil {
		return 0
	}
	switch n := props[key].(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case json.Number:
		v, _ := n.Int64()
		return int(v)
	case string:
		v, _ := strconv.Atoi(n)
		return v
	default:
		return 0
	}
}

func validBudgetThreshold(value int) bool {
	return value >= 1 && value <= 100
}

func decodePayload(req plugin.Request) chargePayload {
	var body chargePayload
	if len(req.Payload) > 0 {
		_ = json.Unmarshal(req.Payload, &body)
	}
	return body
}

func accountCreate(req plugin.Request) ([]byte, error) {
	body := decodePayload(req)
	if body.OrganizationID == "" {
		return nil, fmt.Errorf("organizationId is required")
	}
	var props struct {
		Code        string `json:"code"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Active      *bool  `json:"active"`
	}
	_ = json.Unmarshal(req.Payload, &props)
	active := true
	if props.Active != nil {
		active = *props.Active
	}
	node, err := createNode(req, "Account", body.OrganizationID, "Organization", map[string]any{
		"code":        props.Code,
		"name":        props.Name,
		"description": props.Description,
		"active":      active,
	})
	if err != nil {
		return nil, err
	}
	return plugin.Reply(node)
}

func accountList(req plugin.Request) ([]byte, error) {
	body := decodePayload(req)
	if body.OrganizationID == "" {
		return nil, fmt.Errorf("organizationId is required")
	}
	nodes, err := hostNodes(req, map[string]any{
		"kind": "Account", "scopeId": body.OrganizationID, "scopeType": "Organization",
	})
	if err != nil {
		return nil, err
	}
	return plugin.Reply(nodes)
}

func accountUpdate(req plugin.Request) ([]byte, error) {
	var body struct {
		ID         string         `json:"id"`
		Properties map[string]any `json:"properties"`
	}
	_ = json.Unmarshal(req.Payload, &body)
	if body.ID == "" {
		return nil, fmt.Errorf("id is required")
	}
	raw, err := plugin.Host("graph.nodes.update", req.ScopeID, map[string]any{
		"id": body.ID, "properties": body.Properties,
	})
	if err != nil {
		return nil, err
	}
	var node graphNode
	if err := json.Unmarshal(raw, &node); err != nil {
		return nil, err
	}
	return plugin.Reply(node)
}

func accountDelete(req plugin.Request) ([]byte, error) {
	var body struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(req.Payload, &body)
	if body.ID == "" {
		return nil, fmt.Errorf("id is required")
	}
	budgets, err := hostNodes(req, map[string]any{
		"kind": "Budget", "scopeId": body.ID, "scopeType": "Extension",
	})
	if err != nil {
		return nil, err
	}
	if len(budgets) > 0 {
		return nil, fmt.Errorf("account has budgets; delete them first")
	}
	if _, err := plugin.Host("graph.nodes.delete", req.ScopeID, map[string]any{"id": body.ID}); err != nil {
		return nil, err
	}
	return plugin.Reply(nil)
}

func budgetCreate(req plugin.Request) ([]byte, error) {
	body := decodePayload(req)
	if body.AccountID == "" {
		return nil, fmt.Errorf("accountId is required")
	}
	var props struct {
		Name        string `json:"name"`
		Seconds     int    `json:"seconds"`
		Threshold   int    `json:"threshold"`
		PeriodStart string `json:"period_start"`
		PeriodEnd   string `json:"period_end"`
	}
	_ = json.Unmarshal(req.Payload, &props)
	if props.Threshold == 0 {
		props.Threshold = defaultBudgetThreshold
	}
	if !validBudgetThreshold(props.Threshold) {
		return nil, fmt.Errorf("threshold must be between 1 and 100")
	}
	node, err := createNode(req, "Budget", body.AccountID, "Extension", map[string]any{
		"name":         props.Name,
		"seconds":      props.Seconds,
		"threshold":    props.Threshold,
		"period_start": props.PeriodStart,
		"period_end":   props.PeriodEnd,
	})
	if err != nil {
		return nil, err
	}
	if err := createRelation(req, "HAS_BUDGET", body.AccountID, "Extension", node.ID, "Extension"); err != nil {
		return nil, err
	}
	return plugin.Reply(node)
}

func budgetList(req plugin.Request) ([]byte, error) {
	body := decodePayload(req)
	if body.AccountID == "" {
		return nil, fmt.Errorf("accountId is required")
	}
	nodes, err := hostNodes(req, map[string]any{
		"kind": "Budget", "scopeId": body.AccountID, "scopeType": "Extension",
	})
	if err != nil {
		return nil, err
	}
	return plugin.Reply(nodes)
}

func budgetUpdate(req plugin.Request) ([]byte, error) {
	var body struct {
		ID         string         `json:"id"`
		Properties map[string]any `json:"properties"`
	}
	_ = json.Unmarshal(req.Payload, &body)
	if body.ID == "" {
		return nil, fmt.Errorf("id is required")
	}
	if _, ok := body.Properties["threshold"]; ok {
		if threshold := intProp(body.Properties, "threshold"); !validBudgetThreshold(threshold) {
			return nil, fmt.Errorf("threshold must be between 1 and 100")
		}
	}
	raw, err := plugin.Host("graph.nodes.update", req.ScopeID, map[string]any{
		"id": body.ID, "properties": body.Properties,
	})
	if err != nil {
		return nil, err
	}
	var node graphNode
	if err := json.Unmarshal(raw, &node); err != nil {
		return nil, err
	}
	return plugin.Reply(node)
}

func budgetDelete(req plugin.Request) ([]byte, error) {
	var body struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(req.Payload, &body)
	if body.ID == "" {
		return nil, fmt.Errorf("id is required")
	}
	if _, err := plugin.Host("graph.nodes.delete", req.ScopeID, map[string]any{"id": body.ID}); err != nil {
		return nil, err
	}
	return plugin.Reply(nil)
}

func firstCharge(req plugin.Request, kind, parentID, parentType string) (*graphNode, error) {
	nodes, err := hostNodes(req, map[string]any{
		"kind": kind, "scopeId": parentID, "scopeType": parentType,
	})
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, nil
	}
	return &nodes[0], nil
}

func billedKind(parentType string) (chargeKind, billedKind, chargedKind, chargedType string) {
	if parentType == "Issue" {
		return "IssueCharge", "ISSUE_BILLED_TO", "ISSUE_CHARGED_ON", "Issue"
	}
	return "ProjectCharge", "BILLED_TO", "CHARGED_ON", "Project"
}

func chargeSet(req plugin.Request) ([]byte, error) {
	body := decodePayload(req)
	if body.ParentID == "" || body.ParentType == "" || body.BudgetID == "" {
		return nil, fmt.Errorf("parentId, parentType, and budgetId are required")
	}
	chargeKind, billed, charged, chargedType := billedKind(body.ParentType)
	charge, err := firstCharge(req, chargeKind, body.ParentID, body.ParentType)
	if err != nil {
		return nil, err
	}
	if charge == nil {
		created, err := createNode(req, chargeKind, body.ParentID, body.ParentType, map[string]any{})
		if err != nil {
			return nil, err
		}
		charge = &created
		if err := createRelation(req, charged, created.ID, "Extension", body.ParentID, chargedType); err != nil {
			return nil, err
		}
	}
	if err := deleteOutgoing(req, billed, charge.ID, "Extension"); err != nil {
		return nil, err
	}
	if err := createRelation(req, billed, charge.ID, "Extension", body.BudgetID, "Extension"); err != nil {
		return nil, err
	}
	if err := backfillCounted(req, body.ParentID, body.ParentType, body.BudgetID); err != nil {
		return nil, err
	}
	return plugin.Reply(charge)
}

func chargeClear(req plugin.Request) ([]byte, error) {
	body := decodePayload(req)
	if body.ParentID == "" || body.ParentType == "" {
		return nil, fmt.Errorf("parentId and parentType are required")
	}
	chargeKind, _, _, _ := billedKind(body.ParentType)
	charge, err := firstCharge(req, chargeKind, body.ParentID, body.ParentType)
	if err != nil {
		return nil, err
	}
	if charge != nil {
		if _, err := plugin.Host("graph.nodes.delete", req.ScopeID, map[string]any{"id": charge.ID}); err != nil {
			return nil, err
		}
	}
	return plugin.Reply(nil)
}

func chargeGet(req plugin.Request) ([]byte, error) {
	body := decodePayload(req)
	if body.ParentID == "" || body.ParentType == "" {
		return nil, fmt.Errorf("parentId and parentType are required")
	}
	chargeKind, billed, _, _ := billedKind(body.ParentType)
	charge, err := firstCharge(req, chargeKind, body.ParentID, body.ParentType)
	if err != nil {
		return nil, err
	}
	if charge == nil {
		return plugin.Reply(map[string]any{"charge": nil, "budgetId": nil})
	}
	rels, err := hostRelations(req, map[string]any{
		"kind": billed, "nodeId": charge.ID, "nodeType": "Extension", "direction": "outgoing",
	})
	if err != nil {
		return nil, err
	}
	budgetID := ""
	if len(rels) > 0 {
		budgetID = rels[0].To
	}
	return plugin.Reply(map[string]any{"charge": charge, "budgetId": budgetID})
}

func resolveBudget(req plugin.Request, issueID string) (string, error) {
	issueCharge, err := firstCharge(req, "IssueCharge", issueID, "Issue")
	if err != nil {
		return "", err
	}
	if issueCharge != nil {
		rels, err := hostRelations(req, map[string]any{
			"kind": "ISSUE_BILLED_TO", "nodeId": issueCharge.ID, "nodeType": "Extension", "direction": "outgoing",
		})
		if err != nil {
			return "", err
		}
		if len(rels) > 0 {
			return rels[0].To, nil
		}
	}
	raw, err := plugin.Host("issues.get", req.ScopeID, map[string]any{"id": issueID})
	if err != nil {
		return "", err
	}
	var issue struct {
		ProjectID string `json:"projectId"`
	}
	if err := json.Unmarshal(raw, &issue); err != nil || issue.ProjectID == "" {
		return "", nil
	}
	projectCharge, err := firstCharge(req, "ProjectCharge", issue.ProjectID, "Project")
	if err != nil || projectCharge == nil {
		return "", err
	}
	rels, err := hostRelations(req, map[string]any{
		"kind": "BILLED_TO", "nodeId": projectCharge.ID, "nodeType": "Extension", "direction": "outgoing",
	})
	if err != nil || len(rels) == 0 {
		return "", err
	}
	return rels[0].To, nil
}

func replaceCountedAgainst(req plugin.Request, entryID, budgetID string) error {
	if err := deleteOutgoing(req, "COUNTED_AGAINST", entryID, "Extension"); err != nil {
		return err
	}
	return createRelation(req, "COUNTED_AGAINST", entryID, "Extension", budgetID, "Extension")
}

func backfillCounted(req plugin.Request, parentID, parentType, budgetID string) error {
	binding, ok := timeSource(req)
	if !ok {
		return nil
	}
	issueIDs := []string{}
	if parentType == "Issue" {
		issueIDs = []string{parentID}
	} else {
		raw, err := plugin.Host("issues.list", req.ScopeID, map[string]any{"projectId": parentID})
		if err != nil {
			return err
		}
		var issues []struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(raw, &issues); err != nil {
			return err
		}
		for _, issue := range issues {
			issueIDs = append(issueIDs, issue.ID)
		}
	}
	for _, issueID := range issueIDs {
		nodes, err := hostNodes(req, map[string]any{
			"kind":          binding.Kind,
			"scopeId":       issueID,
			"scopeType":     "Issue",
			"ownerPluginId": binding.PluginID,
		})
		if err != nil {
			return err
		}
		for _, node := range nodes {
			if err := replaceCountedAgainst(req, node.ID, budgetID); err != nil {
				return err
			}
		}
	}
	return nil
}

func budgetSummary(req plugin.Request) ([]byte, error) {
	var body struct {
		BudgetID string `json:"budgetId"`
	}
	_ = json.Unmarshal(req.Payload, &body)
	if body.BudgetID == "" {
		return nil, fmt.Errorf("budgetId is required")
	}
	raw, err := plugin.Host("graph.nodes.get", req.ScopeID, map[string]any{"id": body.BudgetID})
	if err != nil {
		return nil, err
	}
	var budget graphNode
	if err := json.Unmarshal(raw, &budget); err != nil {
		return nil, err
	}
	envelope := intProp(budget.Properties, "seconds")
	binding, bound := timeSource(req)
	_, used, err := budgetUsage(req, body.BudgetID, binding, bound)
	if err != nil {
		return nil, err
	}
	remaining := envelope - used
	if remaining < 0 {
		remaining = 0
	}
	return plugin.Reply(map[string]any{
		"budgetId":  body.BudgetID,
		"seconds":   envelope,
		"used":      used,
		"remaining": remaining,
	})
}

func reportGet(req plugin.Request) ([]byte, error) {
	body := decodePayload(req)
	if body.OrganizationID == "" {
		return nil, fmt.Errorf("organizationId is required")
	}
	accounts, err := hostNodes(req, map[string]any{
		"kind":      "Account",
		"scopeId":   body.OrganizationID,
		"scopeType": "Organization",
		"pageSize":  reportPageSize,
	})
	if err != nil {
		return nil, err
	}
	binding, bound := timeSource(req)
	report := make([]accountReport, 0, len(accounts))
	for _, account := range accounts {
		budgets, err := hostNodes(req, map[string]any{
			"kind":      "Budget",
			"scopeId":   account.ID,
			"scopeType": "Extension",
			"pageSize":  reportPageSize,
		})
		if err != nil {
			return nil, err
		}
		accountBudgets := make([]budgetReport, 0, len(budgets))
		for _, budget := range budgets {
			entries, used, err := budgetUsage(req, budget.ID, binding, bound)
			if err != nil {
				return nil, err
			}
			seconds := intProp(budget.Properties, "seconds")
			remaining := seconds - used
			if remaining < 0 {
				remaining = 0
			}
			accountBudgets = append(accountBudgets, budgetReport{
				Budget:    budget,
				Entries:   entries,
				Seconds:   seconds,
				Used:      used,
				Remaining: remaining,
			})
		}
		report = append(report, accountReport{
			Account: account,
			Budgets: accountBudgets,
		})
	}
	return plugin.Reply(map[string]any{"accounts": report})
}

func budgetUsage(
	req plugin.Request,
	budgetID string,
	binding graphBinding,
	bound bool,
) ([]graphNode, int, error) {
	rels, err := hostRelations(req, map[string]any{
		"kind":      "COUNTED_AGAINST",
		"nodeId":    budgetID,
		"nodeType":  "Extension",
		"direction": "incoming",
		"pageSize":  reportPageSize,
	})
	if err != nil {
		return nil, 0, err
	}
	entries := make([]graphNode, 0, len(rels))
	used := 0
	for _, rel := range rels {
		payload := map[string]any{"id": rel.From}
		if bound {
			payload["ownerPluginId"] = binding.PluginID
		}
		nodeRaw, err := plugin.Host("graph.nodes.get", req.ScopeID, payload)
		if err != nil {
			continue
		}
		var node graphNode
		if err := json.Unmarshal(nodeRaw, &node); err != nil {
			continue
		}
		entries = append(entries, node)
		used += intProp(node.Properties, "seconds")
	}
	return entries, used, nil
}

func onEvent(req plugin.Request) ([]byte, error) {
	var evt domainEvent
	if err := json.Unmarshal(req.Payload, &evt); err != nil {
		return plugin.Reply(nil)
	}
	binding, ok := timeSource(req)
	if !ok || evt.Payload.PluginID != binding.PluginID || evt.Payload.Kind != binding.Kind {
		return plugin.Reply(nil)
	}
	if evt.Payload.ParentType != "Issue" || evt.Payload.ID == "" {
		return plugin.Reply(nil)
	}
	if evt.Type == "extension.deleted" {
		_ = deleteOutgoing(req, "COUNTED_AGAINST", evt.Payload.ID, "Extension")
		return plugin.Reply(nil)
	}
	budgetID, err := resolveBudget(req, evt.Payload.ParentID)
	if err != nil || budgetID == "" {
		return plugin.Reply(nil)
	}
	if err := replaceCountedAgainst(req, evt.Payload.ID, budgetID); err != nil {
		return nil, err
	}
	return plugin.Reply(nil)
}

func main() {}
