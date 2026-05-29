package base

// var (
//   config map[string]any
// )
//
// func init() {
//   config = GetConfig("env")
// }

func Model() string {
	config := GetConfig()
	return config.Get("model", "development")
}
