package seeds

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
	"gorm.io/gorm"
)

// Run seeds the database with initial data.
func Run(db *gorm.DB) error {
	if err := seedProvinces(db); err != nil {
		return fmt.Errorf("seed provinces: %w", err)
	}
	if err := seedCities(db); err != nil {
		return fmt.Errorf("seed cities: %w", err)
	}
	if err := seedDistrictsAndFacilities(db); err != nil {
		return fmt.Errorf("seed districts/facilities: %w", err)
	}
	if err := seedUsers(db); err != nil {
		return fmt.Errorf("seed users: %w", err)
	}
	if err := seedProperties(db); err != nil {
		return fmt.Errorf("seed properties: %w", err)
	}

	fmt.Println("All seeds completed successfully!")
	return nil
}

func seedProvinces(db *gorm.DB) error {
	var count int64
	db.Model(&entity.Province{}).Count(&count)
	if count > 0 {
		fmt.Println("Provinces already seeded, skipping...")
		return nil
	}

	provinces := []entity.Province{
		{Name: "DKI Jakarta"},
		{Name: "Jawa Barat"},
		{Name: "Jawa Tengah"},
		{Name: "Jawa Timur"},
		{Name: "Banten"},
		{Name: "DI Yogyakarta"},
		{Name: "Bali"},
		{Name: "Sumatera Utara"},
		{Name: "Sumatera Selatan"},
		{Name: "Sulawesi Selatan"},
	}

	for i := range provinces {
		if err := db.Create(&provinces[i]).Error; err != nil {
			return err
		}
	}
	fmt.Printf("Seeded %d provinces\n", len(provinces))
	return nil
}

func seedCities(db *gorm.DB) error {
	var count int64
	db.Model(&entity.City{}).Count(&count)
	if count > 0 {
		fmt.Println("Cities already seeded, skipping...")
		return nil
	}

	// Get province IDs
	var provinces []entity.Province
	db.Find(&provinces)
	provinceMap := make(map[string]entity.Province)
	for _, p := range provinces {
		provinceMap[p.Name] = p
	}

	type cityData struct {
		Province string
		City     string
	}

	cities := []cityData{
		{"DKI Jakarta", "Jakarta Selatan"},
		{"DKI Jakarta", "Jakarta Pusat"},
		{"DKI Jakarta", "Jakarta Barat"},
		{"DKI Jakarta", "Jakarta Timur"},
		{"DKI Jakarta", "Jakarta Utara"},
		{"Jawa Barat", "Bandung"},
		{"Jawa Barat", "Bekasi"},
		{"Jawa Barat", "Depok"},
		{"Jawa Barat", "Bogor"},
		{"Jawa Tengah", "Semarang"},
		{"Jawa Tengah", "Solo"},
		{"Jawa Timur", "Surabaya"},
		{"Jawa Timur", "Malang"},
		{"Banten", "Tangerang"},
		{"Banten", "Tangerang Selatan"},
		{"DI Yogyakarta", "Yogyakarta"},
		{"DI Yogyakarta", "Sleman"},
		{"Bali", "Denpasar"},
		{"Bali", "Badung"},
		{"Sumatera Utara", "Medan"},
		{"Sulawesi Selatan", "Makassar"},
	}

	for _, c := range cities {
		prov, ok := provinceMap[c.Province]
		if !ok {
			continue
		}
		city := entity.City{ProvinceID: prov.ID, Name: c.City}
		if err := db.Create(&city).Error; err != nil {
			return err
		}
	}
	fmt.Printf("Seeded %d cities\n", len(cities))
	return nil
}

func seedUsers(db *gorm.DB) error {
	var count int64
	db.Model(&entity.User{}).Count(&count)
	if count > 0 {
		fmt.Println("Users already seeded, skipping...")
		return nil
	}

	users := []struct {
		Name     string
		Email    string
		Phone    string
		Password string
		Role     string
	}{
		{"Super Admin", "admin@leventa.com", "081200000001", "admin123", entity.RoleAdmin},
		{"Ahmad Sales", "sales@leventa.com", "081200000002", "sales123", entity.RoleSales},
		{"Budi Sales", "sales2@leventa.com", "081200000003", "sales123", entity.RoleSales},
		{"Sari Sales", "sales3@leventa.com", "081200000007", "sales123", entity.RoleSales},
	}

	for _, u := range users {
		user := &entity.User{
			Name:           u.Name,
			Email:          u.Email,
			Phone:          u.Phone,
			WhatsAppNumber: waFor(u.Role, u.Phone),
			Role:           u.Role,
			Status:         entity.UserStatusActive,
		}
		if err := user.SetPassword(u.Password); err != nil {
			return err
		}
		if err := db.Create(user).Error; err != nil {
			return err
		}
	}
	fmt.Printf("Seeded %d users\n", len(users))
	return nil
}

