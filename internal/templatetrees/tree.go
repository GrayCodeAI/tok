package templatetrees

import (
	"strings"
)

type TreeNode struct {
	Key      string               `json:"key"`
	Value    interface{}          `json:"value,omitempty"`
	Children map[string]*TreeNode `json:"children,omitempty"`
	Parent   *TreeNode            `json:"-"`
}

type TemplateTree struct {
	root   *TreeNode
	fields map[string]interface{}
}

func NewTemplateTree() *TemplateTree {
	return &TemplateTree{
		root:   &TreeNode{Children: make(map[string]*TreeNode)},
		fields: make(map[string]interface{}),
	}
}

func (t *TemplateTree) Add(path string, value interface{}) {
	parts := strings.Split(path, ".")
	current := t.root
	for _, part := range parts {
		if _, ok := current.Children[part]; !ok {
			current.Children[part] = &TreeNode{
				Key:      part,
				Children: make(map[string]*TreeNode),
				Parent:   current,
			}
		}
		current = current.Children[part]
	}
	current.Value = value
}

func (t *TemplateTree) Get(path string) interface{} {
	parts := strings.Split(path, ".")
	current := t.root
	for _, part := range parts {
		child, ok := current.Children[part]
		if !ok {
			return nil
		}
		current = child
	}
	return current.Value
}

func (t *TemplateTree) SetField(key string, value interface{}) {
	t.fields[key] = value
}

func (t *TemplateTree) GetField(key string) interface{} {
	return t.fields[key]
}

func (t *TemplateTree) CarryForward(keys []string) map[string]interface{} {
	result := make(map[string]interface{})
	for _, key := range keys {
		if val, ok := t.fields[key]; ok {
			result[key] = val
		}
	}
	return result
}

func (t *TemplateTree) Render(template string) string {
	result := template
	for key, value := range t.fields {
		result = strings.ReplaceAll(result, "{{"+key+"}}", toString(value))
	}
	return result
}

func (t *TemplateTree) ChildrenAs(path, delimiter string) string {
	parts := strings.Split(path, ".")
	current := t.root
	for _, part := range parts {
		child, ok := current.Children[part]
		if !ok {
			return ""
		}
		current = child
	}

	var items []string
	for key, child := range current.Children {
		if child.Value != nil {
			items = append(items, key+"="+toString(child.Value))
		} else {
			items = append(items, key)
		}
	}
	return strings.Join(items, delimiter)
}

func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return string(rune(val + '0'))
	case float64:
		return string(rune(int(val) + '0'))
	default:
		return ""
	}
}

type StructuredCollection struct {
	Name   string        `json:"name"`
	Items  []interface{} `json:"items"`
	Fields []string      `json:"fields"`
}

func NewStructuredCollection(name string, fields []string) *StructuredCollection {
	return &StructuredCollection{
		Name:   name,
		Fields: fields,
	}
}

func (c *StructuredCollection) Add(item interface{}) {
	c.Items = append(c.Items, item)
}

func (c *StructuredCollection) Render(template string) string {
	var results []string
	for _, item := range c.Items {
		result := template
		if m, ok := item.(map[string]interface{}); ok {
			for key, val := range m {
				result = strings.ReplaceAll(result, "{{"+key+"}}", toString(val))
			}
		}
		results = append(results, result)
	}
	return strings.Join(results, "\n")
}
