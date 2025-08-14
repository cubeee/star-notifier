package lib

type DiscordMessage struct {
	Files       *[]DiscordFile       `json:"-"`
	Content     string               `json:"content,omitempty"`
	Embeds      *[]DiscordEmbed      `json:"embeds,omitempty"`
	Attachments *[]DiscordAttachment `json:"attachments"`
}

type DiscordFile struct {
	Id   *string
	Name string
	Data *[]byte
}

type DiscordAttachment struct {
	Id          string `json:"id"`
	FileName    string `json:"filename"`
	Description string `json:"description"`
}

type DiscordEmbed struct {
	Title       string               `json:"title,omitempty"`
	Description string               `json:"description,omitempty"`
	Image       *DiscordEmbedImage   `json:"image,omitempty"`
	Fields      *[]DiscordEmbedField `json:"fields,omitempty"`
}

type DiscordEmbedImage struct {
	Url string `json:"url"`
}

type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}
