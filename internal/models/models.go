package models

import (
	"context"
	"database/sql"
	"time"
)

type DBModel struct {
	DB *sql.DB
}

type Models struct {
	DBModel
}

type Widget struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Price          int    `json:"price"`
	InventoryLevel int    `json:"inventory_level"`
	CreatedAt      string `json:"-"`
	UpdatedAt      string `json:"-"`
}

func (m *DBModel) GetWidget(id int) (Widget, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	row := m.DB.QueryRowContext(ctx, "SELECT id, name FROM widgets WHERE id = ?", id)
	widget := Widget{}
	err := row.Scan(&widget.ID, &widget.Name)
	return widget, err
}
