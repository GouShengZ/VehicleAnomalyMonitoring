package configs

// VehicleTypeConfig 表示不同车辆类型的配置信息
type VehicleTypeConfig struct {
	DefaultQueue string `yaml:"default_queue"` // 默认队列
	// 车辆类型（A、B、C）对应的Redis队列名称
	TypeAQueue string `yaml:"type_a_queue"` // A类型车辆队列
	TypeBQueue string `yaml:"type_b_queue"` // B类型车辆队列
	TypeCQueue string `yaml:"type_c_queue"` // C类型车辆队列

	// 使用类型对应的Redis队列名称
	ProductionCarQueue string `yaml:"production_car_queue"` // 量产车队列
	TestDriveCarQueue  string `yaml:"test_drive_car_queue"` // 试驾车队列
	MediaCarQueue      string `yaml:"media_car_queue"`      // 媒体车队列
	InternalCarQueue   string `yaml:"internal_car_queue"`   // 内部车队列

	// 车辆类型和使用类型组合的映射规则
	VehicleTypeMap map[string]string `yaml:"vehicle_type_map"` // 车辆类型映射（如：A_量产车 -> type_a_production_queue）
}

// DefaultVehicleTypeConfig 返回默认的车辆类型配置
func DefaultVehicleTypeConfig() *VehicleTypeConfig {
	return &VehicleTypeConfig{
		DefaultQueue: "default_triggers",
		// 车辆类型队列默认值
		TypeAQueue: "type_a_triggers",
		TypeBQueue: "type_b_triggers",
		TypeCQueue: "type_c_triggers",

		// 使用类型队列默认值
		ProductionCarQueue: "production_car_triggers",
		TestDriveCarQueue:  "test_drive_car_triggers",
		MediaCarQueue:      "media_car_triggers",
		InternalCarQueue:   "internal_car_triggers",

		// 车辆类型和使用类型组合的映射规则
		VehicleTypeMap: map[string]string{
			"A_量产车": "type_a_production_triggers",
			"A_试驾车": "type_a_test_drive_triggers",
			"A_媒体车": "type_a_media_triggers",
			"A_内部车": "type_a_internal_triggers",
			"B_量产车": "type_b_production_triggers",
			"B_试驾车": "type_b_test_drive_triggers",
			"B_媒体车": "type_b_media_triggers",
			"B_内部车": "type_b_internal_triggers",
			"C_量产车": "type_c_production_triggers",
			"C_试驾车": "type_c_test_drive_triggers",
			"C_媒体车": "type_c_media_triggers",
			"C_内部车": "type_c_internal_triggers",
		},
	}
}
