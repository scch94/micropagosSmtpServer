package response

type SendEmailResponse struct {
	Status      string `json:"status"`
	ForwardRef  string `json:"forwardRef"`
	Description string `json:"description"`
}

func NewSendEmailResponse(status string, forwardRef string, description string) *SendEmailResponse {
	return &SendEmailResponse{
		status,
		forwardRef,
		description,
	}
}
