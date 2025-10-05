package service

type IRenderer interface {
	reload() error
	Render(name string, data any) (htmlBody, textBody string, err error)
}
