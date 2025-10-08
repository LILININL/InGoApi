package httpapi

import "net/http"

// Router ช่วยรวบรวม route ต่าง ๆ ไว้ที่เดียว
type Router struct {
	mux *http.ServeMux
}

// NewRouter สร้าง ServeMux ใหม่และเตรียมพร้อมให้ handler อื่นเพิ่มเส้นทาง
func NewRouter() *Router {
	return &Router{mux: http.NewServeMux()}
}

// RegisterAuthRoutes แม็ปเส้นทางที่เกี่ยวข้องกับ auth
func (r *Router) RegisterAuthRoutes(handler *AuthHandler) {
	r.mux.HandleFunc(AuthRegisterPath, handler.Register)
	r.mux.HandleFunc(AuthLoginPath, handler.Login)
	r.mux.HandleFunc(AuthChangePasswordPath, handler.ChangePassword)
}

// RegisterUserRoutes แม็ปเส้นทางที่เกี่ยวข้องกับข้อมูลผู้ใช้
func (r *Router) RegisterUserRoutes(handler *UserHandler) {
	r.mux.HandleFunc(UserListPath, handler.List)
}

// ServeDocs เปิดให้เข้าถึงไฟล์เอกสาร OpenAPI และหน้า Swagger UI
func (r *Router) ServeDocs(dir string) {
	fs := http.FileServer(http.Dir(dir))
	r.mux.Handle(DocsPathPrefix, http.StripPrefix(DocsPathPrefix, fs))

	// handle /docs (ไม่มีเครื่องหมาย / ปิดท้าย) ให้ redirect ไปยัง prefix ที่ถูกต้อง
	r.mux.HandleFunc("/docs", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, DocsPathPrefix, http.StatusPermanentRedirect)
	})
}

// Mux คืนค่า http.Handler เพื่อใช้กับ http.Server
func (r *Router) Mux() http.Handler {
	return corsMiddleware(r.mux)
}
