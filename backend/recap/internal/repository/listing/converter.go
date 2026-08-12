package listing

import "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"

func listingPreviewModelToEntity(model listingPreviewModel) entity.ListingPreview {
	return entity.ListingPreview{
		ID:         model.ID,
		Title:      model.Title,
		Price:      model.Price,
		CategoryID: model.CategoryID,
	}
}

func favoriteListingPreviewModelToEntity(
	model favoriteListingPreviewModel,
) entity.FavoriteListingPreview {
	return entity.FavoriteListingPreview{
		ListingPreview: listingPreviewModelToEntity(model.listingPreviewModel),
		AddedAt:        model.AddedAt,
	}
}
