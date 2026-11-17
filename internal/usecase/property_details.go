package usecase

import (
	"strings"

	"github.com/leventa-backend/internal/domain/entity"
)

// PropertyDetails holds the listing detail fields of a new property (D-028).
type PropertyDetails struct {
	ShortDescription    string               `json:"short_description" validate:"omitempty,max=500"`
	AreaLabel           string               `json:"area_label" validate:"omitempty,max=255"`
	SubDistrict         string               `json:"sub_district" validate:"omitempty,max=255"`
	PostalCode          string               `json:"postal_code" validate:"omitempty,max=10"`
	Orientation         string               `json:"orientation" validate:"omitempty,max=50"`
	MaintenanceFee      *float64             `json:"maintenance_fee" validate:"omitempty,gte=0"`
	CertSHM             bool                 `json:"cert_shm"`
	CertHGB             bool                 `json:"cert_hgb"`
	CertIMB             bool                 `json:"cert_imb"`
	CertPBB             bool                 `json:"cert_pbb"`
	CertGirik           bool                 `json:"cert_girik"`
	CertOther           bool                 `json:"cert_other"`
	CertSHMSRS          bool                 `json:"cert_shmsrs"`
	CertHGU             bool                 `json:"cert_hgu"`
	CertSLF             bool                 `json:"cert_slf"`
	LivingRooms         int                  `json:"living_rooms" validate:"gte=0,lte=100"`
	DiningRooms         int                  `json:"dining_rooms" validate:"gte=0,lte=100"`
	MasterBedrooms      int                  `json:"master_bedrooms" validate:"gte=0,lte=100"`
	ChildBedrooms       int                  `json:"child_bedrooms" validate:"gte=0,lte=100"`
	OtherBedrooms       int                  `json:"other_bedrooms" validate:"gte=0,lte=100"`
	MasterBathrooms     int                  `json:"master_bathrooms" validate:"gte=0,lte=100"`
	ChildBathrooms      int                  `json:"child_bathrooms" validate:"gte=0,lte=100"`
	OtherBathrooms      int                  `json:"other_bathrooms" validate:"gte=0,lte=100"`
	FamilyRooms         int                  `json:"family_rooms" validate:"gte=0,lte=100"`
	StudyRooms          int                  `json:"study_rooms" validate:"gte=0,lte=100"`
	Kitchens            int                  `json:"kitchens" validate:"gte=0,lte=100"`
	ElectricityCapacity string               `json:"electricity_capacity" validate:"omitempty,max=50"`
	AirConditioners     int                  `json:"air_conditioners" validate:"gte=0,lte=1000"`
	TVProvider          string               `json:"tv_provider" validate:"omitempty,max=100"`
	WaterSource         string               `json:"water_source" validate:"omitempty,max=100"`
	FurnishingStatus    string               `json:"furnishing_status" validate:"omitempty,oneof=unfurnished semi_furnished furnished"`
	HasPool             bool                 `json:"has_pool"`
	HasGarage           bool                 `json:"has_garage"`
	HasFrontYard        bool                 `json:"has_front_yard"`
	HasBackYard         bool                 `json:"has_back_yard"`
	HasBalcony          bool                 `json:"has_balcony"`
	DistSchool          *int                 `json:"dist_school" validate:"omitempty,gte=0"`
	DistBusTerminal     *int                 `json:"dist_bus_terminal" validate:"omitempty,gte=0"`
	DistMrtLrt          *int                 `json:"dist_mrt_lrt" validate:"omitempty,gte=0"`
	DistToll            *int                 `json:"dist_toll" validate:"omitempty,gte=0"`
	DistHospital        *int                 `json:"dist_hospital" validate:"omitempty,gte=0"`
	ExtraOutdoor        []string             `json:"extra_outdoor" validate:"max=10,dive,required,max=50"`
	ExtraDistances      []entity.NearbyPlace `json:"extra_distances" validate:"max=10,dive"`
}

