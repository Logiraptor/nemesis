package push

// PushClient represents a device which has registered to receive push notifications.
type PushClient struct {
	OS    string
	Token string
}

type Client interface {
	SendMessage(string) error
}
