package seed

import (
	"crypto/sha256"
	"fmt"
	"hash/fnv"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
)

const (
	minSeedYear = 2015
	maxSeedYear = 2026
)

var seedNamespace = uuid.MustParse("0c074a78-6e04-5f4f-9125-e82f14994f2a")

var listingDetails = []string{
	"Можно проверить при встрече",
	"Отвечу на вопросы в сообщениях",
	"Причина продажи — больше не пользуюсь",
	"Самовывоз или доставка по договорённости",
}

type Generator struct {
	year    int
	seed    uint64
	catalog Catalog
}

func NewGenerator(year int, seed uint64, catalog Catalog) (*Generator, error) {
	if year < minSeedYear || year > maxSeedYear {
		return nil, fmt.Errorf("seed year must be in range %d..%d", minSeedYear, maxSeedYear)
	}

	if seed == 0 {
		return nil, fmt.Errorf("seed must be non-zero")
	}

	return &Generator{year: year, seed: seed, catalog: catalog}, nil
}

func (g *Generator) Generate(scenarios []Scenario) (Dataset, error) {
	state := generationState{
		dataset: Dataset{Year: g.year, Seed: g.seed},
		users:   make(map[string]UserRow, len(scenarios)),
		catalog: make(map[CategoryCode]generatedCategory, len(g.catalog.Categories)),
	}

	g.generateCatalog(&state)
	g.generateUsers(&state, scenarios)

	for _, scenario := range scenarios {
		if err := g.generatePublishedListings(&state, scenario); err != nil {
			return Dataset{}, err
		}
	}

	for _, scenario := range scenarios {
		for index, funnel := range scenario.Funnels {
			if err := g.generateFunnel(&state, scenario.Profile.Code, index, funnel); err != nil {
				return Dataset{}, err
			}
		}
	}

	return state.dataset, nil
}

type generatedCategory struct {
	row  CategoryRow
	spec CategorySpec
}

type generationState struct {
	dataset Dataset
	users   map[string]UserRow
	catalog map[CategoryCode]generatedCategory
}

func (g *Generator) generateCatalog(state *generationState) {
	for _, category := range g.catalog.Categories {
		categoryID := stableID("category", string(category.Code))
		categoryRow := CategoryRow{ID: categoryID, Code: category.Code, Title: category.Title}

		state.dataset.Categories = append(state.dataset.Categories, categoryRow)
		state.catalog[category.Code] = generatedCategory{row: categoryRow, spec: category}

		for _, subcategory := range category.Subcategories {
			state.dataset.Subcategories = append(state.dataset.Subcategories, SubcategoryRow{
				ID:         stableID("subcategory", string(category.Code), subcategory.Code),
				Code:       subcategory.Code,
				CategoryID: categoryID,
				Title:      subcategory.Title,
			})
		}
	}
}

func (g *Generator) generateUsers(state *generationState, scenarios []Scenario) {
	for _, scenario := range scenarios {
		profile := scenario.Profile
		rng, _ := randomSources(g.seed, "profile", profile.Code)

		registeredAt := time.Date(
			g.year-profile.YearsOnPlatform,
			time.Month(rng.IntN(12)+1),
			rng.IntN(20)+1,
			10,
			0,
			0,
			0,
			time.UTC,
		)

		row := UserRow{
			ID:           stableID("profile", profile.Code),
			Code:         profile.Code,
			Name:         profile.Name,
			Surname:      profile.Surname,
			AvatarURL:    optionalString(profile.AvatarURL),
			Hint:         optionalString(profile.Hint),
			RegisteredAt: registeredAt,
		}

		state.dataset.Users = append(state.dataset.Users, row)
		state.users[profile.Code] = row
	}
}

