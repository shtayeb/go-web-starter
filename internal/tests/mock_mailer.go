package tests

type MockMailer struct {
	sentEmails []SentEmail
}

type SentEmail struct {
	Recipient    string
	TemplateFile string
	Data         interface{}
}

func NewMockMailer() *MockMailer {
	return &MockMailer{
		sentEmails: []SentEmail{},
	}
}

func (m *MockMailer) Send(recipient, templateFile string, data interface{}) error {
	m.sentEmails = append(m.sentEmails, SentEmail{
		Recipient:    recipient,
		TemplateFile: templateFile,
		Data:         data,
	})

	// Just record the email without actually sending it
	return nil
}

func (m *MockMailer) GetSentEmails() []SentEmail {
	return m.sentEmails
}

func (m *MockMailer) Clear() {
	m.sentEmails = nil
}

func (m *MockMailer) LastEmail() *SentEmail {
	if len(m.sentEmails) == 0 {
		return nil
	}
	return &m.sentEmails[len(m.sentEmails)-1]
}

func (m *MockMailer) EmailCount() int {
	return len(m.sentEmails)
}
