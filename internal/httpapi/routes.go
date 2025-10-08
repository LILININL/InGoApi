package httpapi

// ประกาศเส้นทางทั้งหมดไว้ที่ไฟล์เดียว
// เวลาเปลี่ยน path จะได้แก้เฉพาะตรงนี้แล้วไฟล์อื่นจะตามเอง
const (
	AuthRegisterPath       = "/auth/register"
	AuthLoginPath          = "/auth/login"
	AuthChangePasswordPath = "/auth/change-password"
	UserListPath           = "/users"
	DocsPathPrefix         = "/docs/"
)
