package store

type Store interface {
	AddDeadline(title string, datetime string) (Deadline, error)
	ListDeadlines() ([]Deadline, error)
	DeleteDeadline(id int) error

	AddBasket(name string) error
	ListBaskets() ([]string, error)
	DeleteBasket(name string) error

	AddPin(basketName string, content string) (Pin, error)
	ListPins(basketName string) ([]Pin, error)
	DeletePin(basketName string, id int) error
}