// applyTo copies the details onto the entity.
func (d PropertyDetails) applyTo(p *entity.Property) {
	p.ShortDescription = d.ShortDescription
	p.AreaLabel = d.AreaLabel
	p.SubDistrict = d.SubDistrict
	p.PostalCode = d.PostalCode
	p.Orientation = d.Orientation
	p.MaintenanceFee = d.MaintenanceFee
	p.CertSHM = d.CertSHM
	p.CertHGB = d.CertHGB
	p.CertIMB = d.CertIMB
	p.CertPBB = d.CertPBB
	p.CertGirik = d.CertGirik
	p.CertOther = d.CertOther
	p.CertSHMSRS = d.CertSHMSRS
	p.CertHGU = d.CertHGU
	p.CertSLF = d.CertSLF
	p.LivingRooms = d.LivingRooms
	p.DiningRooms = d.DiningRooms
	p.MasterBedrooms = d.MasterBedrooms
	p.ChildBedrooms = d.ChildBedrooms
	p.OtherBedrooms = d.OtherBedrooms
	p.MasterBathrooms = d.MasterBathrooms
	p.ChildBathrooms = d.ChildBathrooms
	p.OtherBathrooms = d.OtherBathrooms
	p.FamilyRooms = d.FamilyRooms
	p.StudyRooms = d.StudyRooms
	p.Kitchens = d.Kitchens
	p.ElectricityCapacity = d.ElectricityCapacity
	p.AirConditioners = d.AirConditioners
	p.TVProvider = d.TVProvider
	p.WaterSource = d.WaterSource
	p.FurnishingStatus = d.FurnishingStatus
	p.HasPool = d.HasPool
	p.HasGarage = d.HasGarage
	p.HasFrontYard = d.HasFrontYard
	p.HasBackYard = d.HasBackYard
	p.HasBalcony = d.HasBalcony
	p.DistSchool = d.DistSchool
	p.DistBusTerminal = d.DistBusTerminal
	p.DistMrtLrt = d.DistMrtLrt
	p.DistToll = d.DistToll
	p.DistHospital = d.DistHospital
	p.ExtraOutdoor = cleanOutdoor(d.ExtraOutdoor)
	p.ExtraDistances = cleanPlaces(d.ExtraDistances)
}