func (g *Generator) generatePublishedListings(state *generationState, scenario Scenario) error {
	plan := scenario.PublishedListings
	if plan.Count.Max == 0 {
		return nil
	}

	rng, faker := randomSources(g.seed, "published", scenario.Profile.Code)
	count := sampleCount(rng, plan.Count)
	seller := state.users[scenario.Profile.Code]

	for index := range count {
		category := weightedCategory(rng, plan.Categories)
		createdAt := scheduledTime(g.year, plan.Months, index, count, rng)
		key := strings.Join([]string{
			"published",
			scenario.Profile.Code,
			strconv.Itoa(index),
		}, ":")

		listing, err := g.newListing(state, rng, faker, key, index, seller.ID, category, createdAt)
		if err != nil {
			return fmt.Errorf("generate published listing for %s: %w", scenario.Profile.Code, err)
		}

		state.dataset.Listings = append(state.dataset.Listings, listing)
	}

	return nil
}

func (g *Generator) generateFunnel(
	state *generationState,
	buyerCode string,
	funnelIndex int,
	plan FunnelPlan,
) error {
	rng, faker := randomSources(
		g.seed,
		"funnel",
		buyerCode,
		plan.SellerCode,
		string(plan.Category),
		strconv.Itoa(funnelIndex),
	)

	funnel := funnelGeneration{
		generator:   g,
		state:       state,
		rng:         rng,
		faker:       faker,
		buyerCode:   buyerCode,
		funnelIndex: funnelIndex,
		plan:        plan,
		counts:      sampleFunnelCounts(rng, plan),
		buyer:       state.users[buyerCode],
		seller:      state.users[plan.SellerCode],
	}

	if err := funnel.generateListings(); err != nil {
		return err
	}

	funnel.generateRepeatedViews()
	funnel.generateMessages()

	return nil
}

type funnelCounts struct {
	views           int
	favorites       int
	activeFavorites int
	messages        int
	purchases       int
	uniqueListings  int
}

type funnelGeneration struct {
	generator   *Generator
	state       *generationState
	rng         *rand.Rand
	faker       *gofakeit.Faker
	buyerCode   string
	funnelIndex int
	plan        FunnelPlan
	counts      funnelCounts
	buyer       UserRow
	seller      UserRow
	listings    []ListingRow
	baseTimes   []time.Time
}

func sampleFunnelCounts(rng *rand.Rand, plan FunnelPlan) funnelCounts {
	views := sampleCount(rng, plan.Views)
	favorites := sampleCount(rng, plan.Favorites)

	return funnelCounts{
		views:           views,
		favorites:       favorites,
		activeFavorites: sampleCount(rng, plan.ActiveFavorites),
		messages:        sampleCount(rng, plan.Messages),
		purchases:       sampleCount(rng, plan.Purchases),
		uniqueListings:  favorites + (views-favorites+1)/2,
	}
}

func (f *funnelGeneration) generateListings() error {
	f.listings = make([]ListingRow, 0, f.counts.uniqueListings)
	f.baseTimes = make([]time.Time, 0, f.counts.uniqueListings)

	for index := range f.counts.uniqueListings {
		listing, baseTime, err := f.generateListing(index)
		if err != nil {
			return err
		}

		f.listings = append(f.listings, listing)
		f.baseTimes = append(f.baseTimes, baseTime)
	}

	return nil
}

