package models

// HouseholdDeviceRelation 表示户号和设备之间的多对多关系
type HouseholdDeviceRelation struct {
	BaseModel
	HouseholdID uint   `gorm:"not null" json:"household_id"` // 户号ID
	DeviceID    uint   `gorm:"not null" json:"device_id"`    // 设备ID
	Role        string `gorm:"type:varchar(50)" json:"role"` // 如：owner, guest, etc.

	// 关联
	Household *Household `gorm:"foreignKey:HouseholdID" json:"household,omitempty"`
	Device    *Device    `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
}