// PropertyDetailsPatch is the partial-update form: nil means unchanged.
type PropertyDetailsPatch struct {
	ShortDescription    *string  `json:"short_description" validate:"omitempty,max=500"`
	AreaLabel           *string  `json:"area_label" validate:"omitempty,max=255"`
	SubDistrict         *string  `json:"sub_district" validate:"omitempty,max=255"`
	PostalCode          *string  `json:"postal_code" validate:"omitempty,max=10"`
	Orientation         *string  `json:"orientation" validate:"omitempty,max=50"`
	MaintenanceFee      *float64 `json:"maintenance_fee" validate:"omitempty,gte=0"`
	CertSHM             *bool    `json:"cert_shm"`
	CertHGB             *bool    `json:"cert_hgb"`
	CertIMB             *bool    `json:"cert_imb"`
	CertPBB             *bool    `json:"cert_pbb"`
	CertGirik           *bool    `json:"cert_girik"`
	CertOther           *bool    `json:"cert_other"`
	CertSHMSRS          *bool    `json:"cert_shmsrs"`
	CertHGU             *bool    `json:"cert_hgu"`
	CertSLF             *bool    `json:"cert_slf"`
	LivingRooms         *int     `json:"living_rooms" validate:"omitempty,gte=0,lte=100"`
	DiningRooms         *int     `json:"dining_rooms" validate:"omitempty,gte=0,lte=100"`
	MasterBedrooms      *int     `json:"master_bedrooms" validate:"omitempty,gte=0,lte=100"`
	ChildBedrooms       *int     `json:"child_bedrooms" validate:"omitempty,gte=0,lte=100"`
	OtherBedrooms       *int     `json:"other_bedrooms" validate:"omitempty,gte=0,lte=100"`
	MasterBathrooms     *int     `json:"master_bathrooms" validate:"omitempty,gte=0,lte=100"`
	ChildBathrooms      *int     `json:"child_bathrooms" validate:"omitempty,gte=0,lte=100"`
	OtherBathrooms      *int     `json:"other_bathrooms" validate:"omitempty,gte=0,lte=100"`
	FamilyRooms         *int     `json:"family_rooms" validate:"omitempty,gte=0,lte=100"`
	StudyRooms          *int     `json:"study_rooms" validate:"omitempty,gte=0,lte=100"`
	Kitchens            *int     `json:"kitchens" validate:"omitempty,gte=0,lte=100"`
	ElectricityCapacity *string  `json:"electricity_capacity" validate:"omitempty,max=50"`
	AirConditioners     *int     `json:"air_conditioners" validate:"omitempty,gte=0,lte=1000"`
	TVProvider          *string  `json:"tv_provider" validate:"omitempty,max=100"`
	WaterSource         *string  `json:"water_source" validate:"omitempty,max=100"`
	// omitzero lets "" (clear) through behind a pointer; omitempty does not.
	FurnishingStatus *string `json:"furnishing_status" validate:"omitnil,omitzero,oneof=unfurnished semi_furnished furnished"`
	HasPool          *bool   `json:"has_pool"`
	HasGarage        *bool   `json:"has_garage"`
	HasFrontYard     *bool   `json:"has_front_yard"`
	HasBackYard      *bool   `json:"has_back_yard"`
	HasBalcony       *bool   `json:"has_balcony"`
	DistSchool       *int    `json:"dist_school" validate:"omitempty,gte=0"`
	DistBusTerminal  *int    `json:"dist_bus_terminal" validate:"omitempty,gte=0"`
	DistMrtLrt       *int    `json:"dist_mrt_lrt" validate:"omitempty,gte=0"`
	DistToll         *int    `json:"dist_toll" validate:"omitempty,gte=0"`
	DistHospital     *int    `json:"dist_hospital" validate:"omitempty,gte=0"`
	// nil leaves the list unchanged; [] clears it.
	ExtraOutdoor   *[]string             `json:"extra_outdoor" validate:"omitempty,max=10,dive,required,max=50"`
	ExtraDistances *[]entity.NearbyPlace `json:"extra_distances" validate:"omitempty,max=10,dive"`
}

