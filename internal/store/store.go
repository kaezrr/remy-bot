package store

import "errors"

type Deadline struct {
	ID    int
	Title string
	Date  string
}

type Pin struct {
	ID      int
	Content string
}

type Basket struct {
	Name      string
	Pins      []Pin
	NextPinID int
}

type Store struct {
	deadlines      []Deadline
	nextDeadlineID int
	baskets        map[string]*Basket
}

func New() *Store {
	return &Store{
		deadlines:      make([]Deadline, 0),
		nextDeadlineID: 1,
		baskets:        make(map[string]*Basket),
	}
}

func (s *Store) AddDeadline(title string, date string) Deadline {
	deadline := Deadline{
		ID:    s.nextDeadlineID,
		Title: title,
		Date:  date,
	}

	s.nextDeadlineID += 1
	s.deadlines = append(s.deadlines, deadline)
	return deadline
}

func (s *Store) ListDeadlines() []Deadline {
	return s.deadlines
}

func (s *Store) DeleteDeadline(id int) error {
	for i, d := range s.deadlines {
		if d.ID != id {
			continue
		}
		s.deadlines = append(s.deadlines[:i], s.deadlines[i+1:]...)
		return nil
	}

	return errors.New("deadline does not exist")
}

func (s *Store) CreateBasket(name string) error {
	_, ok := s.baskets[name]

	if ok {
		return errors.New("basket already exists")
	}

	s.baskets[name] = &Basket{
		Name:      name,
		NextPinID: 1,
		Pins:      make([]Pin, 0),
	}

	return nil
}

func (s *Store) ListBaskets() []string {
	baskets := make([]string, 0)

	for _, b := range s.baskets {
		baskets = append(baskets, b.Name)
	}

	return baskets
}

func (s *Store) AddPin(basketName string, content string) (Pin, error) {
	basket, ok := s.baskets[basketName]

	if !ok {
		return Pin{}, errors.New("basket does not exist")
	}

	pin := Pin{
		ID:      basket.NextPinID,
		Content: content,
	}

	basket.NextPinID += 1
	basket.Pins = append(basket.Pins, pin)
	return pin, nil
}

func (s *Store) ListPins(basketName string) ([]Pin, error) {
	basket, ok := s.baskets[basketName]

	if !ok {
		return nil, errors.New("basket does not exist")
	}

	return basket.Pins, nil
}
