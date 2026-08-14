package entity

import "errors"

var (
	ErrProfileIDRequired = errors.New("profile id is required")

	ErrProfileNotFound = errors.New("profile not found")

	ErrRecapIDRequired = errors.New("recap id is required")

	ErrRecapNotFound = errors.New("recap not found")

	ErrRecommendationNotFound = errors.New("recommendation not found")

	ErrFavoriteCategoryNotFound = errors.New("favorite category not found")

	ErrUnsupportedRecapAction = errors.New("unsupported recap action")

	ErrNotEnoughActivity = errors.New("not enough activity to build a recap")

	ErrSharedRecapTokenInvalid = errors.New("shared recap token is invalid")

	ErrSharedRecapNotFound = errors.New("shared recap not found")
)
