package texts

// Text is a type, which the surface gives out.
type Text struct{ Body string }

// Kind is a type, and a constant is one of its values.
type Kind int

// Plain is a constant of Kind.
const Plain Kind = 0

func Save() {} // want `Save is a free function on the surface of features/texts: make it a method of UseCases in texts_use_cases\.go, or unexport it as a helper`

var Default = Text{} // want `Default is a variable on the surface of features/texts, which gives out only UseCases and types`

// trim is a helper the surface leaves out.
func trim(text Text) Text { return text }

var _ = trim
