package di

//func init() {
//	obj := xdi.Object{
//		Name: "gorm",
//		New: func() (i interface{}, e error) {
//			return gorm.Open(mysql.Open(xenv.Getenv("DATABASE_DSN").String()))
//		},
//	}
//	if err := xdi.Provide(&obj); err != nil {
//		panic(err)
//	}
//}
//
//func Gorm() (db *gorm.DB) {
//	if err := xdi.Populate("gorm", &db); err != nil {
//		panic(err)
//	}
//	return
//}
