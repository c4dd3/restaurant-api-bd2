package repository

import (
	"database/sql"
	"errors"
	"restaurant-api/internal/models"
)

// MenuRepository handles all database operations for menus and their items.
type MenuRepository struct {
	db *sql.DB
}

// NewMenuRepository creates a MenuRepository with the given database connection.
func NewMenuRepository(db *sql.DB) *MenuRepository {
	return &MenuRepository{db: db}
}

// Create inserts a new menu and all its items in a single transaction.
// Using a transaction ensures that if any item insertion fails, the whole
// operation is rolled back and no partial data is left in the database.
func (r *MenuRepository) Create(menu *models.Menu) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	// Rollback is a no-op if Commit has already been called successfully.
	defer tx.Rollback()

	// Insert the menu row and read back the generated ID and timestamps.
	err = tx.QueryRow(
		`INSERT INTO menus (id, restaurant_id, name, description, created_at, updated_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at`,
		menu.RestaurantID, menu.Name, menu.Description,
	).Scan(&menu.ID, &menu.CreatedAt, &menu.UpdatedAt)
	if err != nil {
		return err
	}

	// Insert each menu item, linking it to the newly created menu ID.
	// We iterate by index so we can write the generated ID back into the slice element.
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

// FindByID returns the menu with the given ID, including all its items.
// Returns (nil, nil) if no menu with that ID exists.
func (r *MenuRepository) FindByID(id string) (*models.Menu, error) {
	menu := &models.Menu{}

	// Fetch the menu row first.
	err := r.db.QueryRow(
		`SELECT id, restaurant_id, name, description, created_at, updated_at FROM menus WHERE id = $1`, id,
	).Scan(&menu.ID, &menu.RestaurantID, &menu.Name, &menu.Description, &menu.CreatedAt, &menu.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // Menu not found — return nil without an error.
	}
	if err != nil {
		return nil, err
	}

	// Fetch all items that belong to this menu.
	rows, err := r.db.Query(`SELECT id, menu_id, name, description, price, available FROM menu_items WHERE menu_id = $1`, menu.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Scan each item row and append it to the menu's Items slice.
	for rows.Next() {
		var item models.MenuItem
		if err := rows.Scan(&item.ID, &item.MenuID, &item.Name, &item.Description, &item.Price, &item.Available); err != nil {
			return nil, err
		}
		menu.Items = append(menu.Items, item)
	}

	return menu, nil
}

// Update modifies the menu's name and description, and replaces its items if new ones are provided.
// COALESCE(NULLIF($1,''), name) means: keep the existing value if the request sends an empty string.
// Item replacement is done by deleting all existing items and re-inserting the new ones,
// all within a single transaction to keep the data consistent.
// Returns (nil, nil) if no menu with that ID exists.
func (r *MenuRepository) Update(id string, req *models.UpdateMenuRequest) (*models.Menu, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback is a no-op if Commit has already been called successfully.
	defer tx.Rollback()

	menu := &models.Menu{}

	// Update the menu row. COALESCE(NULLIF(value,''), column) preserves the existing
	// value when an empty string is sent, allowing partial updates.
	err = tx.QueryRow(
		`UPDATE menus SET name = COALESCE(NULLIF($1,''), name), description = COALESCE(NULLIF($2,''), description),
		updated_at = NOW() WHERE id = $3
		RETURNING id, restaurant_id, name, description, created_at, updated_at`,
		req.Name, req.Description, id,
	).Scan(&menu.ID, &menu.RestaurantID, &menu.Name, &menu.Description, &menu.CreatedAt, &menu.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // Menu not found — return nil without an error.
	}
	if err != nil {
		return nil, err
	}

	// Only replace items if the request includes a new item list.
	if len(req.Items) > 0 {
		// Delete all existing items for this menu before inserting the new ones.
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
			// Populate the remaining fields so the returned menu is complete.
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

// Delete removes a menu by ID. Deleting the menu also removes its items
// automatically via the ON DELETE CASCADE constraint defined in the schema.
func (r *MenuRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM menus WHERE id = $1`, id)
	return err
}

// FindItemByID returns a single menu item by its ID.
// Returns (nil, nil) if no item with that ID exists.
func (r *MenuRepository) FindItemByID(id string) (*models.MenuItem, error) {
	item := &models.MenuItem{}
	err := r.db.QueryRow(
		`SELECT id, menu_id, name, description, price, available FROM menu_items WHERE id = $1`, id,
	).Scan(&item.ID, &item.MenuID, &item.Name, &item.Description, &item.Price, &item.Available)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // Item not found — return nil without an error.
	}
	return item, err
}
