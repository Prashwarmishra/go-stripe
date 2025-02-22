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

type Order struct {
	ID            int    `json:"id"`
	WidgetID      int    `json:"widget_id"`
	TransactionID int    `json:"transaction_id"`
	StatusID      int    `json:"status_id"`
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
	TransactionStatusID int    `json:"transaction_status_id"`
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

	row := m.DB.QueryRowContext(ctx, `
		SELECT 
			id, name, description, coalesce(image, ''),
			inventory_level, price, created_at, 
			updated_at 
		FROM widgets 
		WHERE id = ?`, id)
	widget := Widget{}

	err := row.Scan(&widget.ID, &widget.Name,
		&widget.Description, &widget.Image, &widget.InventoryLevel,
		&widget.Price, &widget.CreatedAt, &widget.UpdatedAt)

	return widget, err
}

func (m *DBModel) InsertTransaction(txn Transaction) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	stmt := `INSERT INTO transactions (
		amount, currency, last_four, 
		bank_return_code, transaction_status_id,
		created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := m.DB.ExecContext(ctx, stmt,
		txn.Amount, txn.Currency, txn.LastFour,
		txn.BankReturnCode, txn.TransactionStatusID,
		time.Now(), time.Now())

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *DBModel) InsertOrder(order Order) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	stmt := `INSERT INTO orders VALUES (
		id, widget_id, transaction_id, status_id, 
		quantity, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := m.DB.ExecContext(ctx, stmt,
		order.WidgetID, order.TransactionID, order.StatusID,
		order.Quantity, time.Now(), time.Now())

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return 0, err
	}

	return int(id), nil
}
