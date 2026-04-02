package prometheus

import (
	"fmt"
	"sync"
)

type Metric struct {
	Name   string            `json:"name"`
	Type   string            `json:"type"`
	Value  float64           `json:"value"`
	Labels map[string]string `json:"labels"`
}

type PrometheusCollector struct {
	metrics map[string]*Metric
	mu      sync.RWMutex
}

func NewPrometheusCollector() *PrometheusCollector {
	return &PrometheusCollector{
		metrics: make(map[string]*Metric),
	}
}

func (c *PrometheusCollector) Inc(name string, labels map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := c.makeKey(name, labels)
	if m, ok := c.metrics[key]; ok {
		m.Value++
	} else {
		c.metrics[key] = &Metric{Name: name, Type: "counter", Value: 1, Labels: labels}
	}
}

func (c *PrometheusCollector) Set(name string, value float64, labels map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := c.makeKey(name, labels)
	c.metrics[key] = &Metric{Name: name, Type: "gauge", Value: value, Labels: labels}
}

func (c *PrometheusCollector) Observe(name string, value float64, labels map[string]string) {
	c.Inc(name+"_total", labels)
	c.Set(name, value, labels)
}

func (c *PrometheusCollector) Export() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var result string
	for _, m := range c.metrics {
		labels := ""
		for k, v := range m.Labels {
			if labels != "" {
				labels += ","
			}
			labels += fmt.Sprintf("%s=\"%s\"", k, v)
		}
		if labels != "" {
			result += fmt.Sprintf("%s{%s} %v\n", m.Name, labels, m.Value)
		} else {
			result += fmt.Sprintf("%s %v\n", m.Name, m.Value)
		}
	}
	return result
}

func (c *PrometheusCollector) makeKey(name string, labels map[string]string) string {
	key := name
	for k, v := range labels {
		key += ":" + k + "=" + v
	}
	return key
}

func (c *PrometheusCollector) GetMetric(name string) *Metric {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, m := range c.metrics {
		if m.Name == name {
			return m
		}
	}
	return nil
}
