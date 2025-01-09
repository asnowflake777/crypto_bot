package pgdb

import (
	"context"
	sq "github.com/Masterminds/squirrel"
)

type User struct {
	UID   string
	Login string
}

type Balance struct {
	Asset  string
	Free   float64
	Locked float64
}

type CreateUserRequest struct {
	Login string
}

func (c *Client) CreateUser(ctx context.Context, r CreateUserRequest) (*User, error) {
	queryStr, args, err := sq.
		Insert("users").
		Columns("login").
		Values(r.Login).
		Suffix("RETURNING uid").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}
	user := &User{Login: r.Login}
	if err = c.conn.QueryRow(ctx, queryStr, args...).Scan(&user.UID); err != nil {
		return nil, err
	}
	return user, nil
}

type ReadUserRequest struct {
	UID   string
	Login string
}

func (c *Client) ReadUser(ctx context.Context, r ReadUserRequest) (*User, error) {
	query := sq.
		Select("uid", "login").
		From("users").
		PlaceholderFormat(sq.Dollar)
	if r.UID != "" {
		query = query.Where(sq.Eq{"uid": r.UID})
	}
	if r.Login != "" {
		query = query.Where(sq.Eq{"login": r.Login})
	}
	queryStr, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	var user User
	if err = c.conn.QueryRow(ctx, queryStr, args...).Scan(&user.UID, &user.Login); err != nil {
		return nil, err
	}
	return &user, nil
}

type UpdateUserRequest struct {
	UID   string
	Login string
}

func (c *Client) UpdateUser(ctx context.Context, r UpdateUserRequest) (*User, error) {
	queryStr, args, err := sq.
		Update("users").
		Set("login = ?", r.Login).
		Where(sq.Eq{"uid": r.UID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}
	if _, err = c.conn.Exec(ctx, queryStr, args...); err != nil {
		return nil, err
	}
	return &User{Login: r.Login, UID: r.UID}, nil
}

type DeleteUserRequest struct {
	UID string
}

func (c *Client) DeleteUser(ctx context.Context, r DeleteUserRequest) error {
	queryStr, args, err := sq.
		Delete("users").
		Where(sq.Eq{"uid": r.UID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}
	if _, err = c.conn.Exec(ctx, queryStr, args...); err != nil {
		return err
	}
	return nil
}

type CreateBalanceRequest struct {
	UserUID string
	Asset   string
	Free    float64
	Locked  float64
}

func (c *Client) CreateBalance(ctx context.Context, r CreateBalanceRequest) (*Balance, error) {
	queryStr, args, err := sq.
		Insert("balance").
		Columns("user_uid", "asset", "free", "locked").
		Values(r.UserUID, r.Asset, r.Free, r.Locked).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}
	if _, err = c.conn.Exec(ctx, queryStr, args...); err != nil {
		return nil, err
	}
	return &Balance{Asset: r.Asset, Free: r.Free, Locked: r.Locked}, nil
}

type ReadBalanceRequest struct {
	UserUID string
	Asset   string
}

func (c *Client) ReadBalance(ctx context.Context, r ReadBalanceRequest) (*Balance, error) {
	queryStr, args, err := sq.
		Select("asset", "free", "locked").
		From("balance").
		Where(sq.Eq{"user_uid": r.UserUID, "asset": r.Asset}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}
	var balance Balance
	if err = c.conn.QueryRow(ctx, queryStr, args...).Scan(&balance.Asset, &balance.Free, &balance.Locked); err != nil {
		return nil, err
	}
	return &balance, nil
}

type ReadBalancesRequest struct {
	UserUID string
}

func (c *Client) ReadBalances(ctx context.Context, r ReadBalancesRequest) ([]*Balance, error) {
	queryStr, args, err := sq.
		Select("asset", "free", "locked").
		From("balance").
		Where(sq.Eq{"user_uid": r.UserUID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := c.conn.Query(ctx, queryStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var balances []*Balance
	for rows.Next() {
		var balance Balance
		if err = rows.Scan(&balance.Asset, &balance.Free, &balance.Locked); err != nil {
			return nil, err
		}
		balances = append(balances, &balance)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return balances, nil
}

type UpdateBalanceRequest struct {
	UserUID string
	Asset   string
	Free    float64
	Locked  float64
}

func (c *Client) UpdateBalance(ctx context.Context, r UpdateBalanceRequest) (*Balance, error) {
	queryStr, args, err := sq.
		Update("balance").
		Set("free", r.Free).
		Set("locked", r.Locked).
		Where(sq.Eq{"user_uid": r.UserUID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}
	if _, err = c.conn.Exec(ctx, queryStr, args...); err != nil {
		return nil, err
	}
	return &Balance{Free: r.Free, Locked: r.Locked}, nil
}

type DeleteBalanceRequest struct {
	UserUID string
	Asset   string
}

func (c *Client) DeleteBalance(ctx context.Context, r DeleteBalanceRequest) error {
	queryStr, args, err := sq.
		Delete("balance").
		Where(sq.Eq{"user_id": r.UserUID, "asset": r.Asset}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}
	if _, err = c.conn.Exec(ctx, queryStr, args...); err != nil {
		return err
	}
	return nil
}
