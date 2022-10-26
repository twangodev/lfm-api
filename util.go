package lfm_api

import (
	"golang.org/x/exp/slices"
	"golang.org/x/net/html"
)

var EmptyScrobble = Scrobble{
	Active: false,
}

func searchHTMLAttribute(attributes []html.Attribute, name string) string {
	index := slices.IndexFunc(attributes, func(attribute html.Attribute) bool { return attribute.Key == name })
	if index < 0 {
		return ""
	}
	return attributes[index].Val
}
