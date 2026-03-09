package repository

import (
	"database/sql"
	"errors"
	"restaurant-api/internal/models"
)

type MenuRepository struct {
	db *sql.DB
}

func NewMenuRepository(db *sql.DB) *MenuRepository {
	return &MenuRepository{db: db}
}

func (r *MenuRepository) Create(menu *models.Menu) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.QueryRow(
		`INSERT INTO menus (id, restaurant_id, name, description, created_at, updated_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at`,
		menu.RestaurantID, menu.Name, menu.Description,
	).Scan(&menu.ID, &menu.CreatedAt, &menu.UpdatedAt)
	if err != nil {
		return err
	}

	for i := range menu.Items {
		item := &menu.Items[i]
		err = tx.QueryRow(
			`INSERT INTO menu_items (id, menu_id, name, description, price, available)
			VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5)
			RETURNING id`,
			menu.ID, item.Name, item.Description, item.Price, item.Available,
		).Scan(&item.ID)
		if err != nil {
			return err
		}
		item.MenuID = menu.ID
	}

	return tx.Commit()
}

func (r *MenuRepository) FindByID(id string) (*models.Menu, error) {
	menu := &models.Menu{}
	err := r.db.QueryRow(
		`SELECT id, restaurant_id, name, description, created_at, updated_at FROM menus WHERE id = $1`, id,
	).Scan(&menu.ID, &menu.RestaurantID, &menu.Name, &menu.Description, &menu.CreatedAt, &menu.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(`SELECT id, menu_id, name, description, price, available FROM menu_items WHERE menu_id = $1`, menu.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.MenuItem
		if err := rows.Scan(&item.ID, &item.MenuID, &item.Name, &item.Description, &item.Price, &item.Available); err != nil {
			return nil, err
		}
		menu.Items = append(menu.Items, item)
	}

	return menu, nil
}

func (r *MenuRepository) Update(id string, req *models.UpdateMenuRequest) (*models.Menu, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	menu := &models.Menu{}
	err = tx.QueryRow(
		`UPDATE menus SET name = COALESCE(NULLIF($1,''), name), description = COALESCE(NULLIF($2,''), description),
		updated_at = NOW() WHERE id = $3
		RETURNING id, restaurant_id, name, description, created_at, updated_at`,
		req.Name, req.Description, id,
	).Scan(&menu.ID, &menu.RestaurantID, &menu.Name, &menu.Description, &menu.CreatedAt, &menu.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if len(req.Items) > 0 {
		if _, err := tx.Exec(`DELETE FROM menu_items WHERE menu_id = $1`, id); err != nil {
			return nil, err
		}
		for _, itemReq := range req.Items {
			var item models.MenuItem
			err = tx.QueryRow(
				`INSERT INTO menu_items (id, menu_id, name, description, price, available)
				VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5) RETURNING id`,
				id, itemReq.Name, itemReq.Description, itemReq.Price, itemReq.Available,
			).Scan(&item.ID)
			if err != nil {
				return nil, err
			}
			item.MenuID = id
			item.Name = itemReq.Name
			item.Description = itemReq.Description
			item.Price = itemReq.Price
			item.Available = itemReq.Available
			menu.Items = append(menu.Items, item)
		}
	}

	return menu, tx.Commit()
}

func (r *MenuRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM menus WHERE id = $1`, id)
	return err
}

func (r *MenuRepository) FindItemByID(id string) (*models.MenuItem, error) {
	item := &models.MenuItem{}
	err := r.db.QueryRow(
		`SELECT id, menu_id, name, description, price, available FROM menu_items WHERE id = $1`, id,
	).Scan(&item.ID, &item.MenuID, &item.Name, &item.Description, &item.Price, &item.Available)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return item, err
}
