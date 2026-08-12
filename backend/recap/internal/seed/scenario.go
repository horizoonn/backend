package seed

import (
	"time"
)

type CategoryCode string

const (
	CategoryElectronics CategoryCode = "electronics"
	CategoryHome        CategoryCode = "home"
	CategoryTransport   CategoryCode = "transport"
	CategoryHobby       CategoryCode = "hobby"
	CategoryFashion     CategoryCode = "fashion"
	CategoryRealEstate  CategoryCode = "real_estate"
)

type CountRange struct {
	Min int
	Max int
}

type ProfileSpec struct {
	Code            string
	Name            string
	Surname         string
	AvatarURL       string
	Hint            string
	YearsOnPlatform int
}

type CategoryWeight struct {
	Category CategoryCode
	Weight   int
}

type ListingPlan struct {
	Count      CountRange
	Categories []CategoryWeight
	Months     []time.Month
}

type FunnelPlan struct {
	SellerCode      string
	Category        CategoryCode
	Months          []time.Month
	Views           CountRange
	Favorites       CountRange
	ActiveFavorites CountRange
	Messages        CountRange
	Purchases       CountRange
}

type Scenario struct {
	Profile           ProfileSpec
	PublishedListings ListingPlan
	Funnels           []FunnelPlan
}

func DefaultScenarios() []Scenario {
	allYear := []time.Month{
		time.January,
		time.February,
		time.March,
		time.April,
		time.May,
		time.June,
		time.July,
		time.August,
		time.September,
		time.October,
		time.November,
		time.December,
	}

	return []Scenario{
		collectorScenario(),
		dealmakerScenario(allYear),
		negotiatorScenario(allYear),
		explorerScenario(),
		inactiveScenario(),
	}
}

func collectorScenario() Scenario {
	seasonalMonths := []time.Month{
		time.January,
		time.April,
		time.July,
		time.October,
	}

	return Scenario{
		Profile: ProfileSpec{
			Code:            "collector",
			Name:            "Анна",
			Surname:         "Воронова",
			Hint:            "Сохраняет находки и возвращается к избранному",
			YearsOnPlatform: 6,
		},
		Funnels: []FunnelPlan{
			{
				SellerCode:      "dealmaker",
				Category:        CategoryHome,
				Months:          seasonalMonths,
				Views:           CountRange{Min: 135, Max: 145},
				Favorites:       CountRange{Min: 115, Max: 125},
				ActiveFavorites: CountRange{Min: 70, Max: 80},
				Messages:        CountRange{Min: 4, Max: 6},
				Purchases:       CountRange{Min: 8, Max: 8},
			},
			{
				SellerCode:      "dealmaker",
				Category:        CategoryElectronics,
				Months:          seasonalMonths,
				Views:           CountRange{Min: 70, Max: 80},
				Favorites:       CountRange{Min: 60, Max: 68},
				ActiveFavorites: CountRange{Min: 38, Max: 45},
				Messages:        CountRange{Min: 3, Max: 5},
				Purchases:       CountRange{Min: 3, Max: 3},
			},
		},
	}
}

func dealmakerScenario(allYear []time.Month) Scenario {
	return Scenario{
		Profile: ProfileSpec{
			Code:            "dealmaker",
			Name:            "Михаил",
			Surname:         "Орлов",
			Hint:            "Регулярно публикует объявления и закрывает сделки",
			YearsOnPlatform: 8,
		},
		PublishedListings: ListingPlan{
			Count: CountRange{Min: 180, Max: 240},
			Categories: []CategoryWeight{
				{Category: CategoryHome, Weight: 60},
				{Category: CategoryElectronics, Weight: 40},
			},
			Months: allYear,
		},
	}
}

