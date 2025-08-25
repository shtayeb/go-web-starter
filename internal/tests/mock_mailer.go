package tests

// MockMailer is a test implementation that tracks sent emails
type MockMailer struct {
	sentEmails []SentEmail
}

// SentEmail represents an email that was sent during testing
type SentEmail struct {
	Recipient    string
	TemplateFile string
	Data         interface{}
}

// NewMockMailer creates a new mock mailer for testing
func NewMockMailer() *MockMailer {
	return &MockMailer{
		sentEmails: []SentEmail{},
	}
}

// Send implements the same interface as mailer.Mailer.Send for testing
func (m *MockMailer) Send(recipient, templateFile string, data interface{}) error {
	m.sentEmails = append(m.sentEmails, SentEmail{
		Recipient:    recipient,
		TemplateFile: templateFile,
		Data:         data,
	})
	// Just record the email without actually sending it
	return nil
}

// GetSentEmails returns all emails that were sent during the test
func (m *MockMailer) GetSentEmails() []SentEmail {
	return m.sentEmails
}

// Clear removes all sent emails
func (m *MockMailer) Clear() {
	m.sentEmails = nil
}

// LastEmail returns the last email that was sent, or nil if no emails were sent
func (m *MockMailer) LastEmail() *SentEmail {
	if len(m.sentEmails) == 0 {
		return nil
	}
	return &m.sentEmails[len(m.sentEmails)-1]
}

// EmailCount returns the number of emails sent
func (m *MockMailer) EmailCount() int {
	return len(m.sentEmails)
}
