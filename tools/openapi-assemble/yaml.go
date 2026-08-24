package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func loadMappingFile(path string) (*yaml.Node, error) {
	data, err := os.ReadFile(path) // #nosec
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return nil, fmt.Errorf("parse %s: empty document", path)
	}

	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("parse %s: expected a YAML mapping", path)
	}

	return root, nil
}

func writeYAMLFile(path string, node *yaml.Node, header string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create directory for %s: %w", path, err)
	}

	encoded, err := marshalYAML(node)
	if err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}

	var buf bytes.Buffer
	if header != "" {
		buf.WriteString(header)
		if header[len(header)-1] != '\n' {
			buf.WriteByte('\n')
		}
		buf.WriteByte('\n')
	}
	buf.Write(encoded)

	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil { // #nosec
		return fmt.Errorf("write %s: %w", path, err)
	}

	return nil
}

func marshalYAML(node *yaml.Node) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(node); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}

	out := bytes.TrimPrefix(buf.Bytes(), []byte("---\n"))
	out = bytes.TrimPrefix(out, []byte("---\r\n"))
	return out, nil
}

func mappingLookup(n *yaml.Node, key string) (keyNode, valueNode *yaml.Node) {
	if n == nil || n.Kind != yaml.MappingNode {
		return nil, nil
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		if n.Content[i].Value == key {
			return n.Content[i], n.Content[i+1]
		}
	}
	return nil, nil
}

func mappingKeys(n *yaml.Node) []string {
	if n == nil || n.Kind != yaml.MappingNode {
		return nil
	}
	keys := make([]string, 0, len(n.Content)/2)
	for i := 0; i+1 < len(n.Content); i += 2 {
		keys = append(keys, n.Content[i].Value)
	}
	return keys
}

func deleteMappingKey(n *yaml.Node, key string) {
	if n == nil || n.Kind != yaml.MappingNode {
		return
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		if n.Content[i].Value == key {
			n.Content = append(n.Content[:i], n.Content[i+2:]...)
			return
		}
	}
}

func setMappingKey(n *yaml.Node, key string, value *yaml.Node) {
	if n == nil || n.Kind != yaml.MappingNode {
		return
	}
	if keyNode, _ := mappingLookup(n, key); keyNode != nil {
		deleteMappingKey(n, key)
	}
	n.Content = append(n.Content, &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: key,
	}, value)
}

func extractKeys(src *yaml.Node, keys []string, srcName string) (*yaml.Node, error) {
	out := &yaml.Node{Kind: yaml.MappingNode}
	for _, key := range keys {
		keyNode, valueNode := mappingLookup(src, key)
		if valueNode == nil {
			return nil, fmt.Errorf("missing key %q in %s", key, srcName)
		}
		out.Content = append(out.Content, keyNode, valueNode)
	}
	return out, nil
}

func mergeMappings(dst *yaml.Node, src *yaml.Node, srcPath string, origins map[string]string) error {
	if dst == nil || dst.Kind != yaml.MappingNode {
		return fmt.Errorf("merge destination is not a mapping")
	}
	if src == nil {
		return nil
	}
	if src.Kind != yaml.MappingNode {
		return fmt.Errorf("merge %s: expected a YAML mapping", srcPath)
	}

	for i := 0; i+1 < len(src.Content); i += 2 {
		key := src.Content[i].Value
		if prev, ok := origins[key]; ok {
			return fmt.Errorf("duplicate key %q: already defined in %s, also in %s", key, prev, srcPath)
		}
		dst.Content = append(dst.Content, src.Content[i], src.Content[i+1])
		origins[key] = srcPath
	}

	return nil
}

func mergeFiles(srcDir string, files []string) (*yaml.Node, error) {
	out := &yaml.Node{Kind: yaml.MappingNode}
	origins := make(map[string]string)

	for _, rel := range files {
		path := filepath.Join(srcDir, rel)
		node, err := loadMappingFile(path)
		if err != nil {
			return nil, err
		}
		if err := mergeMappings(out, node, rel, origins); err != nil {
			return nil, err
		}
	}

	return out, nil
}