// waFor returns a normalized WhatsApp number for marketing users only.
func waFor(role, phone string) string {
	if role != entity.RoleSales {
		return ""
	}
	if len(phone) > 1 && phone[0] == '0' {
		return "62" + phone[1:]
	}
	return phone
}

// seedProperties seeds 5 fully detailed listings modeled on the reference
// listing page (docs/syahira_prop_desc.png): documents, room specs, utilities,
// outdoor features, distances to public facilities, and a 3-photo gallery.
func seedProperties(db *gorm.DB) error {
	var count int64
	db.Model(&entity.Property{}).Count(&count)
	if count > 0 {
		fmt.Println("Properties already seeded, skipping...")
		return nil
	}

	var sales1, sales2 entity.User
	if err := db.Where("email = ?", "sales@leventa.com").First(&sales1).Error; err != nil {
		return err
	}
	if err := db.Where("email = ?", "sales2@leventa.com").First(&sales2).Error; err != nil {
		return err
	}

	// loc resolves province/city/district IDs by name; missing rows stay nil.
	loc := func(province, city, district string) (p, c, d *uuid.UUID) {
		var prov entity.Province
		if db.Where("name = ?", province).First(&prov).Error == nil {
			p = &prov.ID
		}
		var ct entity.City
		if db.Where("name = ?", city).First(&ct).Error == nil {
			c = &ct.ID
		}
		var dt entity.District
		if district != "" && c != nil && db.Where("city_id = ? AND name = ?", ct.ID, district).First(&dt).Error == nil {
			d = &dt.ID
		}
		return
	}

	f := func(v float64) *float64 { return &v }
	i := func(v int) *int { return &v }

	type seed struct {
		marketing entity.User
		province  string
		city      string
		district  string
		owner     entity.Owner
		property  entity.Property
		images    []string
	}

	// shortcut: seed photos are external Unsplash URLs, not files in /uploads; replace with uploaded images for production data.
	seeds := []seed{
		{
			marketing: sales1, province: "Jawa Barat", city: "Bogor",
			owner: entity.Owner{Name: "Bapak Andi Wijaya", Email: "andi.wijaya@email.com", Phone: "081311112222", PrivateAddress: "Jl. Mawar Raya No. 7, Cibubur"},
			property: entity.Property{
				Code: "PROP-BGR-001", Name: "Rumah diJual Cluster Mawar", Type: entity.PropertyTypeHouse,
				TransactionType: entity.TransactionTypeSale, Status: entity.PropertyStatusAvailable, Price: f(1400000000),
				Description:      "Rumah 2 lantai di Cluster Mawar, Kota Wisata Cibubur. Lingkungan asri, one gate system, dekat sekolah, rumah sakit, dan akses tol Cimanggis-Cibitung.",
				ShortDescription: "Dijual rumah yang asri dan siap huni",
				GeneralLocation:  "Kota Wisata Cibubur", AreaLabel: "Cibubur", SubDistrict: "Ciangsana", PostalCode: "16968",
				ClusterName: "Cluster Mawar", StreetName: "Jl. Mawar Raya", BlockNumber: "A3", HouseNumber: "7",
				Latitude: f(-6.3958), Longitude: f(106.9436), Orientation: "Timur", MaintenanceFee: f(1000000),
				LandArea: f(120), BuildingArea: f(90), Floor: i(2), Bedrooms: 6, Bathrooms: 6, YearBuilt: i(2021),
				CertSHM:        true,
				MasterBedrooms: 3, ChildBedrooms: 2, OtherBedrooms: 1, MasterBathrooms: 3, ChildBathrooms: 2, OtherBathrooms: 1,
				LivingRooms: 1, DiningRooms: 1, FamilyRooms: 1, StudyRooms: 1, Kitchens: 1,
				ElectricityCapacity: "2200 VA", AirConditioners: 1, TVProvider: "Indihome", WaterSource: "PAM",
				FurnishingStatus: entity.FurnishingFurnished,
				HasPool:          true, HasGarage: true, HasFrontYard: true, HasBackYard: true, HasBalcony: true,
				DistSchool: i(100), DistBusTerminal: i(600), DistMrtLrt: i(700), DistToll: i(500), DistHospital: i(500),
			},
			images: []string{
				"https://images.unsplash.com/photo-1600596542815-ffad4c1539a9?w=1600",
				"https://images.unsplash.com/photo-1600585154340-be6161a56a0c?w=1600",
				"https://images.unsplash.com/photo-1600566753190-17f0baa2a6c3?w=1600",
			},
		},
		{
			marketing: sales1, province: "Jawa Barat", city: "Bogor",
			owner: entity.Owner{Name: "Ibu Ratna Dewi", Email: "ratna.dewi@email.com", Phone: "081298765432", PrivateAddress: "Jl. Melati No. 15, Bogor"},
			property: entity.Property{
				Code: "PROP-BGR-002", Name: "Rumah di Cluster 01", Type: entity.PropertyTypeHouse,
				TransactionType: entity.TransactionTypeSale, Status: entity.PropertyStatusAvailable, Price: f(900000000),
				Description:      "Rumah 2 lantai kondisi baik di Cluster 01 Cibubur. Luas dan berada di lokasi strategis dekat pusat perbelanjaan dan stasiun LRT.",
				ShortDescription: "Dijual rumah dengan kondisi baik, luas dan lokasi strategis",
				GeneralLocation:  "Cibubur, Bogor", AreaLabel: "Cibubur", SubDistrict: "Jatikarya", PostalCode: "17435",
				ClusterName: "Cluster 01", StreetName: "Jl. Alternatif Cibubur", BlockNumber: "C1", HouseNumber: "21",
				Latitude: f(-6.3825), Longitude: f(106.9102), Orientation: "Utara", MaintenanceFee: f(500000),
				LandArea: f(96), BuildingArea: f(72), Floor: i(2), Bedrooms: 3, Bathrooms: 2, YearBuilt: i(2019),
				CertSHM: true, CertIMB: true, CertPBB: true,
				MasterBedrooms: 1, ChildBedrooms: 1, OtherBedrooms: 1, MasterBathrooms: 1, OtherBathrooms: 1,
				LivingRooms: 1, DiningRooms: 1, FamilyRooms: 1, Kitchens: 1,
				ElectricityCapacity: "1300 VA", AirConditioners: 2, TVProvider: "First Media", WaterSource: "PAM",
				FurnishingStatus: entity.FurnishingSemiFurnished,
				HasGarage:        true, HasFrontYard: true, HasBalcony: true,
				DistSchool: i(300), DistBusTerminal: i(1500), DistMrtLrt: i(1200), DistToll: i(800), DistHospital: i(1000),
			},
			images: []string{
				"https://images.unsplash.com/photo-1568605114967-8130f3a36994?w=1600",
				"https://images.unsplash.com/photo-1600210492486-724fe5c67fb0?w=1600",
				"https://images.unsplash.com/photo-1556911220-bff31c812dba?w=1600",
			},
		},
		{
			marketing: sales1, province: "DKI Jakarta", city: "Jakarta Selatan", district: "Kebayoran Baru",
			owner: entity.Owner{Name: "PT Graha Sejahtera", Email: "info@grahasejahtera.co.id", Phone: "02112345678", PrivateAddress: "Jl. Sudirman No. 1, Jakarta"},
			property: entity.Property{
				Code: "PROP-JKT-001", Name: "Rumah Mewah Kebayoran", Type: entity.PropertyTypeHouse,
				TransactionType: entity.TransactionTypeSale, Status: entity.PropertyStatusAvailable, Price: f(2500000000),
				Description:      "Rumah mewah 2 lantai di Kebayoran Baru dengan private pool, dekat Senayan City dan stasiun MRT.",
				ShortDescription: "Rumah mewah siap huni di jantung Jakarta Selatan",
				GeneralLocation:  "Kebayoran Baru, Jakarta Selatan", AreaLabel: "Kebayoran Baru", SubDistrict: "Senayan", PostalCode: "12190",
				StreetName: "Jl. Kebayoran Baru", HouseNumber: "12", Latitude: f(-6.2615), Longitude: f(106.8106),
				Orientation: "Selatan", MaintenanceFee: f(2500000),
				LandArea: f(200), BuildingArea: f(150), Floor: i(2), Bedrooms: 4, Bathrooms: 4, YearBuilt: i(2020),
				CertSHM: true, CertIMB: true, CertPBB: true,
				MasterBedrooms: 2, ChildBedrooms: 1, OtherBedrooms: 1, MasterBathrooms: 2, ChildBathrooms: 1, OtherBathrooms: 1,
				LivingRooms: 1, DiningRooms: 1, FamilyRooms: 1, StudyRooms: 1, Kitchens: 2,
				ElectricityCapacity: "5500 VA", AirConditioners: 5, TVProvider: "Biznet", WaterSource: "PAM",
				FurnishingStatus: entity.FurnishingFurnished,
				HasPool:          true, HasGarage: true, HasFrontYard: true, HasBackYard: true, HasBalcony: true,
				DistSchool: i(400), DistBusTerminal: i(2000), DistMrtLrt: i(600), DistToll: i(1200), DistHospital: i(900),
			},
			images: []string{
				"https://images.unsplash.com/photo-1613490493576-7fde63acd811?w=1600",
				"https://images.unsplash.com/photo-1600607687939-ce8a6c25118c?w=1600",
				"https://images.unsplash.com/photo-1600566753086-00f18fb6b3ea?w=1600",
			},
		},
		{
			marketing: sales2, province: "Jawa Barat", city: "Bandung", district: "Coblong",
			owner: entity.Owner{Name: "CV Maju Properti", Email: "contact@majuproperti.co.id", Phone: "02287654321", PrivateAddress: "Jl. Diponegoro No. 10, Bandung"},
			property: entity.Property{
				Code: "PROP-BDG-001", Name: "Rumah Asri Dago Atas", Type: entity.PropertyTypeHouse,
				TransactionType: entity.TransactionTypeRent, Status: entity.PropertyStatusAvailable, RentalPrice: f(15000000),
				Description:      "Rumah sejuk di Dago Atas dengan view kota Bandung, halaman luas, cocok untuk keluarga.",
				ShortDescription: "Disewakan rumah sejuk dengan view kota Bandung",
				GeneralLocation:  "Dago, Bandung", AreaLabel: "Dago", SubDistrict: "Dago", PostalCode: "40135",
				StreetName: "Jl. Ir. H. Juanda", HouseNumber: "88", Latitude: f(-6.8706), Longitude: f(107.6181),
				Orientation: "Barat", MaintenanceFee: f(750000),
				LandArea: f(250), BuildingArea: f(180), Floor: i(2), Bedrooms: 4, Bathrooms: 3, YearBuilt: i(2017),
				CertSHM: true, CertIMB: true,
				MasterBedrooms: 1, ChildBedrooms: 2, OtherBedrooms: 1, MasterBathrooms: 1, ChildBathrooms: 1, OtherBathrooms: 1,
				LivingRooms: 1, DiningRooms: 1, FamilyRooms: 1, StudyRooms: 1, Kitchens: 1,
				ElectricityCapacity: "3500 VA", AirConditioners: 0, TVProvider: "Indihome", WaterSource: "Sumur Bor",
				FurnishingStatus: entity.FurnishingFurnished,
				HasGarage:        true, HasFrontYard: true, HasBackYard: true, HasBalcony: true,
				DistSchool: i(500), DistBusTerminal: i(4000), DistToll: i(6000), DistHospital: i(1500),
			},
			images: []string{
				"https://images.unsplash.com/photo-1572120360610-d971b9d7767c?w=1600",
				"https://images.unsplash.com/photo-1586023492125-27b2c045efd7?w=1600",
				"https://images.unsplash.com/photo-1600585154526-990dced4db0d?w=1600",
			},
		},
		{
			marketing: sales2, province: "Bali", city: "Denpasar", district: "Denpasar Barat",
			owner: entity.Owner{Name: "Bapak Hartono", Email: "hartono@email.com", Phone: "081345678901", PrivateAddress: "Jl. Malioboro No. 5, Yogyakarta"},
			property: entity.Property{
				Code: "PROP-BALI-001", Name: "Villa Sunset Seminyak", Type: entity.PropertyTypeHouse,
				TransactionType: entity.TransactionTypeSale, Status: entity.PropertyStatusAvailable, Price: f(3200000000),
				Description:      "Villa tropis dengan private pool di Seminyak, view sunset, 10 menit ke pantai.",
				ShortDescription: "Villa tropis dengan private pool dekat pantai",
				GeneralLocation:  "Seminyak, Bali", AreaLabel: "Seminyak", SubDistrict: "Padangsambian", PostalCode: "80117",
				StreetName: "Jl. Sunset Road", HouseNumber: "99", Latitude: f(-8.6705), Longitude: f(115.2126),
				Orientation: "Barat", MaintenanceFee: f(3000000),
				LandArea: f(500), BuildingArea: f(350), Floor: i(2), Bedrooms: 3, Bathrooms: 4, YearBuilt: i(2022),
				CertSHM: true, CertIMB: true, CertPBB: true,
				MasterBedrooms: 3, MasterBathrooms: 3, OtherBathrooms: 1,
				LivingRooms: 1, DiningRooms: 1, FamilyRooms: 1, Kitchens: 1,
				ElectricityCapacity: "7700 VA", AirConditioners: 4, TVProvider: "Biznet", WaterSource: "PDAM",
				FurnishingStatus: entity.FurnishingFurnished,
				HasPool:          true, HasGarage: true, HasFrontYard: true, HasBackYard: true, HasBalcony: true,
				DistSchool: i(1200), DistBusTerminal: i(5000), DistToll: i(7000), DistHospital: i(2000),
			},
			images: []string{
				"https://images.unsplash.com/photo-1582268611958-ebfd161ef9cf?w=1600",
				"https://images.unsplash.com/photo-1602343168117-bb8ffe3e2e9f?w=1600",
				"https://images.unsplash.com/photo-1600210491892-03d54c0aaf87?w=1600",
			},
		},
	}

	for _, sd := range seeds {
		err := db.Transaction(func(tx *gorm.DB) error {
			owner := sd.owner
			owner.CreatedBy = &sd.marketing.ID
			if err := tx.Create(&owner).Error; err != nil {
				return err
			}
			prop := sd.property
			prop.OwnerID = owner.ID
			prop.PrimaryMarketingID = sd.marketing.ID
			prop.ProvinceID, prop.CityID, prop.DistrictID = loc(sd.province, sd.city, sd.district)
			if len(sd.images) > 0 {
				prop.Thumbnail = sd.images[0]
			}
			if err := tx.Omit("Owner", "PrimaryMarketing", "Province", "City", "District", "Images", "Facilities").Create(&prop).Error; err != nil {
				return err
			}
			for idx, url := range sd.images {
				img := entity.PropertyImage{PropertyID: prop.ID, URL: url, SortOrder: idx, IsThumbnail: idx == 0}
				if err := tx.Create(&img).Error; err != nil {
					return err
				}
			}
			return tx.Create(&entity.PropertyMarketingAssignment{
				PropertyID: prop.ID, MarketingID: sd.marketing.ID,
				AssignmentType: entity.AssignmentTypePrimary, AssignedBy: sd.marketing.ID,
			}).Error
		})
		if err != nil {
			return err
		}
	}
	fmt.Printf("Seeded %d properties\n", len(seeds))
	return nil
}

func seedDistrictsAndFacilities(db *gorm.DB) error {
	var n int64
	db.Model(&entity.Facility{}).Count(&n)
	if n == 0 {
		for _, name := range []string{"Kolam Renang", "Carport", "Taman", "Keamanan 24 Jam", "CCTV", "Gym", "Lift", "Dekat Transportasi Umum"} {
			if err := db.Create(&entity.Facility{Name: name}).Error; err != nil {
				return err
			}
		}
	}
	db.Model(&entity.District{}).Count(&n)
	if n > 0 {
		return nil
	}
	districts := map[string][]string{
		"Jakarta Selatan": {"Kebayoran Baru", "Tebet", "Cilandak"},
		"Bandung":         {"Coblong", "Cicendo"},
		"Denpasar":        {"Denpasar Selatan", "Denpasar Barat"},
	}
	for city, names := range districts {
		var c entity.City
		if err := db.Where("name = ?", city).First(&c).Error; err != nil {
			continue
		}
		for _, name := range names {
			if err := db.Create(&entity.District{CityID: c.ID, Name: name}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
