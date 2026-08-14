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
	Titles      []string
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
						{
							Titles: []string{
								"iPhone 13 128 ГБ, синий",
								"iPhone 14 128 ГБ, чёрный",
								"Samsung Galaxy S23 256 ГБ",
								"Google Pixel 7 128 ГБ",
							},
							Description: "Полностью исправен, бережное использование",
							MinPrice:    15_000_00,
							MaxPrice:    95_000_00,
						},
						{
							Titles: []string{
								"Samsung Galaxy A54 128 ГБ",
								"Xiaomi Redmi Note 12 128 ГБ",
								"OnePlus Nord 2T 128 ГБ",
								"Huawei P40 128 ГБ",
							},
							Description: "Комплект с зарядным устройством и коробкой",
							MinPrice:    8_000_00,
							MaxPrice:    65_000_00,
						},
					},
				},
				{
					Code:  "computers",
					Title: "Ноутбуки",
					Products: []ProductTemplate{
						{
							Titles: []string{
								"Lenovo IdeaPad 5, Ryzen 5, 16/512 ГБ",
								"ASUS VivoBook 15, Core i5, 16/512 ГБ",
								"MSI Modern 14, Core i5, 16/512 ГБ",
								"Honor MagicBook X16, 16/512 ГБ",
							},
							Description: "Подходит для работы, учёбы и повседневных задач",
							MinPrice:    25_000_00,
							MaxPrice:    160_000_00,
						},
						{
							Titles: []string{
								"MacBook Air 13 M2, 8/256 ГБ",
								"Huawei MateBook 14, Core i5, 16/512 ГБ",
								"ASUS Zenbook 14 OLED, 16/512 ГБ",
								"Lenovo Yoga Slim 7, 16/512 ГБ",
							},
							Description: "Лёгкий корпус, аккумулятор хорошо держит заряд",
							MinPrice:    35_000_00,
							MaxPrice:    180_000_00,
						},
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
						{
							Titles: []string{
								"Стеллаж IKEA KALLAX, белый, 4 секции",
								"Стеллаж металлический 180×90 см",
								"Стеллаж в стиле лофт, дуб и металл",
								"Узкий стеллаж для ванной, 35×30 см",
							},
							Description: "В хорошем состоянии, легко собирается",
							MinPrice:    2_000_00,
							MaxPrice:    18_000_00,
						},
						{
							Titles: []string{
								"Письменный стол 120×60 см, дуб",
								"Компьютерный стол с тумбой, белый",
								"Рабочий стол IKEA MICKE, 105×50 см",
								"Стол в стиле лофт, 140×70 см",
							},
							Description: "Устойчивый стол для дома или офиса",
							MinPrice:    3_500_00,
							MaxPrice:    30_000_00,
						},
					},
				},
				{
					Code:  "appliances",
					Title: "Бытовая техника",
					Products: []ProductTemplate{
						{
							Titles: []string{
								"Робот-пылесос Xiaomi S10",
								"Робот-пылесос Dreame D10 Plus",
								"Робот-пылесос Roborock Q7 Max",
								"Робот-пылесос Samsung Jet Bot",
							},
							Description: "Работает без нареканий, расходники заменены",
							MinPrice:    7_000_00,
							MaxPrice:    45_000_00,
						},
						{
							Titles: []string{
								"Кофемашина De'Longhi Magnifica S",
								"Кофемашина Philips LatteGo 5400",
								"Кофемашина Bosch VeroCup 100",
								"Кофемашина Melitta Caffeo Solo",
							},
							Description: "Регулярно обслуживалась и очищалась",
							MinPrice:    9_000_00,
							MaxPrice:    70_000_00,
						},
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
						{
							Titles: []string{
								"Горный велосипед Stels Navigator 500, рама 18″",
								"Горный велосипед Trek Marlin 5, рама M",
								"Горный велосипед Merida Big Nine 20",
								"Горный велосипед Forward Apache 29",
							},
							Description: "Настроен и готов к сезону",
							MinPrice:    12_000_00,
							MaxPrice:    85_000_00,
						},
						{
							Titles: []string{
								"Городской велосипед Shulz Goa 3",
								"Городской велосипед Stels Navigator 350",
								"Городской велосипед Schwinn Sierra",
								"Велосипед круизер Electra Cruiser 1",
							},
							Description: "Удобная посадка, есть багажник",
							MinPrice:    9_000_00,
							MaxPrice:    55_000_00,
						},
					},
				},
				{
					Code:  "parts",
					Title: "Запчасти и аксессуары",
					Products: []ProductTemplate{
						{
							Titles: []string{
								"Комплект литых дисков R17, 5×114.3",
								"Колёса R16 на летней резине 205/55",
								"Колёса R19 для кроссовера, 235/55",
								"Штампованные диски R16, 5×112",
							},
							Description: "Без повреждений, ровные диски",
							MinPrice:    10_000_00,
							MaxPrice:    120_000_00,
						},
						{
							Titles: []string{
								"Автомагнитола Pioneer MVH-S520BT",
								"Магнитола Sony DSX-A410BT",
								"Автомагнитола Alpine UTE-200BT",
								"Магнитола JVC KD-X382BT",
							},
							Description: "Все функции работают, полный комплект",
							MinPrice:    3_000_00,
							MaxPrice:    35_000_00,
						},
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
						{
							Titles: []string{
								"Палатка туристическая трёхместная Naturehike",
								"Палатка Quechua 2 Seconds, 3 места",
								"Палатка Alexika Scout 2, двухместная",
								"Палатка четырёхместная Coleman Darwin",
							},
							Description: "Вместительная палатка для поездок и походов",
							MinPrice:    4_000_00,
							MaxPrice:    35_000_00,
						},
						{
							Titles: []string{
								"Сноуборд Burton Process 155 см",
								"Сноуборд Salomon Pulse 158 см",
								"Сноуборд Jones Mountain Twin 157 см",
								"Сноуборд Rossignol District 156 см",
							},
							Description: "Скользяк обслужен, крепления в комплекте",
							MinPrice:    8_000_00,
							MaxPrice:    60_000_00,
						},
					},
				},
				{
					Code:  "music",
					Title: "Музыкальные инструменты",
					Products: []ProductTemplate{
						{
							Titles: []string{
								"Акустическая гитара Yamaha F310",
								"Электроакустическая гитара Cort AD810E",
								"Акустическая гитара Ibanez V50NJP",
								"Гитара Takamine GD11M-NS",
							},
							Description: "Чистый звук, удобный гриф",
							MinPrice:    6_000_00,
							MaxPrice:    75_000_00,
						},
						{
							Titles: []string{
								"Синтезатор Yamaha PSR-E373",
								"Синтезатор Casio CT-S300",
								"Цифровое пианино Casio CDP-S110",
								"Синтезатор Yamaha MX49",
							},
							Description: "Клавиши и разъёмы работают исправно",
							MinPrice:    10_000_00,
							MaxPrice:    90_000_00,
						},
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
						{
							Titles: []string{
								"Куртка The North Face, размер M",
								"Демисезонная куртка Columbia, размер S",
								"Парка Levi's, размер L",
								"Ветровка Nike ACG, размер XL",
							},
							Description: "Аккуратное состояние, без заметных дефектов",
							MinPrice:    2_500_00,
							MaxPrice:    25_000_00,
						},
						{
							Titles: []string{
								"Худи Uniqlo, размер L, серое",
								"Худи Nike Sportswear, размер M",
								"Толстовка Adidas Originals, размер S",
								"Худи Carhartt WIP, размер XL",
							},
							Description: "Мягкий материал и свободная посадка",
							MinPrice:    1_500_00,
							MaxPrice:    12_000_00,
						},
					},
				},
				{
					Code:  "accessories",
					Title: "Аксессуары",
					Products: []ProductTemplate{
						{
							Titles: []string{
								"Рюкзак Fjallraven Kanken, 16 л",
								"Городской рюкзак The North Face Borealis",
								"Рюкзак для ноутбука Xiaomi, 20 л",
								"Рюкзак Herschel Little America",
							},
							Description: "Прочный материал, удобные отделения",
							MinPrice:    1_800_00,
							MaxPrice:    18_000_00,
						},
						{
							Titles: []string{
								"Часы Casio G-Shock GA-2100",
								"Часы Seiko 5 Sports Automatic",
								"Часы Tissot PRX Quartz 40 мм",
								"Apple Watch Series 8, 45 мм",
							},
							Description: "Точный ход, ремешок в хорошем состоянии",
							MinPrice:    3_000_00,
							MaxPrice:    80_000_00,
						},
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
						{
							Titles: []string{
								"1-к. квартира, 38 м², 6/12 эт.",
								"2-к. квартира, 54 м², 7/16 эт.",
								"3-к. квартира, 76 м², 4/9 эт.",
								"Студия, 28 м², 12/25 эт.",
							},
							Description: "Светлая квартира в районе с развитой инфраструктурой",
							MinPrice:    3_500_000_00,
							MaxPrice:    18_000_000_00,
						},
					},
				},
				{
					Code:  "country_houses",
					Title: "Дома и дачи",
					Products: []ProductTemplate{
						{
							Titles: []string{
								"Дом 120 м² на участке 8 сот.",
								"Коттедж 180 м² на участке 12 сот.",
								"Дача 65 м² на участке 6 сот.",
								"Таунхаус 110 м² с участком 3 сот.",
							},
							Description: "Участок ухожен, круглогодичный подъезд",
							MinPrice:    2_500_000_00,
							MaxPrice:    25_000_000_00,
						},
					},
				},
			},
		},
	}}
}
