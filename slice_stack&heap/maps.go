package main

import "fmt"

type TagItems struct {
	Items map[string]map[string]struct{}
	Tags  map[string]map[string]struct{}
}

type TagResult struct {
	Items map[string][]string
	Tags  map[string][]string
}

func (t *TagItems) AddTag(item, tag string) {
	if t.Items == nil {
		t.Items = make(map[string]map[string]struct{})
	}
	if t.Items[item] == nil {
		t.Items[item] = make(map[string]struct{})
	}
	if t.Tags == nil {
		t.Tags = make(map[string]map[string]struct{})
	}
	if t.Tags[tag] == nil {
		t.Tags[tag] = make(map[string]struct{})
	}
	t.Items[item][tag] = struct{}{}
	t.Tags[tag][item] = struct{}{}
}

func (t *TagItems) GetAll() TagResult {
	result := TagResult{
		Items: make(map[string][]string),
		Tags:  make(map[string][]string),
	}
	for item, tags := range t.Items {
		for tag := range tags {
			result.Items[item] = append(result.Items[item], tag)
		}
	}
	for tag, items := range t.Tags {
		for item := range items {
			result.Tags[tag] = append(result.Tags[tag], item)
		}
	}
	return result
}
func (t *TagItems) View() {
	displayItems := t.GetAll()
	fmt.Println("for items")
	for item, tags := range displayItems.Items {
		fmt.Println(item, ":", tags)
	}
	fmt.Println()
	fmt.Println("for tags")
	for tag, items := range displayItems.Tags {
		fmt.Println(tag, ":", items)
	}
}
