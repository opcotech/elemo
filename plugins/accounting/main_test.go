package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	plugin "github.com/opcotech/elemo/sdk/plugin"
)

func TestIntProp(t *testing.T) {
	t.Parallel()
	if got := intProp(map[string]any{"seconds": float64(90)}, "seconds"); got != 90 {
		t.Fatalf("intProp float64 = %d, want 90", got)
	}
	if got := intProp(map[string]any{"seconds": json.Number("15")}, "seconds"); got != 15 {
		t.Fatalf("intProp json.Number = %d, want 15", got)
	}
}

func TestAccountCreatePersistsDescription(t *testing.T) {
	prev := plugin.HostFn
	t.Cleanup(func() { plugin.HostFn = prev })

	var properties map[string]any
	plugin.HostFn = func(method, _ string, payload any) (json.RawMessage, error) {
		if method != "graph.nodes.create" {
			t.Fatalf("unexpected host method %s", method)
		}
		properties = asMap(payload)["properties"].(map[string]any)
		return mustJSON(graphNode{ID: "account-1", Kind: "Account"}), nil
	}

	_, err := handle(plugin.Request{
		Function: "account.create",
		ScopeID:  "Organization:org",
		Payload: mustJSON(map[string]any{
			"organizationId": "org",
			"code":           "CONS",
			"name":           "Consulting",
			"description":    "Client delivery work",
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if properties["description"] != "Client delivery work" {
		t.Fatalf("description = %v", properties["description"])
	}
}

func TestBudgetCreatePersistsThreshold(t *testing.T) {
	prev := plugin.HostFn
	t.Cleanup(func() { plugin.HostFn = prev })

	var properties map[string]any
	plugin.HostFn = func(method, _ string, payload any) (json.RawMessage, error) {
		switch method {
		case "graph.nodes.create":
			properties = asMap(payload)["properties"].(map[string]any)
			return mustJSON(graphNode{ID: "budget-1", Kind: "Budget"}), nil
		case "graph.relations.create":
			return mustJSON(graphRelation{ID: "relation-1"}), nil
		default:
			t.Fatalf("unexpected host method %s", method)
			return nil, nil
		}
	}

	_, err := handle(plugin.Request{
		Function: "budget.create",
		ScopeID:  "Organization:org",
		Payload: mustJSON(map[string]any{
			"accountId": "account-1",
			"name":      "Q1",
			"seconds":   3600,
			"threshold": 80,
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if properties["threshold"] != 80 {
		t.Fatalf("threshold = %v", properties["threshold"])
	}
}

func TestOnEventSkipsWhenTimeSourceUnbound(t *testing.T) {
	prev := plugin.HostFn
	t.Cleanup(func() { plugin.HostFn = prev })
	plugin.HostFn = func(method, _ string, _ any) (json.RawMessage, error) {
		if method == "plugin.config.get" {
			return json.RawMessage(`{}`), nil
		}
		t.Fatalf("unexpected host method %s", method)
		return nil, nil
	}

	_, err := handle(plugin.Request{
		Function: "onEvent",
		ScopeID:  "Organization:org",
		Payload: mustJSON(map[string]any{
			"type": "extension.created",
			"payload": map[string]any{
				"plugin_id":   "com.elemo.timetracking",
				"kind":        "TimeEntry",
				"id":          "ext-1",
				"parent_id":   "issue-1",
				"parent_type": "Issue",
			},
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestOnEventCreatesCountedAgainst(t *testing.T) {
	prev := plugin.HostFn
	t.Cleanup(func() { plugin.HostFn = prev })

	var created []map[string]any
	plugin.HostFn = func(method, _ string, payload any) (json.RawMessage, error) {
		switch method {
		case "plugin.config.get":
			return json.RawMessage(`{"time_source":{"plugin_id":"com.elemo.timetracking","kind":"TimeEntry"}}`), nil
		case "graph.nodes.list":
			body := asMap(payload)
			if body["kind"] == "IssueCharge" {
				return json.RawMessage(`[]`), nil
			}
			if body["kind"] == "ProjectCharge" {
				return mustJSON([]graphNode{{ID: "charge-1", Kind: "ProjectCharge"}}), nil
			}
			return json.RawMessage(`[]`), nil
		case "issues.get":
			return json.RawMessage(`{"id":"issue-1","projectId":"project-1"}`), nil
		case "graph.relations.list":
			body := asMap(payload)
			if body["kind"] == "BILLED_TO" {
				return mustJSON([]graphRelation{{
					ID: "rel-billed", Kind: "BILLED_TO", From: "charge-1", To: "budget-1",
				}}), nil
			}
			if body["kind"] == "COUNTED_AGAINST" {
				return json.RawMessage(`[]`), nil
			}
			return json.RawMessage(`[]`), nil
		case "graph.relations.create":
			body := asMap(payload)
			created = append(created, body)
			return mustJSON(graphRelation{ID: "rel-new", Kind: fmt.Sprint(body["kind"])}), nil
		case "graph.relations.delete":
			return json.RawMessage(`null`), nil
		default:
			t.Fatalf("unexpected host method %s", method)
			return nil, nil
		}
	}

	_, err := handle(plugin.Request{
		Function: "onEvent",
		ScopeID:  "Organization:org",
		Payload: mustJSON(map[string]any{
			"type": "extension.created",
			"payload": map[string]any{
				"plugin_id":   "com.elemo.timetracking",
				"kind":        "TimeEntry",
				"id":          "entry-1",
				"parent_id":   "issue-1",
				"parent_type": "Issue",
			},
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(created) != 1 {
		t.Fatalf("created %d relations, want 1", len(created))
	}
	if created[0]["kind"] != "COUNTED_AGAINST" {
		t.Fatalf("kind = %v, want COUNTED_AGAINST", created[0]["kind"])
	}
	if created[0]["fromId"] != "entry-1" || created[0]["toId"] != "budget-1" {
		t.Fatalf("unexpected relation ends %#v", created[0])
	}
}

func TestBudgetSummarySumsIncoming(t *testing.T) {
	prev := plugin.HostFn
	t.Cleanup(func() { plugin.HostFn = prev })
	plugin.HostFn = func(method, _ string, payload any) (json.RawMessage, error) {
		switch method {
		case "plugin.config.get":
			return json.RawMessage(`{"time_source":{"plugin_id":"com.elemo.timetracking","kind":"TimeEntry"}}`), nil
		case "graph.nodes.get":
			body := asMap(payload)
			if body["id"] == "budget-1" {
				return mustJSON(graphNode{
					ID: "budget-1", Kind: "Budget",
					Properties: map[string]any{"seconds": float64(3600)},
				}), nil
			}
			return mustJSON(graphNode{
				ID: fmt.Sprint(body["id"]), Kind: "TimeEntry",
				Properties: map[string]any{"seconds": float64(900)},
			}), nil
		case "graph.relations.list":
			return mustJSON([]graphRelation{{
				ID: "rel-1", Kind: "COUNTED_AGAINST", From: "entry-1", To: "budget-1",
			}}), nil
		default:
			t.Fatalf("unexpected host method %s", method)
			return nil, nil
		}
	}

	raw, err := handle(plugin.Request{
		Function: "budget.summary",
		ScopeID:  "Organization:org",
		Payload:  mustJSON(map[string]any{"budgetId": "budget-1"}),
	})
	if err != nil {
		t.Fatal(err)
	}
	var resp plugin.Response
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatal(err)
	}
	var out struct {
		Used      int `json:"used"`
		Seconds   int `json:"seconds"`
		Remaining int `json:"remaining"`
	}
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		t.Fatal(err)
	}
	if out.Used != 900 || out.Seconds != 3600 || out.Remaining != 2700 {
		t.Fatalf("summary = %+v", out)
	}
}

func TestReportGetReturnsAccountUsageDetails(t *testing.T) {
	prev := plugin.HostFn
	t.Cleanup(func() { plugin.HostFn = prev })
	plugin.HostFn = func(method, _ string, payload any) (json.RawMessage, error) {
		switch method {
		case "plugin.config.get":
			return json.RawMessage(
				`{"time_source":{"plugin_id":"com.elemo.timetracking","kind":"TimeEntry"}}`,
			), nil
		case "graph.nodes.list":
			body := asMap(payload)
			if body["pageSize"] != reportPageSize {
				t.Fatalf("pageSize = %v, want %d", body["pageSize"], reportPageSize)
			}
			switch body["kind"] {
			case "Account":
				return mustJSON([]graphNode{
					{
						ID:         "account-1",
						Kind:       "Account",
						Properties: map[string]any{"code": "CONS", "name": "Consulting"},
					},
					{
						ID:         "account-2",
						Kind:       "Account",
						Properties: map[string]any{"code": "OPS", "name": "Operations"},
					},
				}), nil
			case "Budget":
				if body["scopeId"] == "account-2" {
					return json.RawMessage(`[]`), nil
				}
				return mustJSON([]graphNode{{
					ID:   "budget-1",
					Kind: "Budget",
					Properties: map[string]any{
						"name":    "September",
						"seconds": float64(7200),
					},
				}}), nil
			default:
				t.Fatalf("unexpected node kind %v", body["kind"])
				return nil, nil
			}
		case "graph.relations.list":
			body := asMap(payload)
			if body["kind"] != "COUNTED_AGAINST" || body["direction"] != "incoming" {
				t.Fatalf("unexpected relation query %#v", body)
			}
			if body["pageSize"] != reportPageSize {
				t.Fatalf("pageSize = %v, want %d", body["pageSize"], reportPageSize)
			}
			return mustJSON([]graphRelation{
				{
					ID:   "counted-1",
					Kind: "COUNTED_AGAINST",
					From: "entry-1",
					To:   "budget-1",
				},
				{
					ID:   "counted-missing",
					Kind: "COUNTED_AGAINST",
					From: "entry-missing",
					To:   "budget-1",
				},
			}), nil
		case "graph.nodes.get":
			body := asMap(payload)
			if body["ownerPluginId"] != "com.elemo.timetracking" {
				t.Fatalf("ownerPluginId = %v", body["ownerPluginId"])
			}
			if body["id"] == "entry-missing" {
				return nil, fmt.Errorf("entry was deleted")
			}
			return mustJSON(graphNode{
				ID:         "entry-1",
				PluginID:   "com.elemo.timetracking",
				Kind:       "TimeEntry",
				ParentID:   "issue-1",
				ParentType: "Issue",
				CreatedAt:  "2026-09-01T08:30:00Z",
				Properties: map[string]any{
					"seconds": float64(1800),
					"note":    "Discovery",
					"user_id": "user-1",
				},
			}), nil
		default:
			t.Fatalf("unexpected host method %s", method)
			return nil, nil
		}
	}

	raw, err := handle(plugin.Request{
		Function: "report.get",
		ScopeID:  "Organization:org",
		Payload:  mustJSON(map[string]any{"organizationId": "org"}),
	})
	if err != nil {
		t.Fatal(err)
	}
	var resp plugin.Response
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatal(err)
	}
	var out struct {
		Accounts []accountReport `json:"accounts"`
	}
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Accounts) != 2 {
		t.Fatalf("accounts = %d, want 2", len(out.Accounts))
	}
	if len(out.Accounts[1].Budgets) != 0 {
		t.Fatalf("empty account budgets = %d, want 0", len(out.Accounts[1].Budgets))
	}
	budget := out.Accounts[0].Budgets[0]
	if budget.Used != 1800 || budget.Seconds != 7200 || budget.Remaining != 5400 {
		t.Fatalf("budget totals = %+v", budget)
	}
	if len(budget.Entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(budget.Entries))
	}
	entry := budget.Entries[0]
	if entry.ParentID != "issue-1" || entry.CreatedAt != "2026-09-01T08:30:00Z" {
		t.Fatalf("entry metadata = %+v", entry)
	}
}

func TestReportGetWorksWithoutTimeSourceBinding(t *testing.T) {
	prev := plugin.HostFn
	t.Cleanup(func() { plugin.HostFn = prev })
	plugin.HostFn = func(method, _ string, payload any) (json.RawMessage, error) {
		switch method {
		case "plugin.config.get":
			return json.RawMessage(`{}`), nil
		case "graph.nodes.list":
			body := asMap(payload)
			if body["kind"] == "Account" {
				return mustJSON([]graphNode{{ID: "account-1", Kind: "Account"}}), nil
			}
			return mustJSON([]graphNode{{
				ID:         "budget-1",
				Kind:       "Budget",
				Properties: map[string]any{"seconds": float64(3600)},
			}}), nil
		case "graph.relations.list":
			return json.RawMessage(`[]`), nil
		default:
			t.Fatalf("unexpected host method %s", method)
			return nil, nil
		}
	}

	raw, err := handle(plugin.Request{
		Function: "report.get",
		ScopeID:  "Organization:org",
		Payload:  mustJSON(map[string]any{"organizationId": "org"}),
	})
	if err != nil {
		t.Fatal(err)
	}
	var resp plugin.Response
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatal(err)
	}
	var out struct {
		Accounts []accountReport `json:"accounts"`
	}
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		t.Fatal(err)
	}
	budget := out.Accounts[0].Budgets[0]
	if budget.Used != 0 || budget.Remaining != 3600 || len(budget.Entries) != 0 {
		t.Fatalf("budget = %+v", budget)
	}
}

func TestAccountUpdateRequiresID(t *testing.T) {
	t.Parallel()
	_, err := handle(plugin.Request{
		Function: "account.update",
		ScopeID:  "Organization:org",
		Payload:  mustJSON(map[string]any{"properties": map[string]any{"name": "X"}}),
	})
	if err == nil || !strings.Contains(err.Error(), "id is required") {
		t.Fatalf("err = %v", err)
	}
}

func TestBudgetDeleteRequiresID(t *testing.T) {
	t.Parallel()
	_, err := handle(plugin.Request{
		Function: "budget.delete",
		ScopeID:  "Organization:org",
		Payload:  mustJSON(map[string]any{}),
	})
	if err == nil || !strings.Contains(err.Error(), "id is required") {
		t.Fatalf("err = %v", err)
	}
}

func TestAccountDeleteBlockedWhenBudgetsExist(t *testing.T) {
	prev := plugin.HostFn
	t.Cleanup(func() { plugin.HostFn = prev })
	deleted := false
	plugin.HostFn = func(method, _ string, payload any) (json.RawMessage, error) {
		switch method {
		case "graph.nodes.list":
			body := asMap(payload)
			if body["kind"] != "Budget" {
				t.Fatalf("unexpected list kind %v", body["kind"])
			}
			return mustJSON([]graphNode{{ID: "budget-1", Kind: "Budget"}}), nil
		case "graph.nodes.delete":
			deleted = true
			return json.RawMessage(`null`), nil
		default:
			t.Fatalf("unexpected host method %s", method)
			return nil, nil
		}
	}

	_, err := handle(plugin.Request{
		Function: "account.delete",
		ScopeID:  "Organization:org",
		Payload:  mustJSON(map[string]any{"id": "acct-1"}),
	})
	if err == nil || !strings.Contains(err.Error(), "account has budgets") {
		t.Fatalf("err = %v", err)
	}
	if deleted {
		t.Fatal("must not delete an account that still has budgets")
	}
}

func TestAccountDeleteWhenEmpty(t *testing.T) {
	prev := plugin.HostFn
	t.Cleanup(func() { plugin.HostFn = prev })
	var deletedID string
	plugin.HostFn = func(method, _ string, payload any) (json.RawMessage, error) {
		switch method {
		case "graph.nodes.list":
			return json.RawMessage(`[]`), nil
		case "graph.nodes.delete":
			body := asMap(payload)
			deletedID = fmt.Sprint(body["id"])
			return json.RawMessage(`null`), nil
		default:
			t.Fatalf("unexpected host method %s", method)
			return nil, nil
		}
	}

	_, err := handle(plugin.Request{
		Function: "account.delete",
		ScopeID:  "Organization:org",
		Payload:  mustJSON(map[string]any{"id": "acct-1"}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if deletedID != "acct-1" {
		t.Fatalf("deleted %q, want acct-1", deletedID)
	}
}

func TestAccountUpdate(t *testing.T) {
	prev := plugin.HostFn
	t.Cleanup(func() { plugin.HostFn = prev })
	plugin.HostFn = func(method, _ string, payload any) (json.RawMessage, error) {
		if method != "graph.nodes.update" {
			t.Fatalf("unexpected host method %s", method)
		}
		body := asMap(payload)
		if body["id"] != "acct-1" {
			t.Fatalf("id = %v", body["id"])
		}
		return mustJSON(graphNode{
			ID:         "acct-1",
			Kind:       "Account",
			Properties: map[string]any{"code": "CONS", "name": "Consulting"},
		}), nil
	}

	raw, err := handle(plugin.Request{
		Function: "account.update",
		ScopeID:  "Organization:org",
		Payload: mustJSON(map[string]any{
			"id":         "acct-1",
			"properties": map[string]any{"code": "CONS", "name": "Consulting"},
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	var resp plugin.Response
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatal(err)
	}
	var node graphNode
	if err := json.Unmarshal(resp.Data, &node); err != nil {
		t.Fatal(err)
	}
	if node.ID != "acct-1" {
		t.Fatalf("node = %+v", node)
	}
}

func asMap(payload any) map[string]any {
	if body, ok := payload.(map[string]any); ok {
		return body
	}
	return map[string]any{}
}

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
