package main

import (
	"errors"
	"time"
)

type (
	todoModel struct {
		ID        uint64    `json:"id"`
		Title     string    `json:"title"`
		Completed bool      `json:"completed"`
		CreatedAt time.Time `json:"created_at"`
		Owner     string    `json:"owner"`
	}
)

func (m *todoModel) Pack() []interface{} {
	out := make([]interface{}, 0, 4)
	out = append(out, m.ID)
	out = append(out, m.Title)
	out = append(out, m.Completed)
	out = append(out, m.CreatedAt.Unix())
	out = append(out, m.Owner)
	return out
}

func (m *todoModel) Unpack(data []interface{}) error {
	if len(data) != 5 {
		return errors.New("bad data length")
	}

	var err bool

	m.ID, err = data[0].(uint64)
	if !err {
		return errors.New("can't convert todo.ID")
	}

	m.Title, err = data[1].(string)
	if !err {
		return errors.New("can't convert todo.Title")
	}

	m.Completed, err = data[2].(bool)
	if !err {
		return errors.New("can't convert todo.Completed")
	}

	t, err := data[3].(uint64)
	if !err {
		return errors.New("can't convert todo.CreatedAt")
	}
	m.CreatedAt = time.Unix(int64(t), 0)

	m.Owner, err = data[4].(string)
	if !err {
		return errors.New("can't convert todo.Owner")
	}

	return nil
}
