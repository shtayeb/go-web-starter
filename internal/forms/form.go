package forms

import "go-web-starter/internal/forms/validator"

type MessageType string

const (
	MessageTypeSuccess MessageType = "success"
	MessageTypeError   MessageType = "error"
	MessageTypeWarning MessageType = "warning"
	MessageTypeInfo    MessageType = "info"
)

type Message struct {
	Text string      `json:"text"`
	Type MessageType `json:"type"`
}

type Form struct {
	validator.Validator `form:"-"`
	Message             *Message `form:"-"`
}

// Helpers methods for form
func (f *Form) SetMessage(text string, msgType MessageType) {
	f.Message = &Message{
		Text: text,
		Type: msgType,
	}
}

func (f *Form) ClearMessage() {
	f.Message = nil
}

func (f *Form) HasMessage() bool {
	return f.Message != nil
}
