package repository

import (
	"context"
	"goshop/internal/domain/entities"
	"goshop/internal/domain/types"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type ReviewRepository struct {
	db     *pgxpool.Pool
	psql   squirrel.StatementBuilderType
	logger *zap.Logger
}

func NewReviewRepository(db *pgxpool.Pool, logger *zap.Logger) *ReviewRepository {
	return &ReviewRepository{
		db:     db,
		psql:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger: logger,
	}
}

func (r *ReviewRepository) CreateReview(ctx context.Context, review *entities.Review) (*int64, error) {
	query := `INSERT INTO reviews (uuid, product_id, user_id, rating, comment, created_at)
              VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	args := []interface{}{
		review.UUID,
		review.ProductID,
		review.UserID,
		review.Rating,
		review.Comment,
		review.CreatedAt,
	}

	var id int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&id)
	if err != nil {
		r.logger.Error("Failed to create review", zap.Error(err))
		return nil, err
	}

	return &id, nil
}

func (r *ReviewRepository) GetReviewsWithFilters(ctx context.Context, filters types.ReviewFilters) ([]*entities.Review, int64, error) {
	query := r.psql.Select(
		"id", "uuid", "product_id", "user_id", "rating", "comment", "created_at",
	).From("reviews")

	if filters.ProductID != nil {
		query = query.Where(squirrel.Eq{"product_id": *filters.ProductID})
	}

	if filters.UserID != nil {
		query = query.Where(squirrel.Eq{"user_id": *filters.UserID})
	}

	if filters.Rating != nil {
		query = query.Where(squirrel.Eq{"rating": *filters.Rating})
	}

	sortBy := "created_at"
	if filters.SortBy != nil {
		switch *filters.SortBy {
		case "created_at", "rating":
			sortBy = *filters.SortBy
		}
	}

	sortOrder := "DESC"
	if filters.SortOrder != nil && *filters.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	query = query.OrderBy(sortBy + " " + sortOrder)

	offset := (filters.Page - 1) * filters.Limit
	query = query.Limit(uint64(filters.Limit)).Offset(uint64(offset))

	sql, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("Failed to build review query", zap.Error(err))
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		r.logger.Error("Failed to execute review query", zap.Error(err), zap.String("sql", sql))
		return nil, 0, err
	}
	defer rows.Close()

	reviews := make([]*entities.Review, 0, filters.Limit)
	for rows.Next() {
		review := &entities.Review{}
		err := rows.Scan(
			&review.ID,
			&review.UUID,
			&review.ProductID,
			&review.UserID,
			&review.Rating,
			&review.Comment,
			&review.CreatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan review", zap.Error(err))
			return nil, 0, err
		}
		reviews = append(reviews, review)
	}

	if rows.Err() != nil {
		r.logger.Error("Row iteration error", zap.Error(rows.Err()))
		return nil, 0, rows.Err()
	}

	countQuery := r.psql.Select("COUNT(*)").From("reviews")

	if filters.ProductID != nil {
		countQuery = countQuery.Where(squirrel.Eq{"product_id": *filters.ProductID})
	}
	if filters.UserID != nil {
		countQuery = countQuery.Where(squirrel.Eq{"user_id": *filters.UserID})
	}
	if filters.Rating != nil {
		countQuery = countQuery.Where(squirrel.Eq{"rating": *filters.Rating})
	}

	countSql, countArgs, err := countQuery.ToSql()
	if err != nil {
		r.logger.Error("Failed to build count query", zap.Error(err))
		return nil, 0, err
	}

	var totalCount int64
	err = r.db.QueryRow(ctx, countSql, countArgs...).Scan(&totalCount)
	if err != nil {
		r.logger.Error("Failed to get reviews count", zap.Error(err))
		return nil, 0, err
	}

	return reviews, totalCount, nil
}

func (r *ReviewRepository) GetReviewByID(ctx context.Context, reviewID int64) (*entities.Review, error) {
	sql, args, err := r.psql.Select(
		"id", "uuid", "product_id", "user_id", "rating", "comment", "created_at",
	).From("reviews").
		Where(squirrel.Eq{"id": reviewID}).
		ToSql()

	if err != nil {
		r.logger.Error("Failed to build get review query", zap.Error(err))
		return nil, err
	}

	review := &entities.Review{}
	err = r.db.QueryRow(ctx, sql, args...).Scan(
		&review.ID,
		&review.UUID,
		&review.ProductID,
		&review.UserID,
		&review.Rating,
		&review.Comment,
		&review.CreatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to get review by ID", zap.Error(err), zap.Int64("review_id", reviewID))
		return nil, err
	}

	return review, nil
}

func (r *ReviewRepository) UpdateReview(ctx context.Context, reviewID int64, rating *int, comment *string) error {
	updateQuery := r.psql.Update("reviews")

	if rating != nil {
		updateQuery = updateQuery.Set("rating", rating)
	}

	if comment != nil {
		updateQuery = updateQuery.Set("comment", comment)
	}

	updateQuery = updateQuery.Where(squirrel.Eq{"id": reviewID})

	sql, args, err := updateQuery.ToSql()
	if err != nil {
		r.logger.Error("Failed to build update review query", zap.Error(err))
		return err
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		r.logger.Error("Failed to update review", zap.Error(err), zap.Int64("review_id", reviewID))
		return err
	}

	return nil
}

func (r *ReviewRepository) DeleteReview(ctx context.Context, reviewID int64) error {
	sql, args, err := r.psql.Delete("reviews").
		Where(squirrel.Eq{"id": reviewID}).
		ToSql()

	if err != nil {
		r.logger.Error("Failed to build delete review query", zap.Error(err))
		return err
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		r.logger.Error("Failed to delete review", zap.Error(err), zap.Int64("review_id", reviewID))
		return err
	}

	return nil
}

func (r *ReviewRepository) CheckUserReviewExists(ctx context.Context, userID, productID int64) (bool, error) {

	query := "SELECT COUNT(*) FROM reviews WHERE user_id = $1 AND product_id = $2"

	var count int
	err := r.db.QueryRow(ctx, query, userID, productID).Scan(&count)
	if err != nil {
		r.logger.Error("Failed to check if user review exists",
			zap.Error(err),
			zap.Int64("user_id", userID),
			zap.Int64("product_id", productID))
		return false, err
	}

	return count > 0, nil
}

func (r *ReviewRepository) GetReviewStats(ctx context.Context, productID int64) (totalReviews int64, averageRating float64, ratingCounts map[int]int64, err error) {
	sql, args, err := r.psql.Select(
		"COUNT(*) as total",
		"COALESCE(AVG(rating), 0) as avg_rating",
		"COUNT(CASE WHEN rating = 1 THEN 1 END) as rating_1",
		"COUNT(CASE WHEN rating = 2 THEN 1 END) as rating_2",
		"COUNT(CASE WHEN rating = 3 THEN 1 END) as rating_3",
		"COUNT(CASE WHEN rating = 4 THEN 1 END) as rating_4",
		"COUNT(CASE WHEN rating = 5 THEN 1 END) as rating_5",
	).From("reviews").
		Where(squirrel.Eq{"product_id": productID}).
		ToSql()

	if err != nil {
		r.logger.Error("Failed to build review stats query", zap.Error(err))
		return 0, 0, nil, err
	}

	var rating1, rating2, rating3, rating4, rating5 int64
	err = r.db.QueryRow(ctx, sql, args...).Scan(
		&totalReviews,
		&averageRating,
		&rating1, &rating2, &rating3, &rating4, &rating5,
	)

	if err != nil {
		r.logger.Error("Failed to get review stats", zap.Error(err), zap.Int64("product_id", productID))
		return 0, 0, nil, err
	}

	ratingCounts = map[int]int64{
		1: rating1,
		2: rating2,
		3: rating3,
		4: rating4,
		5: rating5,
	}

	return totalReviews, averageRating, ratingCounts, nil
}
