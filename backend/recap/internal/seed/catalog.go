package seed

type Catalog struct {
	Categories []CategorySpec
}

type CategorySpec struct {
	Code          CategoryCode
	Title         string
	Subcategories []SubcategorySpec
}

type SubcategorySpec struct {
	Code     string
	Title    string
	Products []ProductTemplate
}

type ProductTemplate struct {
	Title       string
	Description string
	MinPrice    int64
	MaxPrice    int64
}

func DefaultCatalog() Catalog {
	return Catalog{Categories: []CategorySpec{
		{
			Code:  CategoryElectronics,
			Title: "Электроника",
			Subcategories: []SubcategorySpec{
				{
					Code:  "smartphones",
					Title: "Мобильные телефоны",
					Products: []ProductTemplate{
						{Title: "Смартфон", Description: "Полностью исправен, бережное использование", MinPrice: 15_000_00, MaxPrice: 95_000_00},
						{Title: "Телефон", Description: "Комплект с зарядным устройством и коробкой", MinPrice: 8_000_00, MaxPrice: 65_000_00},
					},
				},
				{
					Code:  "computers",
					Title: "Ноутбуки",
					Products: []ProductTemplate{
						{Title: "Ноутбук", Description: "Подходит для работы, учёбы и повседневных задач", MinPrice: 25_000_00, MaxPrice: 160_000_00},
						{Title: "Ультрабук", Description: "Лёгкий корпус, аккумулятор хорошо держит заряд", MinPrice: 35_000_00, MaxPrice: 180_000_00},
					},
				},
			},
		},
		{
			Code:  CategoryHome,
			Title: "Для дома и дачи",
			Subcategories: []SubcategorySpec{
				{
					Code:  "furniture",
					Title: "Мебель и интерьер",
					Products: []ProductTemplate{
						{Title: "Стеллаж", Description: "В хорошем состоянии, легко собирается", MinPrice: 2_000_00, MaxPrice: 18_000_00},
						{Title: "Рабочий стол", Description: "Устойчивый стол для дома или офиса", MinPrice: 3_500_00, MaxPrice: 30_000_00},
					},
				},
				{
					Code:  "appliances",
					Title: "Бытовая техника",
					Products: []ProductTemplate{
						{Title: "Робот-пылесос", Description: "Работает без нареканий, расходники заменены", MinPrice: 7_000_00, MaxPrice: 45_000_00},
						{Title: "Кофемашина", Description: "Регулярно обслуживалась и очищалась", MinPrice: 9_000_00, MaxPrice: 70_000_00},
					},
				},
			},
		},
		{
			Code:  CategoryTransport,
			Title: "Транспорт",
			Subcategories: []SubcategorySpec{
				{
					Code:  "bicycles",
					Title: "Велосипеды",
					Products: []ProductTemplate{
						{Title: "Горный велосипед", Description: "Настроен и готов к сезону", MinPrice: 12_000_00, MaxPrice: 85_000_00},
						{Title: "Городской велосипед", Description: "Удобная посадка, есть багажник", MinPrice: 9_000_00, MaxPrice: 55_000_00},
					},
				},
				{
					Code:  "parts",
					Title: "Запчасти и аксессуары",
					Products: []ProductTemplate{
						{Title: "Комплект колёс", Description: "Без повреждений, ровные диски", MinPrice: 10_000_00, MaxPrice: 120_000_00},
						{Title: "Автомагнитола", Description: "Все функции работают, полный комплект", MinPrice: 3_000_00, MaxPrice: 35_000_00},
					},
				},
			},
		},
		{
			Code:  CategoryHobby,
			Title: "Хобби и отдых",
			Subcategories: []SubcategorySpec{
				{
					Code:  "sports",
					Title: "Спорт и отдых",
					Products: []ProductTemplate{
						{Title: "Палатка", Description: "Вместительная палатка для поездок и походов", MinPrice: 4_000_00, MaxPrice: 35_000_00},
						{Title: "Сноуборд", Description: "Скользяк обслужен, крепления в комплекте", MinPrice: 8_000_00, MaxPrice: 60_000_00},
					},
				},
				{
					Code:  "music",
					Title: "Музыкальные инструменты",
					Products: []ProductTemplate{
						{Title: "Акустическая гитара", Description: "Чистый звук, удобный гриф", MinPrice: 6_000_00, MaxPrice: 75_000_00},
						{Title: "Синтезатор", Description: "Клавиши и разъёмы работают исправно", MinPrice: 10_000_00, MaxPrice: 90_000_00},
					},
				},
			},
		},
		{
			Code:  CategoryFashion,
			Title: "Одежда и аксессуары",
			Subcategories: []SubcategorySpec{
				{
					Code:  "clothes",
					Title: "Одежда",
					Products: []ProductTemplate{
						{Title: "Куртка", Description: "Аккуратное состояние, без заметных дефектов", MinPrice: 2_500_00, MaxPrice: 25_000_00},
						{Title: "Худи", Description: "Мягкий материал и свободная посадка", MinPrice: 1_500_00, MaxPrice: 12_000_00},
					},
				},
				{
					Code:  "accessories",
					Title: "Аксессуары",
					Products: []ProductTemplate{
						{Title: "Рюкзак", Description: "Прочный материал, удобные отделения", MinPrice: 1_800_00, MaxPrice: 18_000_00},
						{Title: "Наручные часы", Description: "Точный ход, ремешок в хорошем состоянии", MinPrice: 3_000_00, MaxPrice: 80_000_00},
					},
				},
			},
		},
		{
			Code:  CategoryRealEstate,
			Title: "Недвижимость",
			Subcategories: []SubcategorySpec{
				{
					Code:  "apartments",
					Title: "Квартиры",
					Products: []ProductTemplate{
						{Title: "Квартира", Description: "Светлая квартира в районе с развитой инфраструктурой", MinPrice: 3_500_000_00, MaxPrice: 18_000_000_00},
					},
				},
				{
					Code:  "country_houses",
					Title: "Дома и дачи",
					Products: []ProductTemplate{
						{Title: "Загородный дом", Description: "Участок ухожен, круглогодичный подъезд", MinPrice: 2_500_000_00, MaxPrice: 25_000_000_00},
					},
				},
			},
		},
	}}
}
