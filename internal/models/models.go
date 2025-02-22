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
	Image          string `json:"image"`
	Description    string `json:"description"`
	Price          int    `json:"price"`
	InventoryLevel int    `json:"inventory_level"`
	CreatedAt      string `json:"-"`
	UpdatedAt      string `json:"-"`
}

type Orders struct {
	ID            int    `json:"id"`
	WidgetId      int    `json:"widget_id"`
	TransactionId int    `json:"transaction_id"`
	StatusId      int    `json:"status_id"`
	Quantity      int    `json:"quantity"`
	CreatedAt     string `json:"-"`
	UpdatedAt     string `json:"-"`
}

type Status struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"-"`
	UpdatedAt string `json:"-"`
}

type TransactionStatus struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"-"`
	UpdatedAt string `json:"-"`
}

type Transaction struct {
	ID                  int    `json:"id"`
	Amount              int    `json:"amount"`
	Currency            string `json:"currency"`
	LastFour            string `json:"last_four"`
	BankReturnCode      string `json:"bank_return_code"`
	TransactionStatusId int    `json:"transaction_status_id"`
	CreatedAt           string `json:"-"`
	UpdatedAt           string `json:"-"`
}

type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	CreatedAt string `json:"-"`
	UpdatedAt string `json:"-"`
}

func (m *DBModel) GetWidget(id int) (Widget, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	row := m.DB.QueryRowContext(ctx, "SELECT id, name FROM widgets WHERE id = ?", id)
	widget := Widget{}
	err := row.Scan(&widget.ID, &widget.Name)
	return widget, err
}