// applyTo applies the non-nil fields onto the entity.
func (d PropertyDetailsPatch) applyTo(p *entity.Property) {
	if d.ShortDescription != nil {
		p.ShortDescription = *d.ShortDescription
	}
	if d.AreaLabel != nil {
		p.AreaLabel = *d.AreaLabel
	}
	if d.SubDistrict != nil {
		p.SubDistrict = *d.SubDistrict
	}
	if d.PostalCode != nil {
		p.PostalCode = *d.PostalCode
	}
	if d.Orientation != nil {
		p.Orientation = *d.Orientation
	}
	if d.MaintenanceFee != nil {
		p.MaintenanceFee = d.MaintenanceFee
	}
	if d.CertSHM != nil {
		p.CertSHM = *d.CertSHM
	}
	if d.CertHGB != nil {
		p.CertHGB = *d.CertHGB
	}
	if d.CertIMB != nil {
		p.CertIMB = *d.CertIMB
	}
	if d.CertPBB != nil {
		p.CertPBB = *d.CertPBB
	}
	if d.CertGirik != nil {
		p.CertGirik = *d.CertGirik
	}
	if d.CertOther != nil {
		p.CertOther = *d.CertOther
	}
	if d.CertSHMSRS != nil {
		p.CertSHMSRS = *d.CertSHMSRS
	}
	if d.CertHGU != nil {
		p.CertHGU = *d.CertHGU
	}
	if d.CertSLF != nil {
		p.CertSLF = *d.CertSLF
	}
	if d.LivingRooms != nil {
		p.LivingRooms = *d.LivingRooms
	}
	if d.DiningRooms != nil {
		p.DiningRooms = *d.DiningRooms
	}
	if d.MasterBedrooms != nil {
		p.MasterBedrooms = *d.MasterBedrooms
	}
	if d.ChildBedrooms != nil {
		p.ChildBedrooms = *d.ChildBedrooms
	}
	if d.OtherBedrooms != nil {
		p.OtherBedrooms = *d.OtherBedrooms
	}
	if d.MasterBathrooms != nil {
		p.MasterBathrooms = *d.MasterBathrooms
	}
	if d.ChildBathrooms != nil {
		p.ChildBathrooms = *d.ChildBathrooms
	}
	if d.OtherBathrooms != nil {
		p.OtherBathrooms = *d.OtherBathrooms
	}
	if d.FamilyRooms != nil {
		p.FamilyRooms = *d.FamilyRooms
	}
	if d.StudyRooms != nil {
		p.StudyRooms = *d.StudyRooms
	}
	if d.Kitchens != nil {
		p.Kitchens = *d.Kitchens
	}
	if d.ElectricityCapacity != nil {
		p.ElectricityCapacity = *d.ElectricityCapacity
	}
	if d.AirConditioners != nil {
		p.AirConditioners = *d.AirConditioners
	}
	if d.TVProvider != nil {
		p.TVProvider = *d.TVProvider
	}
	if d.WaterSource != nil {
		p.WaterSource = *d.WaterSource
	}
	if d.FurnishingStatus != nil {
		p.FurnishingStatus = *d.FurnishingStatus
	}
	if d.HasPool != nil {
		p.HasPool = *d.HasPool
	}
	if d.HasGarage != nil {
		p.HasGarage = *d.HasGarage
	}
	if d.HasFrontYard != nil {
		p.HasFrontYard = *d.HasFrontYard
	}
	if d.HasBackYard != nil {
		p.HasBackYard = *d.HasBackYard
	}
	if d.HasBalcony != nil {
		p.HasBalcony = *d.HasBalcony
	}
	if d.DistSchool != nil {
		p.DistSchool = d.DistSchool
	}
	if d.DistBusTerminal != nil {
		p.DistBusTerminal = d.DistBusTerminal
	}
	if d.DistMrtLrt != nil {
		p.DistMrtLrt = d.DistMrtLrt
	}
	if d.DistToll != nil {
		p.DistToll = d.DistToll
	}
	if d.DistHospital != nil {
		p.DistHospital = d.DistHospital
	}
	if d.ExtraOutdoor != nil {
		p.ExtraOutdoor = cleanOutdoor(*d.ExtraOutdoor)
	}
	if d.ExtraDistances != nil {
		p.ExtraDistances = cleanPlaces(*d.ExtraDistances)
	}
}

// PropertyDetailsView is the public projection of the detail fields.
type PropertyDetailsView struct {
	ShortDescription    string               `json:"short_description,omitempty"`
	AreaLabel           string               `json:"area_label,omitempty"`
	SubDistrict         string               `json:"sub_district,omitempty"`
	PostalCode          string               `json:"postal_code,omitempty"`
	Orientation         string               `json:"orientation,omitempty"`
	MaintenanceFee      *float64             `json:"maintenance_fee,omitempty"`
	CertSHM             bool                 `json:"cert_shm"`
	CertHGB             bool                 `json:"cert_hgb"`
	CertIMB             bool                 `json:"cert_imb"`
	CertPBB             bool                 `json:"cert_pbb"`
	CertGirik           bool                 `json:"cert_girik"`
	CertOther           bool                 `json:"cert_other"`
	CertSHMSRS          bool                 `json:"cert_shmsrs"`
	CertHGU             bool                 `json:"cert_hgu"`
	CertSLF             bool                 `json:"cert_slf"`
	LivingRooms         int                  `json:"living_rooms"`
	DiningRooms         int                  `json:"dining_rooms"`
	MasterBedrooms      int                  `json:"master_bedrooms"`
	ChildBedrooms       int                  `json:"child_bedrooms"`
	OtherBedrooms       int                  `json:"other_bedrooms"`
	MasterBathrooms     int                  `json:"master_bathrooms"`
	ChildBathrooms      int                  `json:"child_bathrooms"`
	OtherBathrooms      int                  `json:"other_bathrooms"`
	FamilyRooms         int                  `json:"family_rooms"`
	StudyRooms          int                  `json:"study_rooms"`
	Kitchens            int                  `json:"kitchens"`
	ElectricityCapacity string               `json:"electricity_capacity,omitempty"`
	AirConditioners     int                  `json:"air_conditioners"`
	TVProvider          string               `json:"tv_provider,omitempty"`
	WaterSource         string               `json:"water_source,omitempty"`
	FurnishingStatus    string               `json:"furnishing_status,omitempty"`
	HasPool             bool                 `json:"has_pool"`
	HasGarage           bool                 `json:"has_garage"`
	HasFrontYard        bool                 `json:"has_front_yard"`
	HasBackYard         bool                 `json:"has_back_yard"`
	HasBalcony          bool                 `json:"has_balcony"`
	DistSchool          *int                 `json:"dist_school,omitempty"`
	DistBusTerminal     *int                 `json:"dist_bus_terminal,omitempty"`
	DistMrtLrt          *int                 `json:"dist_mrt_lrt,omitempty"`
	DistToll            *int                 `json:"dist_toll,omitempty"`
	DistHospital        *int                 `json:"dist_hospital,omitempty"`
	ExtraOutdoor        []string             `json:"extra_outdoor"`
	ExtraDistances      []entity.NearbyPlace `json:"extra_distances"`
}

