package discord

// Attachment Discord attachment object
type Attachment struct {
	Content   string
	Username  string
	AvatarURL string
	tts       bool
	Embeds    []Embed
}

// Embed Discord embed object
type Embed struct {
	Title       string
	Description string
	URL         string
	Timestamp   string
	Color       int16
	Footer      Footer
	Image       Image
	Thumbnail   Image
	Provider    Reference
	Author      Author
	Fields      []Field
}

// Footer Discord footer object
type Footer struct {
	Text         string
	IconURL      string
	ProxyIconURL string
}

// Reference Discord reference object
type Reference struct {
	Name string
	URL  string
}

// Author Discord authro object
type Author struct {
	*Reference
	IconURL      string
	ProxyIconURL string
}

// Image Discord image object
type Image struct {
	URL      string
	ProxyURL string
	Height   int16
	Width    int16
}

// Field Discord field object
type Field struct {
	Name   string
	Value  string
	Inline bool
}
