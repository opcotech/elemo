package main

import (
	"encoding/json"
	"fmt"
	"time"

	plugin "github.com/opcotech/elemo/sdk/plugin"
)

func init() {
	plugin.Register(handle)
}

func handle(req plugin.Request) ([]byte, error) {
	switch req.Function {
	case "start", "stop", "onEvent":
		return plugin.Reply(nil)
	case "timer.start":
		return timerStart(req)
	case "timer.stop":
		return timerStop(req)
	case "timer.status":
		return timerStatus(req)
	case "timer.log":
		return timerLog(req)
	default:
		return nil, fmt.Errorf("unknown function %s", req.Function)
	}
}

type timerPayload struct {
	IssueID string `json:"issueId"`
	Note    string `json:"note"`
	Seconds int    `json:"seconds"`
}

type runningTimer struct {
	StartedAt int64  `json:"startedAt"`
	IssueID   string `json:"issueId"`
	UserID    string `json:"userId"`
}

func issueID(req plugin.Request) (string, error) {
	var body timerPayload
	if len(req.Payload) > 0 {
		_ = json.Unmarshal(req.Payload, &body)
	}
	if body.IssueID != "" {
		return body.IssueID, nil
	}
	if req.ScopeID == "" {
		return "", fmt.Errorf("issue id is required")
	}
	_, id, ok := splitComposite(req.ScopeID)
	if ok {
		return id, nil
	}
	return req.ScopeID, nil
}

func splitComposite(value string) (string, string, bool) {
	for i := 0; i < len(value); i++ {
		if value[i] == ':' {
			return value[:i], value[i+1:], true
		}
	}
	return "", "", false
}

func timerKey(userID, issueID string) string {
	if userID == "" {
		return "timer:" + issueID
	}
	return "timer:" + userID + ":" + issueID
}

func createTimeEntry(req plugin.Request, issueID string, seconds int, note string) (string, error) {
	if seconds < 1 {
		seconds = 1
	}
	props := map[string]any{
		"seconds": seconds,
	}
	if note != "" {
		props["note"] = note
	}
	created, err := plugin.Host("graph.nodes.create", req.ScopeID, map[string]any{
		"kind":       "TimeEntry",
		"parentId":   issueID,
		"parentType": "Issue",
		"properties": props,
	})
	if err != nil {
		return "", err
	}
	var node struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(created, &node); err != nil {
		return "", err
	}
	if err := createRelation(req, "LOGGED_ON", node.ID, issueID, "Issue"); err != nil {
		return "", err
	}
	if req.UserID != "" {
		if err := createRelation(req, "LOGGED_BY", node.ID, req.UserID, "User"); err != nil {
			return "", err
		}
	}
	return node.ID, nil
}

func createRelation(req plugin.Request, kind, fromID, toID, toType string) error {
	_, err := plugin.Host("graph.relations.create", req.ScopeID, map[string]any{
		"kind":     kind,
		"fromId":   fromID,
		"fromType": "Extension",
		"toId":     toID,
		"toType":   toType,
	})
	return err
}

func timerStart(req plugin.Request) ([]byte, error) {
	id, err := issueID(req)
	if err != nil {
		return nil, err
	}
	_, err = plugin.Host("plugin.storage.set", req.ScopeID, map[string]any{
		"key": timerKey(req.UserID, id),
		"value": runningTimer{
			StartedAt: time.Now().Unix(),
			IssueID:   id,
			UserID:    req.UserID,
		},
	})
	if err != nil {
		return nil, err
	}
	return plugin.Reply(map[string]any{"running": true, "issueId": id})
}

func timerStop(req plugin.Request) ([]byte, error) {
	id, err := issueID(req)
	if err != nil {
		return nil, err
	}
	var body timerPayload
	if len(req.Payload) > 0 {
		_ = json.Unmarshal(req.Payload, &body)
	}
	raw, err := plugin.Host("plugin.storage.get", req.ScopeID, map[string]any{
		"key": timerKey(req.UserID, id),
	})
	if err != nil {
		return nil, err
	}
	var running runningTimer
	if err := json.Unmarshal(raw, &running); err != nil {
		return nil, err
	}
	seconds := int(time.Now().Unix() - running.StartedAt)
	entryID, err := createTimeEntry(req, id, seconds, body.Note)
	if err != nil {
		return nil, err
	}
	_, _ = plugin.Host("plugin.storage.delete", req.ScopeID, map[string]any{
		"key": timerKey(req.UserID, id),
	})
	return plugin.Reply(map[string]any{
		"running": false,
		"seconds": seconds,
		"entryId": entryID,
	})
}

func timerLog(req plugin.Request) ([]byte, error) {
	id, err := issueID(req)
	if err != nil {
		return nil, err
	}
	var body timerPayload
	if len(req.Payload) > 0 {
		_ = json.Unmarshal(req.Payload, &body)
	}
	entryID, err := createTimeEntry(req, id, body.Seconds, body.Note)
	if err != nil {
		return nil, err
	}
	return plugin.Reply(map[string]any{
		"seconds": body.Seconds,
		"entryId": entryID,
	})
}

func timerStatus(req plugin.Request) ([]byte, error) {
	id, err := issueID(req)
	if err != nil {
		return nil, err
	}
	raw, err := plugin.Host("plugin.storage.get", req.ScopeID, map[string]any{
		"key": timerKey(req.UserID, id),
	})
	if err != nil {
		return plugin.Reply(map[string]any{"running": false, "issueId": id})
	}
	var running runningTimer
	if err := json.Unmarshal(raw, &running); err != nil {
		return plugin.Reply(map[string]any{"running": false, "issueId": id})
	}
	return plugin.Reply(map[string]any{
		"running":   true,
		"issueId":   id,
		"startedAt": running.StartedAt,
		"elapsed":   time.Now().Unix() - running.StartedAt,
	})
}

func main() {}
