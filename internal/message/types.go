package message

type CommitMessage struct {
	Subject   string
	BodyLines []string
	Trailers  []string
}

type ActionLine struct {
	Type        string
	Scope       string
	Description string
}

type ValidationError struct {
	Line    int    `json:"line"`
	Code    string `json:"code"`
	Message string `json:"message"`
}
