package lang

import "golang.org/x/text/language"

func Parse(tag string) (language.Tag, error) { return language.Parse(tag) }