func (f *funnelGeneration) generateListing(index int) (ListingRow, time.Time, error) {
	roleIndex, roleCount := funnelRolePosition(
		index,
		f.counts.purchases,
		f.counts.activeFavorites,
		f.counts.favorites,
		f.counts.uniqueListings,
	)
	baseTime := scheduledTime(f.generator.year, f.plan.Months, roleIndex, roleCount, f.rng)
	createdAt := baseTime.AddDate(-1, 0, 0).Add(-time.Duration(f.rng.IntN(30)) * 24 * time.Hour)
	key := f.key("inventory", index)

	listing, err := f.generator.newListing(
		f.state,
		f.rng,
		f.faker,
		key,
		index,
		f.seller.ID,
		f.plan.Category,
		createdAt,
	)
	if err != nil {
		return ListingRow{}, time.Time{}, fmt.Errorf("generate funnel listing for %s: %w", f.buyerCode, err)
	}

	setListingRole(&listing, baseTime, index, f.counts)
	f.state.dataset.Listings = append(f.state.dataset.Listings, listing)
	f.state.dataset.Views = append(f.state.dataset.Views, ViewRow{
		ID:        f.generator.eventID("view", key, "initial"),
		UserID:    f.buyer.ID,
		ListingID: listing.ID,
		ViewedAt:  baseTime,
	})

	if index < f.counts.favorites {
		f.state.dataset.Favorites = append(f.state.dataset.Favorites, FavoriteRow{
			UserID:    f.buyer.ID,
			ListingID: listing.ID,
			CreatedAt: baseTime.Add(2 * time.Hour),
		})
	}
	if index < f.counts.purchases {
		createdAt := baseTime.Add(12 * time.Hour)
		completedAt := baseTime.Add(24 * time.Hour)
		f.state.dataset.Deals = append(f.state.dataset.Deals, DealRow{
			ID:          f.generator.eventID("deal", key),
			ListingID:   listing.ID,
			BuyerID:     f.buyer.ID,
			Price:       listing.Price,
			CreatedAt:   createdAt,
			CompletedAt: &completedAt,
		})
	}

	return listing, baseTime, nil
}

func setListingRole(
	listing *ListingRow,
	baseTime time.Time,
	index int,
	counts funnelCounts,
) {
	switch {
	case index < counts.purchases:
		completedAt := baseTime.Add(24 * time.Hour)
		listing.Status = ListingSold
		listing.ClosedAt = &completedAt
	case index < counts.purchases+counts.activeFavorites:
		listing.Status = ListingActive
	case index < counts.favorites:
		closedAt := baseTime.Add(48 * time.Hour)
		listing.Status = ListingClosed
		listing.ClosedAt = &closedAt
	}
}

func (f *funnelGeneration) generateRepeatedViews() {
	for index := range f.counts.views - f.counts.uniqueListings {
		listingIndex := index % f.counts.uniqueListings
		key := f.key("repeat", index)

		f.state.dataset.Views = append(f.state.dataset.Views, ViewRow{
			ID:        f.generator.eventID("view", key),
			UserID:    f.buyer.ID,
			ListingID: f.listings[listingIndex].ID,
			ViewedAt:  f.baseTimes[listingIndex].Add(time.Duration(index+1) * time.Minute),
		})
	}
}

func (f *funnelGeneration) generateMessages() {
	messageListings := f.counts.purchases
	if messageListings == 0 {
		messageListings = f.counts.favorites
	}
	if messageListings == 0 {
		messageListings = f.counts.uniqueListings
	}

	for index := range f.counts.messages {
		listingIndex := index % messageListings
		conversationRound := index / messageListings
		key := f.key("message", index)

		f.state.dataset.Messages = append(f.state.dataset.Messages, MessageRow{
			ID:        f.generator.eventID(key),
			BuyerID:   f.buyer.ID,
			SellerID:  f.seller.ID,
			ListingID: f.listings[listingIndex].ID,
			CreatedAt: f.baseTimes[listingIndex].Add(4*time.Hour + time.Duration(conversationRound)*5*time.Minute),
		})
	}
}

func (f *funnelGeneration) key(prefix string, index int) string {
	return strings.Join([]string{
		prefix,
		f.buyerCode,
		f.plan.SellerCode,
		string(f.plan.Category),
		strconv.Itoa(f.funnelIndex),
		strconv.Itoa(index),
	}, ":")
}

