package repository

import (
	"context"
	"errors"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"goshop/internal/domain/entities"
	"goshop/internal/domain_errors"
	"time"
)

type AddressRepository struct {
	db   *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewAddressRepository(db *pgxpool.Pool) *AddressRepository {
	return &AddressRepository{
		db:   db,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *AddressRepository) CreateAddress(ctx context.Context, address *entities.UserAddress) error {
	query := r.psql.Insert("user_addresses").
		Columns("uuid", "user_id", "address", "city", "postal_code", "country", "created_at").
		Values(address.UUID, address.UserID, address.Address, address.City, address.PostalCode, address.Country, address.CreatedAt)

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, sql, args...)
	return err
}

func (r *AddressRepository) GetUserAddresses(ctx context.Context, userID int64) ([]*entities.UserAddress, error) {
	query := r.psql.Select("id", "uuid", "user_id", "address", "city", "postal_code", "country", "created_at").
		From("user_addresses").
		Where(squirrel.Eq{"user_id": userID})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addresses []*entities.UserAddress
	for rows.Next() {
		var address entities.UserAddress
		if err := rows.Scan(
			&address.ID,
			&address.UUID,
			&address.UserID,
			&address.Address,
			&address.City,
			&address.PostalCode,
			&address.Country,
			&address.CreatedAt,
		); err != nil {
			return nil, err
		}
		addresses = append(addresses, &address)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return addresses, nil
}

func (r *AddressRepository) GetAddressByID(ctx context.Context, addressID int64) (*entities.UserAddress, error) {
	query := r.psql.Select("id", "uuid", "user_id", "address", "city", "postal_code", "country", "created_at").
		From("user_addresses").
		Where(squirrel.Eq{"id": addressID})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	row := r.db.QueryRow(ctx, sql, args...)
	var address entities.UserAddress
	if err := row.Scan(
		&address.ID,
		&address.UUID,
		&address.UserID,
		&address.Address,
		&address.City,
		&address.PostalCode,
		&address.Country,
		&address.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain_errors.ErrAddressNotFound
		}
		return nil, err
	}
	return &address, nil
}

func (r *AddressRepository) UpdateAddress(ctx context.Context, address *entities.UserAddress) error {
	query := r.psql.Update("user_addresses")
	paramsCnt := 0

	if address.Address != "" {
		query = query.Set("address", address.Address)
		paramsCnt++
	}
	if address.City != nil {
		query = query.Set("city", *address.City)
		paramsCnt++
	}
	if address.PostalCode != nil {
		query = query.Set("postal_code", *address.PostalCode)
		paramsCnt++
	}
	if address.Country != nil {
		query = query.Set("country", *address.Country)
		paramsCnt++
	}
	if paramsCnt == 0 {
		return domain_errors.ErrInvalidInput
	}

	query = query.Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": address.ID, "user_id": address.UserID})

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	result, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain_errors.ErrAddressNotFound
	}
	return nil
}

func (r *AddressRepository) DeleteAddress(ctx context.Context, addressID int64) error {
	query := r.psql.Delete("user_addresses").Where(squirrel.Eq{"id": addressID})

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	result, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain_errors.ErrAddressNotFound
	}
	return nil
}