func toDetailsView(p *entity.Property) PropertyDetailsView {
	return PropertyDetailsView{
		ShortDescription:    p.ShortDescription,
		AreaLabel:           p.AreaLabel,
		SubDistrict:         p.SubDistrict,
		PostalCode:          p.PostalCode,
		Orientation:         p.Orientation,
		MaintenanceFee:      p.MaintenanceFee,
		CertSHM:             p.CertSHM,
		CertHGB:             p.CertHGB,
		CertIMB:             p.CertIMB,
		CertPBB:             p.CertPBB,
		CertGirik:           p.CertGirik,
		CertOther:           p.CertOther,
		LivingRooms:         p.LivingRooms,
		DiningRooms:         p.DiningRooms,
		MasterBedrooms:      p.MasterBedrooms,
		ChildBedrooms:       p.ChildBedrooms,
		OtherBedrooms:       p.OtherBedrooms,
		MasterBathrooms:     p.MasterBathrooms,
		ChildBathrooms:      p.ChildBathrooms,
		OtherBathrooms:      p.OtherBathrooms,
		FamilyRooms:         p.FamilyRooms,
		StudyRooms:          p.StudyRooms,
		Kitchens:            p.Kitchens,
		ElectricityCapacity: p.ElectricityCapacity,
		AirConditioners:     p.AirConditioners,
		TVProvider:          p.TVProvider,
		WaterSource:         p.WaterSource,
		FurnishingStatus:    p.FurnishingStatus,
		HasPool:             p.HasPool,
		HasGarage:           p.HasGarage,
		HasFrontYard:        p.HasFrontYard,
		HasBackYard:         p.HasBackYard,
		HasBalcony:          p.HasBalcony,
		DistSchool:          p.DistSchool,
		DistBusTerminal:     p.DistBusTerminal,
		DistMrtLrt:          p.DistMrtLrt,
		DistToll:            p.DistToll,
		DistHospital:        p.DistHospital,
		ExtraOutdoor:        nonNil(p.ExtraOutdoor),
		ExtraDistances:      nonNil(p.ExtraDistances),
	}
}

// cleanOutdoor trims names and drops blanks and case-insensitive duplicates.
func cleanOutdoor(in []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, n := range in {
		n = strings.TrimSpace(n)
		k := strings.ToLower(n)
		if n == "" || seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, n)
	}
	return out
}

// cleanPlaces trims names and keeps the first entry of each name (case-insensitive).
func cleanPlaces(in []entity.NearbyPlace) []entity.NearbyPlace {
	out := []entity.NearbyPlace{}
	seen := map[string]bool{}
	for _, pl := range in {
		pl.Name = strings.TrimSpace(pl.Name)
		k := strings.ToLower(pl.Name)
		if pl.Name == "" || seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, pl)
	}
	return out
}

// nonNil makes JSON render [] instead of null.
func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