func (g *Generator) newListing(
	state *generationState,
	rng *rand.Rand,
	faker *gofakeit.Faker,
	key string,
	titleVariant int,
	sellerID uuid.UUID,
	categoryCode CategoryCode,
	createdAt time.Time,
) (ListingRow, error) {
	category, ok := state.catalog[categoryCode]
	if !ok {
		return ListingRow{}, fmt.Errorf("unknown category %q", categoryCode)
	}

	subcategory := category.spec.Subcategories[rng.IntN(len(category.spec.Subcategories))]
	product := subcategory.Products[rng.IntN(len(subcategory.Products))]
	if len(product.Titles) == 0 {
		return ListingRow{}, fmt.Errorf(
			"product in subcategory %q has no listing titles",
			subcategory.Code,
		)
	}

	subcategoryID := stableID("subcategory", string(categoryCode), subcategory.Code)
	description := product.Description + ". " + faker.RandomString(listingDetails)
	price := randomPrice(rng, product.MinPrice, product.MaxPrice)

	return ListingRow{
		ID:            g.eventID("listing", key),
		SellerID:      sellerID,
		Title:         product.Titles[titleVariant%len(product.Titles)],
		Description:   &description,
		Price:         price,
		CategoryID:    category.row.ID,
		SubcategoryID: &subcategoryID,
		Status:        ListingActive,
		CreatedAt:     createdAt,
	}, nil
}

func (g *Generator) eventID(parts ...string) uuid.UUID {
	prefix := []string{strconv.FormatUint(g.seed, 10), strconv.Itoa(g.year)}

	return stableID(append(prefix, parts...)...)
}

func stableID(parts ...string) uuid.UUID {
	return uuid.NewHash(
		sha256.New(),
		seedNamespace,
		[]byte(strings.Join(parts, ":")),
		4,
	)
}

func randomSources(seed uint64, parts ...string) (*rand.Rand, *gofakeit.Faker) {
	derived := deriveSeed(seed, parts...)
	//nolint:gosec // Reproducible demo data requires a deterministic PRNG, not cryptographic randomness.
	behavior := rand.New(rand.NewPCG(derived, derived^0x9e3779b97f4a7c15))
	contentSeed := derived ^ 0xd1b54a32d192ed03
	if contentSeed == 0 {
		contentSeed = 1
	}

	return behavior, gofakeit.New(contentSeed)
}

func deriveSeed(seed uint64, parts ...string) uint64 {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(strconv.FormatUint(seed, 10)))

	for _, part := range parts {
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write([]byte(part))
	}

	result := hash.Sum64()
	if result == 0 {
		return 1
	}

	return result
}

func sampleCount(rng *rand.Rand, count CountRange) int {
	if count.Min == count.Max {
		return count.Min
	}

	return count.Min + rng.IntN(count.Max-count.Min+1)
}

func weightedCategory(rng *rand.Rand, categories []CategoryWeight) CategoryCode {
	var total int
	for _, category := range categories {
		total += category.Weight
	}

	value := rng.IntN(total)
	for _, category := range categories {
		if value < category.Weight {
			return category.Category
		}
		value -= category.Weight
	}

	return categories[len(categories)-1].Category
}

func scheduledTime(
	year int,
	months []time.Month,
	index int,
	total int,
	rng *rand.Rand,
) time.Time {
	month := months[index*len(months)/total]

	return time.Date(year, month, rng.IntN(20)+1, rng.IntN(10)+8, rng.IntN(60), 0, 0, time.UTC)
}

func funnelRolePosition(
	index int,
	purchases int,
	activeFavorites int,
	favorites int,
	uniqueListings int,
) (int, int) {
	switch {
	case index < purchases:
		return index, purchases
	case index < purchases+activeFavorites:
		return index - purchases, activeFavorites
	case index < favorites:
		return index - purchases - activeFavorites, favorites - purchases - activeFavorites
	default:
		return index - favorites, uniqueListings - favorites
	}
}

func randomPrice(rng *rand.Rand, minPrice, maxPrice int64) int64 {
	const priceStep = int64(10_000)

	steps := (maxPrice - minPrice) / priceStep
	if steps == 0 {
		return minPrice
	}

	return minPrice + rng.Int64N(steps+1)*priceStep
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}