func negotiatorScenario(allYear []time.Month) Scenario {
	return Scenario{
		Profile: ProfileSpec{
			Code:            "negotiator",
			Name:            "Елена",
			Surname:         "Соколова",
			Hint:            "Уточняет детали, сравнивает предложения и торгуется",
			YearsOnPlatform: 5,
		},
		Funnels: []FunnelPlan{
			{
				SellerCode:      "explorer",
				Category:        CategoryTransport,
				Months:          allYear,
				Views:           CountRange{Min: 50, Max: 65},
				Favorites:       CountRange{Min: 12, Max: 15},
				ActiveFavorites: CountRange{Min: 4, Max: 5},
				Messages:        CountRange{Min: 100, Max: 125},
				Purchases:       CountRange{Min: 5, Max: 6},
			},
			{
				SellerCode:      "explorer",
				Category:        CategoryRealEstate,
				Months:          allYear,
				Views:           CountRange{Min: 25, Max: 35},
				Favorites:       CountRange{Min: 5, Max: 7},
				ActiveFavorites: CountRange{Min: 2, Max: 3},
				Messages:        CountRange{Min: 60, Max: 75},
			},
			{
				SellerCode:      "dealmaker",
				Category:        CategoryElectronics,
				Months:          allYear,
				Views:           CountRange{Min: 20, Max: 25},
				Favorites:       CountRange{Min: 6, Max: 6},
				ActiveFavorites: CountRange{Min: 2, Max: 2},
				Messages:        CountRange{Min: 20, Max: 25},
				Purchases:       CountRange{Min: 3, Max: 4},
			},
		},
	}
}

func explorerScenario() Scenario {
	return Scenario{
		Profile: ProfileSpec{
			Code:            "explorer",
			Name:            "Илья",
			Surname:         "Лебедев",
			Hint:            "Исследует разные категории и меняет интересы по сезонам",
			YearsOnPlatform: 4,
		},
		Funnels: []FunnelPlan{
			{
				SellerCode:      "dealmaker",
				Category:        CategoryElectronics,
				Months:          []time.Month{time.January, time.February, time.December},
				Views:           CountRange{Min: 450, Max: 490},
				Favorites:       CountRange{Min: 25, Max: 30},
				ActiveFavorites: CountRange{Min: 5, Max: 5},
				Purchases:       CountRange{Min: 3, Max: 3},
			},
			{
				SellerCode:      "inactive",
				Category:        CategoryHobby,
				Months:          []time.Month{time.March, time.April, time.May},
				Views:           CountRange{Min: 530, Max: 590},
				Favorites:       CountRange{Min: 40, Max: 50},
				ActiveFavorites: CountRange{Min: 5, Max: 5},
				Purchases:       CountRange{Min: 1, Max: 1},
			},
			{
				SellerCode:      "negotiator",
				Category:        CategoryTransport,
				Months:          []time.Month{time.June, time.July, time.August},
				Views:           CountRange{Min: 450, Max: 500},
				Favorites:       CountRange{Min: 25, Max: 35},
				ActiveFavorites: CountRange{Min: 5, Max: 5},
				Purchases:       CountRange{Min: 1, Max: 1},
			},
			{
				SellerCode:      "dealmaker",
				Category:        CategoryHome,
				Months:          []time.Month{time.September, time.October, time.November},
				Views:           CountRange{Min: 450, Max: 500},
				Favorites:       CountRange{Min: 25, Max: 35},
				ActiveFavorites: CountRange{Min: 5, Max: 5},
				Purchases:       CountRange{Min: 2, Max: 2},
			},
		},
	}
}

func inactiveScenario() Scenario {
	return Scenario{
		Profile: ProfileSpec{
			Code:            "inactive",
			Name:            "Мария",
			Surname:         "Тихонова",
			Hint:            "Редко заходила на площадку в этом году",
			YearsOnPlatform: 2,
		},
		Funnels: []FunnelPlan{
			{
				SellerCode: "dealmaker",
				Category:   CategoryFashion,
				Months:     []time.Month{time.April},
				Views:      CountRange{Min: 3, Max: 3},
			},
		},
	}
}
