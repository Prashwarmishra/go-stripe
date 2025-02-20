package models

import "database/sql"

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
